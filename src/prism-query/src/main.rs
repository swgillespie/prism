use std::{
    io::{self, BufRead, Write},
    sync::Arc,
};

use clap::Parser;
use datafusion::prelude::SessionContext;
use envconfig::Envconfig;
use object_store::aws::AmazonS3Builder;
use prism_proto::meta_service_client::MetaServiceClient;
use providers::PrismCatalogProvider;
use tokio::sync::Mutex;
use url::Url;

mod providers;

#[derive(Envconfig)]
struct Config {
    #[envconfig(from = "AWS_ACCESS_KEY_ID")]
    pub aws_access_key_id: String,

    #[envconfig(from = "AWS_SECRET_ACCESS_KEY")]
    pub aws_secret_access_key: String,

    #[envconfig(from = "QUERY_BUCKET_NAME")]
    pub query_bucket_name: String,
}

#[derive(Parser, Debug)]
#[command(about)]
struct Args {
    #[arg(long, default_value = "http://localhost:8080")]
    meta_url: String,
}

#[tokio::main]
async fn main() {
    if let Err(e) = repl().await {
        eprintln!("error: {:?}", e);
        std::process::exit(1);
    }
}

async fn repl() -> anyhow::Result<()> {
    let args = Args::parse();
    let config = Config::init_from_env()?;
    let store = AmazonS3Builder::new()
        .with_access_key_id(config.aws_access_key_id)
        .with_secret_access_key(config.aws_secret_access_key)
        .with_region("us-west-2")
        .with_bucket_name(config.query_bucket_name.clone())
        .build()?;
    let meta = Arc::new(Mutex::new(MetaServiceClient::connect(args.meta_url).await?));
    let s3_url = Url::parse(&format!("s3://{}", config.query_bucket_name))?;
    let catalog = PrismCatalogProvider::new(meta);
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
    }

    Ok(())
}
