use std::{io, sync::Arc};

use anyhow::Context;
use clap::Parser;
use datafusion::prelude::SessionContext;
use envconfig::Envconfig;
use ingest::Ingestor;
use object_store::{aws::AmazonS3Builder, path::Path};
use tracing::Level;
use tracing_subscriber::prelude::*;
use url::Url;

mod ingest;

#[derive(Envconfig)]
struct Config {
    #[envconfig(from = "AWS_ACCESS_KEY_ID")]
    pub aws_access_key_id: String,

    #[envconfig(from = "AWS_SECRET_ACCESS_KEY")]
    pub aws_secret_access_key: String,
}

#[derive(Parser, Debug)]
#[command(about)]
struct Args {
    #[arg(long)]
    source_bucket: String,
    #[arg(long)]
    location: String,
    #[arg(long)]
    destination_bucket: String,
    #[arg(long)]
    region: String,
    #[arg(long)]
    tenant_id: String,
    #[arg(long)]
    table: String,
}

#[tokio::main]
async fn main() {
    if let Err(e) = run().await {
        eprintln!("error: {:?}", e);
        std::process::exit(1);
    }
}

async fn run() -> anyhow::Result<()> {
    let layer = tracing_subscriber::fmt::layer()
        .with_writer(io::stderr)
        .and_then(
            tracing_subscriber::EnvFilter::from_default_env().add_directive(Level::INFO.into()),
        )
        .boxed();
    tracing_subscriber::registry().with(layer).init();

    let args = Args::parse();
    let config = Config::init_from_env()?;
    let ingestor = Ingestor::new(&args.source_bucket, &args.destination_bucket);

    let source_store = AmazonS3Builder::new()
        .with_access_key_id(&config.aws_access_key_id)
        .with_secret_access_key(&config.aws_secret_access_key)
        .with_region(&args.region)
        .with_bucket_name(&args.source_bucket)
        .build()
        .context("building S3 Source Object Store")?;
    let destination_store = AmazonS3Builder::new()
        .with_access_key_id(&config.aws_access_key_id)
        .with_secret_access_key(&config.aws_secret_access_key)
        .with_region(&args.region)
        .with_bucket_name(&args.source_bucket)
        .build()
        .context("building S3 Source Object Store")?;

    let ctx = SessionContext::new();
    let source_url = Url::parse(&format!("s3://{}", &args.source_bucket))
        .context("building source object store URL")?;
    ctx.runtime_env()
        .register_object_store(&source_url, Arc::new(source_store));
    let destination_url = Url::parse(&format!("s3://{}", &args.destination_bucket))
        .context("building source object store URL")?;
    ctx.runtime_env()
        .register_object_store(&destination_url, Arc::new(destination_store));

    let location = Path::from(args.location);
    ingestor
        .ingest_new_object(&ctx, &args.tenant_id, &args.table, &location)
        .await?;
    Ok(())
}
