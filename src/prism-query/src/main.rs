use std::{
    io::{self, BufRead, Write},
    sync::Arc,
    time::Instant,
};

use anyhow::Context;
use clap::Parser;
use datafusion::prelude::SessionContext;
use envconfig::Envconfig;
use meta::provider::DirectMetaClientProvider;
use object_store::aws::AmazonS3Builder;
use tracing::Level;
use tracing_subscriber::prelude::*;
use url::Url;

use crate::providers::catalog_provider::PrismCatalogProvider;

mod config;
mod meta;
mod providers;

#[derive(Debug, Parser)]
struct Args {
    /// If present, run the given SQL query and exit.
    #[arg(short, long)]
    sql: Option<String>,
}

#[derive(Envconfig)]
struct Config {
    #[envconfig(from = "AWS_ACCESS_KEY_ID")]
    pub aws_access_key_id: String,

    #[envconfig(from = "AWS_SECRET_ACCESS_KEY")]
    pub aws_secret_access_key: String,

    #[envconfig(from = "AWS_REGION")]
    pub aws_region: String,

    #[envconfig(from = "PRISM_QUERY_CONFIG")]
    pub config_path: String,
}

#[tokio::main(flavor = "multi_thread")]
async fn main() {
    let layer = tracing_subscriber::fmt::layer()
        .with_writer(io::stderr)
        .and_then(
            tracing_subscriber::EnvFilter::from_default_env().add_directive(Level::INFO.into()),
        )
        .boxed();
    tracing_subscriber::registry().with(layer).init();

    if let Err(e) = repl().await {
        eprintln!("error: {:?}", e);
        std::process::exit(1);
    }
}

async fn repl() -> anyhow::Result<()> {
    let args = Args::parse();
    let env_config = Config::init_from_env()?;
    let config = config::get_config(&env_config.config_path).context("reading config from file")?;
    let store = {
        let mut builder = AmazonS3Builder::new()
            .with_access_key_id(env_config.aws_access_key_id)
            .with_secret_access_key(env_config.aws_secret_access_key)
            .with_region(env_config.aws_region)
            .with_bucket_name(config.s3.bucket_name.clone());
        builder = if let Some(endpoint) = &config.s3.endpoint {
            builder.with_endpoint(endpoint).with_allow_http(true)
        } else {
            builder
        };

        builder.build()?
    };
    let client_provider = Arc::new(DirectMetaClientProvider::new(config.meta.clone()));
    let s3_url = Url::parse(&format!("s3://{}", &config.s3.bucket_name))?;
    let catalog = PrismCatalogProvider::new(client_provider, config.s3.clone());
    let ctx = SessionContext::new();
    ctx.register_catalog("prism", Arc::new(catalog));
    ctx.runtime_env()
        .register_object_store(&s3_url, Arc::new(store));
    let mut stdin = io::stdin().lock();
    let mut stdout = io::stdout().lock();
    if let Some(sql) = args.sql {
        let df = ctx.sql(&sql).await?;
        df.show().await?;
        return Ok(());
    }

    loop {
        write!(&mut stdout, "prism> ")?;
        stdout.flush()?;
        let mut buf = String::new();
        stdin.read_line(&mut buf)?;
        let buf = buf.trim();
        if buf == "quit" || buf.is_empty() {
            break;
        }

        let start = Instant::now();
        let df = match ctx.sql(buf).await {
            Ok(df) => df,
            Err(e) => {
                writeln!(&mut stdout, "sql error: {}", e)?;
                continue;
            }
        };

        if let Err(e) = df.show().await {
            writeln!(&mut stdout, "show error: {}", e)?;
            continue;
        }

        let end = Instant::now();
        writeln!(&mut stdout, "query took {:?}ms", (end - start).as_millis())?;
    }

    Ok(())
}
