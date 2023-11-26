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

    #[envconfig(from = "AWS_REGION")]
    pub aws_region: String,
}

#[derive(Parser, Debug)]
#[command(about)]
struct Args {
    #[arg(long)]
    source: String,
    #[arg(long)]
    location: String,
    #[arg(long)]
    destination: String,
    #[arg(long)]
    tenant_id: String,
    #[arg(long)]
    table: String,
    #[arg(long)]
    s3_endpoint: Option<String>,
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
    let ctx = SessionContext::new();
    let ingestor = initialize_session(&args, &ctx)?;
    let location = Path::from(args.location);
    let partition = ingestor
        .ingest_new_object(&ctx, &args.tenant_id, &args.table, &location)
        .await?;

    let output = serde_json::to_string_pretty(&partition)?;
    println!("{}", output);
    Ok(())
}

fn initialize_session(args: &Args, ctx: &SessionContext) -> anyhow::Result<Ingestor> {
    tracing::info!(source = %args.source, destination = %args.destination, "using s3 storage");
    let config = Config::init_from_env()?;
    let source_store = {
        let mut builder = AmazonS3Builder::new()
            .with_access_key_id(&config.aws_access_key_id)
            .with_secret_access_key(&config.aws_secret_access_key)
            .with_region(&config.aws_region)
            .with_bucket_name(&args.source);
        builder = if let Some(endpoint) = &args.s3_endpoint {
            tracing::info!(source = %endpoint, "using custom s3 endpoint for source");
            builder.with_endpoint(endpoint).with_allow_http(true)
        } else {
            tracing::info!("using default s3 endpoint for source");
            builder
        };

        builder.build().context("building S3 source object store")?
    };

    let destination_store = {
        let mut builder = AmazonS3Builder::new()
            .with_access_key_id(&config.aws_access_key_id)
            .with_secret_access_key(&config.aws_secret_access_key)
            .with_region(&config.aws_region)
            .with_bucket_name(&args.destination);
        builder = if let Some(endpoint) = &args.s3_endpoint {
            tracing::info!(source = %endpoint, "using custom s3 endpoint for destination");
            builder.with_endpoint(endpoint).with_allow_http(true)
        } else {
            tracing::info!("using default s3 endpoint for destination");
            builder
        };

        builder
            .build()
            .context("building S3 destination object store")?
    };

    let source_url = Url::parse(&format!("s3://{}", &args.source))
        .context("building source object store URL")?;
    ctx.runtime_env()
        .register_object_store(&source_url, Arc::new(source_store));
    let destination_url = Url::parse(&format!("s3://{}", &args.destination))
        .context("building source object store URL")?;
    ctx.runtime_env()
        .register_object_store(&destination_url, Arc::new(destination_store));
    let ingestor = Ingestor::new(&args.source, &args.destination);
    Ok(ingestor)
}
