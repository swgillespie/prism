use std::{
    any::Any,
    io::{self, BufRead, Write},
    sync::Arc,
};

use async_trait::async_trait;
use datafusion::{
    arrow::datatypes::{DataType, Field, Schema, SchemaRef},
    common::{Constraints, ToDFSchema},
    datasource::{
        listing::PartitionedFile,
        physical_plan::{FileScanConfig, ParquetExec},
        TableProvider,
    },
    error::DataFusionError,
    execution::{context::SessionState, object_store::ObjectStoreUrl},
    logical_expr::{expr_fn, TableType},
    physical_expr::{create_physical_expr, execution_props::ExecutionProps},
    physical_plan::{ExecutionPlan, Statistics},
    prelude::{Expr, SessionContext},
    scalar::ScalarValue,
};
use envconfig::Envconfig;
use futures::StreamExt;
use object_store::{aws::AmazonS3Builder, path::Path};
use url::Url;

pub struct PrismTableProvider {
    schema: Arc<Schema>,
    constraints: Constraints,
    table: String,
}

impl PrismTableProvider {
    pub fn new(table: impl AsRef<str>) -> Self {
        let fields = vec![
            Field::new("bytes", DataType::Int64, true),
            Field::new("datetime", DataType::Utf8, true),
            Field::new("host", DataType::Utf8, true),
            Field::new("method", DataType::Utf8, true),
            Field::new("protocol", DataType::Utf8, true),
            Field::new("referer", DataType::Utf8, true),
            Field::new("status", DataType::Utf8, true),
            Field::new("user-identifier", DataType::Utf8, true),
            Field::new("timestamp", DataType::Int64, true),
        ];
        let schema = Arc::new(Schema::new(fields));
        Self {
            schema,
            constraints: Constraints::empty(),
            table: table.as_ref().to_string(),
        }
    }
}

#[async_trait]
impl TableProvider for PrismTableProvider {
    fn as_any(&self) -> &dyn Any {
        self
    }

    fn schema(&self) -> SchemaRef {
        self.schema.clone()
    }

    fn constraints(&self) -> Option<&Constraints> {
        Some(&self.constraints)
    }

    fn table_type(&self) -> TableType {
        TableType::Base
    }

    async fn scan(
        &self,
        state: &SessionState,
        projection: Option<&Vec<usize>>,
        filters: &[Expr],
        limit: Option<usize>,
    ) -> datafusion::common::Result<Arc<dyn ExecutionPlan>> {
        let url_path = format!("s3://prism-storage-b7c0d9c");
        let table_url =
            Url::parse(&url_path).map_err(|e| DataFusionError::External(Box::new(e)))?;
        let obj_store = state
            .runtime_env()
            .object_store_registry
            .get_store(&table_url)?;
        let path = Path::from(self.table.clone());
        let partitions = obj_store
            .list(Some(&path))
            .await?
            .collect::<Vec<_>>()
            .await
            .into_iter()
            .collect::<Result<Vec<_>, _>>()?
            .into_iter()
            .map(|meta| PartitionedFile {
                object_meta: meta,
                partition_values: vec![],
                range: None,
                extensions: None,
            })
            .collect::<Vec<_>>();

        let config = FileScanConfig {
            object_store_url: ObjectStoreUrl::parse(url_path)?,
            file_schema: self.schema.clone(),
            file_groups: vec![partitions],
            statistics: Statistics::default(),
            projection: projection.cloned(),
            limit,
            table_partition_cols: vec![],
            output_ordering: vec![],
            infinite_source: false,
        };

        let df_schema = self.schema.clone().to_dfschema()?;
        let props = ExecutionProps::new();
        let and_expr = filters
            .into_iter()
            .fold(Expr::Literal(ScalarValue::Boolean(Some(true))), |acc, f| {
                expr_fn::and(acc, f.clone())
            });
        let predicate = create_physical_expr(&and_expr, &df_schema, self.schema.as_ref(), &props)?;
        let plan = ParquetExec::new(config, Some(predicate), None);
        Ok(Arc::new(plan))
    }
}

#[derive(Envconfig)]
struct Config {
    #[envconfig(from = "AWS_ACCESS_KEY_ID")]
    pub aws_access_key_id: String,

    #[envconfig(from = "AWS_SECRET_ACCESS_KEY")]
    pub aws_secret_access_key: String,
}

#[tokio::main]
async fn main() {
    if let Err(e) = repl().await {
        eprintln!("error: {:?}", e);
        std::process::exit(1);
    }
}

async fn repl() -> anyhow::Result<()> {
    let config = Config::init_from_env()?;
    let store = AmazonS3Builder::new()
        .with_access_key_id(config.aws_access_key_id)
        .with_secret_access_key(config.aws_secret_access_key)
        .with_region("us-west-2")
        .with_bucket_name("prism-storage-b7c0d9c")
        .build()?;

    let ctx = SessionContext::new();
    let s3_url = Url::parse("s3://prism-storage-b7c0d9c")?;
    ctx.runtime_env()
        .register_object_store(&s3_url, Arc::new(store));
    let provider = PrismTableProvider::new("demo");
    ctx.register_table("demo", Arc::new(provider))?;
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
                writeln!(&mut stdout, "error: {}", e)?;
                continue;
            }
        };

        df.show().await?;
    }

    Ok(())
}
