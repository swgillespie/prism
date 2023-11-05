use std::{
    io::{self, BufRead, Write},
    sync::Arc,
    time::Instant,
};

use anyhow::Context;
use datafusion::prelude::SessionContext;
use envconfig::Envconfig;
use meta::provider::DirectMetaClientProvider;
use object_store::aws::AmazonS3Builder;
use url::Url;

use crate::providers::catalog_provider::PrismCatalogProvider;

mod config;
mod meta;
mod providers;

#[derive(Envconfig)]
struct Config {
    #[envconfig(from = "AWS_ACCESS_KEY_ID")]
    pub aws_access_key_id: String,

    #[envconfig(from = "AWS_SECRET_ACCESS_KEY")]
    pub aws_secret_access_key: String,

    #[envconfig(from = "QUERY_BUCKET_NAME")]
    pub query_bucket_name: String,

    #[envconfig(from = "PRISM_QUERY_CONFIG")]
    pub config_path: String,
}

#[tokio::main]
async fn main() {
    if let Err(e) = repl().await {
        eprintln!("error: {:?}", e);
        std::process::exit(1);
    }
}

async fn repl() -> anyhow::Result<()> {
    let env_config = Config::init_from_env()?;
    let config = config::get_config(&env_config.config_path).context("reading config from file")?;
    let store = AmazonS3Builder::new()
        .with_access_key_id(env_config.aws_access_key_id)
        .with_secret_access_key(env_config.aws_secret_access_key)
        .with_region("us-west-2")
        .with_bucket_name(env_config.query_bucket_name.clone())
        .build()?;
    let client_provider = Arc::new(DirectMetaClientProvider::new(config.meta.clone()));
    let s3_url = Url::parse(&format!("s3://{}", &config.s3.bucket_name))?;
    let catalog = PrismCatalogProvider::new(client_provider);
    let ctx = SessionContext::new();
    ctx.register_catalog("prism", Arc::new(catalog));
    ctx.runtime_env()
        .register_object_store(&s3_url, Arc::new(store));
    let mut stdin = io::stdin().lock();
    let mut stdout = io::stdout().lock();
    loop {
        write!(&mut stdout, "prism> ")?;
        stdout.flush()?;
        let mut buf = String::new();
        stdin.read_line(&mut buf)?;
        let buf = buf.trim();
        if buf == "quit" || buf == "" {
            break;
        }

        let start = Instant::now();
        let df = match ctx.sql(&buf).await {
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
