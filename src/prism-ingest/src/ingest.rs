use anyhow::Context;
use datafusion::{
    arrow::array::TimestampMillisecondArray,
    dataframe::DataFrameWriteOptions,
    datasource::{
        file_format::{
            file_compression_type::FileCompressionType, DEFAULT_SCHEMA_INFER_MAX_RECORD,
        },
        listing::ListingTableInsertMode,
    },
    logical_expr::expr_fn,
    parquet::{
        basic::Compression,
        basic::Encoding,
        file::properties::{WriterProperties, WriterVersion},
    },
    prelude::{DataFrame, Expr, NdJsonReadOptions, SessionContext},
};
use object_store::path::Path;
use tracing::{info_span, Instrument};
use url::Url;

pub struct Ingestor {
    ingest_bucket_name: String,
    query_bucket_name: String,
}

pub struct Partition {
    pub name: String,
    pub size: usize,
    pub max_ts: i64,
    pub min_ts: i64,
}

impl Ingestor {
    pub fn new(ingest_bucket_name: impl AsRef<str>, query_bucket_name: impl AsRef<str>) -> Self {
        Self {
            ingest_bucket_name: ingest_bucket_name.as_ref().to_string(),
            query_bucket_name: query_bucket_name.as_ref().to_string(),
        }
    }

    #[tracing::instrument(skip(self, ctx))]
    pub async fn ingest_new_object(
        &self,
        ctx: &SessionContext,
        tenant_id: &str,
        table: &str,
        location: &Path,
    ) -> anyhow::Result<Partition> {
        let path = format!("s3://{}/{}", self.ingest_bucket_name, location);
        let read_span = info_span!("read_json", path = %path);
        let mut df = ctx
            .read_json(
                path,
                NdJsonReadOptions {
                    schema: None,
                    schema_infer_max_records: DEFAULT_SCHEMA_INFER_MAX_RECORD,
                    file_extension: "json",
                    table_partition_cols: vec![],
                    file_compression_type: FileCompressionType::UNCOMPRESSED,
                    infinite: false,
                    file_sort_order: vec![],
                    insert_mode: ListingTableInsertMode::Error,
                },
            )
            .instrument(read_span)
            .await
            .context(format!("reading raw data {:?} from S3", location))?;
        df = normalize_timestamp(df)?;

        let start_end = df
            .clone()
            .aggregate(
                vec![],
                vec![
                    expr_fn::max(expr_fn::col("timestamp")),
                    expr_fn::min(expr_fn::col("timestamp")),
                ],
            )?
            .collect()
            .await?;
        let (max_ts, min_ts) = if start_end.len() != 1 {
            anyhow::bail!(
                "expected 1 row in timestamp aggregation, got {}",
                start_end.len()
            )
        } else {
            let row = &start_end[0];
            let max_col = row
                .column(0)
                .as_any()
                .downcast_ref::<TimestampMillisecondArray>()
                .ok_or_else(|| anyhow::anyhow!("ono"))?;
            let min_col = row
                .column(1)
                .as_any()
                .downcast_ref::<TimestampMillisecondArray>()
                .ok_or_else(|| anyhow::anyhow!("ono"))?;
            (max_col.value(0), min_col.value(0))
        };

        let props = WriterProperties::builder()
            .set_writer_version(WriterVersion::PARQUET_2_0)
            .set_encoding(Encoding::PLAIN)
            .set_compression(Compression::SNAPPY)
            .build();

        let output_file = location
            .filename()
            .ok_or_else(|| anyhow::anyhow!("getting filename from path"))
            .context("getting filename from path")?;
        let output_path = format!("{}/{}/{}.parquet", tenant_id, table, output_file);
        let s3_output_path = format!("s3://{}/{}", self.query_bucket_name, output_path);
        let write_span = info_span!("write_parquet", path = %s3_output_path);
        df.clone().limit(0, Some(5))?.show().await?;
        df.write_parquet(
            &s3_output_path,
            DataFrameWriteOptions::new().with_single_file_output(true),
            Some(props),
        )
        .instrument(write_span)
        .await
        .context("writing Parquet to s3")?;

        let url = Url::parse(&s3_output_path)?;
        let object = ctx
            .runtime_env()
            .object_store_registry
            .get_store(&url)?
            .head(&Path::from(output_path.as_ref()))
            .await?;

        Ok(Partition {
            name: output_path,
            size: object.size,
            max_ts,
            min_ts,
        })
    }
}

// normalize_timestamp casts the parses `timestamp` column into a Timestamp type for persistence and also for the
// computation of max and min timestamp for this partition.
fn normalize_timestamp(df: DataFrame) -> anyhow::Result<DataFrame> {
    let schema = df.schema();
    let mut cols: Vec<Expr> = schema
        .field_names()
        .iter()
        .map(|name| name.strip_prefix("?table?.").unwrap_or_default())
        .filter(|&name| name != "timestamp")
        .map(|name| expr_fn::col(name).alias(name))
        .collect();

    cols.push(expr_fn::to_timestamp_millis(expr_fn::col("timestamp")).alias("timestamp"));
    Ok(df.select(cols)?)
}

#[cfg(test)]
mod tests {
    use std::sync::Arc;

    use bytes::Bytes;
    use datafusion::prelude::SessionContext;
    use object_store::{memory::InMemory, path::Path, ObjectStore};
    use url::Url;

    use super::Ingestor;

    const INGEST_BUCKET_NAME: &str = "ingest";
    const QUERY_BUCKET_NAME: &str = "query";
    const TENANT_ID: &str = "tenant";
    const TABLE: &str = "web_requests";

    #[tokio::test]
    async fn basic_ingestion() {
        let ctx = SessionContext::new();
        let ingestor_url =
            Url::parse(&format!("s3://{}", INGEST_BUCKET_NAME)).expect("invalid bucket url");
        let query_url =
            Url::parse(&format!("s3://{}", QUERY_BUCKET_NAME)).expect("invalid bucket url");
        let ingestor_memstore = Arc::new(InMemory::new());
        let query_memstore = Arc::new(InMemory::new());
        ctx.runtime_env()
            .register_object_store(&ingestor_url, ingestor_memstore.clone());
        ctx.runtime_env()
            .register_object_store(&query_url, query_memstore.clone());
        let ingest_path = Path::from("demo.json");
        let ingest_data = Bytes::from_static(include_bytes!("./testdata/demo.json"));
        ingestor_memstore
            .put(&ingest_path, ingest_data)
            .await
            .unwrap();
        let target_path = format!("{}/{}/demo.json.parquet", TENANT_ID, TABLE);
        let ingestor = Ingestor::new(INGEST_BUCKET_NAME, QUERY_BUCKET_NAME);
        let partition = ingestor
            .ingest_new_object(&ctx, TENANT_ID, TABLE, &ingest_path)
            .await
            .unwrap();
        assert_eq!(partition.name, target_path);
        assert!(partition.size > 0);
        assert_eq!(partition.max_ts, 1698000995523);
        assert_eq!(partition.min_ts, 1698000992225);
        query_memstore
            .head(&Path::from(target_path))
            .await
            .expect("object should be present in query memstore");
    }
}
