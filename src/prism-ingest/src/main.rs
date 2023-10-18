use std::sync::Arc;

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
        basic::{Compression, Encoding},
        file::properties::{WriterProperties, WriterVersion},
    },
    prelude::{NdJsonReadOptions, SessionContext},
};
use envconfig::Envconfig;
use object_store::aws::AmazonS3Builder;
use url::Url;

#[derive(Envconfig)]
struct Config {
    #[envconfig(from = "AWS_ACCESS_KEY_ID")]
    pub aws_access_key_id: String,

    #[envconfig(from = "AWS_SECRET_ACCESS_KEY")]
    pub aws_secret_access_key: String,

    #[envconfig(from = "INGESTION_BUCKET_NAME")]
    pub bucket_name: String,
}

#[tokio::main]
async fn main() {
    if let Err(e) = run().await {
        eprintln!("error: {:?}", e);
        std::process::exit(1);
    }
}

async fn run() -> anyhow::Result<()> {
    let config = Config::init_from_env()?;
    let store = AmazonS3Builder::new()
        .with_access_key_id(config.aws_access_key_id)
        .with_secret_access_key(config.aws_secret_access_key)
        .with_region("us-west-2")
        .with_bucket_name(config.bucket_name.clone())
        .build()
        .context("building S3 Object Store")?;

    let ctx = SessionContext::new();
    let url =
        Url::parse(&format!("s3://{}", config.bucket_name)).context("building object store URL")?;
    ctx.runtime_env()
        .register_object_store(&url, Arc::new(store));

    let df = ctx
        .read_json(
            format!("s3://{}/demo.json", config.bucket_name),
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
        .await
        .context("reading JSON from s3")?;

    let props = WriterProperties::builder()
        .set_writer_version(WriterVersion::PARQUET_2_0)
        .set_encoding(Encoding::PLAIN)
        .set_compression(Compression::SNAPPY)
        .build();
    df.write_parquet("./.scratch/out", DataFrameWriteOptions::new(), Some(props))
        .await
        .context("writing Parquet to out")?;
    Ok(())
}
