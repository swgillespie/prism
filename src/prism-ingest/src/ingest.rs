use anyhow::Context;
use datafusion::{
    dataframe::DataFrameWriteOptions,
    datasource::{
        file_format::{
            file_compression_type::FileCompressionType, DEFAULT_SCHEMA_INFER_MAX_RECORD,
        },
        listing::ListingTableInsertMode,
    },
    parquet::{
        basic::Compression,
        basic::Encoding,
        file::properties::{WriterProperties, WriterVersion},
    },
    prelude::{NdJsonReadOptions, SessionContext},
};
use object_store::ObjectMeta;
use tracing::{info_span, Instrument};

pub struct Ingestor {
    ingest_bucket_name: String,
    query_bucket_name: String,
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
        object: ObjectMeta,
    ) -> anyhow::Result<()> {
        let path = format!("s3://{}/{}", self.ingest_bucket_name, object.location);
        let read_span = info_span!("read_json", path = %path);
        let df = ctx
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
            .context(format!("reading raw data {:?} from S3", object.location))?;

        let props = WriterProperties::builder()
            .set_writer_version(WriterVersion::PARQUET_2_0)
            .set_encoding(Encoding::PLAIN)
            .set_compression(Compression::SNAPPY)
            .build();

        let output_folder = object
            .location
            .filename()
            .ok_or_else(|| anyhow::anyhow!("getting filename from path"))
            .context("getting filename from path")?;
        let output_path = format!(
            "s3://{}/{}/{}",
            self.query_bucket_name, "web_requests", output_folder,
        );
        let write_span = info_span!("write_parquet", path = %output_path);
        df.write_parquet(&output_path, DataFrameWriteOptions::new(), Some(props))
            .instrument(write_span)
            .await
            .context("writing Parquet to s3")?;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use std::sync::Arc;

    use bytes::Bytes;
    use datafusion::prelude::SessionContext;
    use futures::StreamExt;
    use object_store::{memory::InMemory, path::Path, ObjectStore};
    use url::Url;

    use super::Ingestor;

    const INGEST_BUCKET_NAME: &str = "ingest";
    const QUERY_BUCKET_NAME: &str = "query";

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
        let objmeta = ingestor_memstore.head(&ingest_path).await.unwrap();
        let ingestor = Ingestor::new(INGEST_BUCKET_NAME, QUERY_BUCKET_NAME);

        ingestor.ingest_new_object(&ctx, objmeta).await.unwrap();

        let objects = query_memstore
            .list(Some(&Path::from("web_requests/demo.json")))
            .await
            .unwrap()
            .collect::<Vec<_>>()
            .await
            .into_iter()
            .collect::<Result<Vec<_>, _>>()
            .unwrap();
        println!("{:?}", objects);
        assert_eq!(1, objects.len());
    }
}
