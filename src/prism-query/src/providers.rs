use std::{any::Any, sync::Arc};

use async_trait::async_trait;
use datafusion::{
    arrow::datatypes::{DataType, Field, Schema, SchemaRef},
    catalog::{schema::SchemaProvider, CatalogProvider},
    common::{Constraints, Statistics, ToDFSchema},
    datasource::{
        listing::PartitionedFile,
        physical_plan::{FileScanConfig, ParquetExec},
        TableProvider,
    },
    error::DataFusionError,
    execution::{context::SessionState, object_store::ObjectStoreUrl},
    logical_expr::{expr_fn, TableType},
    physical_expr::{create_physical_expr, execution_props::ExecutionProps},
    physical_plan::ExecutionPlan,
    prelude::Expr,
    scalar::ScalarValue,
};
use futures::StreamExt;
use object_store::path::Path;
use tokio::sync::Mutex;
use tonic::{transport::Channel, Request};
use url::Url;

use prism_proto::{meta_service_client::MetaServiceClient, GetTableSchemaRequest};

pub struct PrismCatalogProvider {
    meta_client: Arc<Mutex<MetaServiceClient<Channel>>>,
}

impl PrismCatalogProvider {
    pub fn new(meta_client: Arc<Mutex<MetaServiceClient<Channel>>>) -> PrismCatalogProvider {
        PrismCatalogProvider { meta_client }
    }
}

impl CatalogProvider for PrismCatalogProvider {
    fn as_any(&self) -> &dyn Any {
        self
    }

    fn schema_names(&self) -> Vec<String> {
        todo!()
    }

    fn schema(&self, name: &str) -> Option<Arc<dyn SchemaProvider>> {
        name;
        Some(Arc::new(PrismSchemaProvider::new(
            name.to_string(),
            self.meta_client.clone(),
        )))
    }
}

pub struct PrismSchemaProvider {
    tenant: String,
    meta_client: Arc<Mutex<MetaServiceClient<Channel>>>,
}

impl PrismSchemaProvider {
    pub fn new(
        tenant: String,
        meta_client: Arc<Mutex<MetaServiceClient<Channel>>>,
    ) -> PrismSchemaProvider {
        PrismSchemaProvider {
            tenant,
            meta_client,
        }
    }
}

#[async_trait]
impl SchemaProvider for PrismSchemaProvider {
    fn as_any(&self) -> &dyn Any {
        self
    }

    fn table_names(&self) -> Vec<String> {
        todo!()
    }

    async fn table(&self, name: &str) -> Option<Arc<dyn TableProvider>> {
        let mut client = self.meta_client.lock().await;
        let resp = client
            .get_table_schema(Request::new(GetTableSchemaRequest {
                table_name: name.to_string(),
            }))
            .await
            .ok()?;
        let schema = resp.into_inner();
        let mut fields = vec![];
        for field in schema.columns {
            let datatype = match field.r#type {
                1 => DataType::Int64,
                2 => DataType::Utf8,
                _ => unimplemented!(),
            };

            fields.push(Field::new(field.name, datatype, true))
        }

        let schema = Arc::new(Schema::new(fields));
        Some(Arc::new(PrismTableProvider::new(schema, name)))
    }

    fn table_exist(&self, _: &str) -> bool {
        true
    }
}

pub struct PrismTableProvider {
    schema: Arc<Schema>,
    constraints: Constraints,
    table: String,
}

impl PrismTableProvider {
    pub fn new(schema: Arc<Schema>, table: impl AsRef<str>) -> Self {
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
