use std::{any::Any, sync::Arc};

use async_trait::async_trait;
use chrono::Utc;
use datafusion::{
    arrow::datatypes::{Schema, SchemaRef},
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
use object_store::{path::Path, ObjectMeta};

use prism_rpc_meta_v1::GetTablePartitionsRequest;

use crate::{config::S3Config, meta::provider::MetaClientProvider};

pub struct PrismTableProvider {
    schema: Arc<Schema>,
    constraints: Constraints,
    tenant: String,
    table: String,
    client_provider: Arc<dyn MetaClientProvider>,
    s3_config: S3Config,
}

impl PrismTableProvider {
    pub fn new(
        schema: Arc<Schema>,
        tenant: impl AsRef<str>,
        table: impl AsRef<str>,
        client_provider: Arc<dyn MetaClientProvider>,
        s3_config: S3Config,
    ) -> Self {
        Self {
            schema,
            constraints: Constraints::empty(),
            tenant: tenant.as_ref().to_string(),
            table: table.as_ref().to_string(),
            client_provider,
            s3_config,
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
        _state: &SessionState,
        projection: Option<&Vec<usize>>,
        filters: &[Expr],
        limit: Option<usize>,
    ) -> datafusion::common::Result<Arc<dyn ExecutionPlan>> {
        let url_path = format!("s3://{}", self.s3_config.bucket_name);
        let client = self
            .client_provider
            .get_client()
            .await
            .map_err(|e| DataFusionError::Internal(format!("{}", e)))?;
        let partitions = client
            .get_table_partitions(GetTablePartitionsRequest {
                tenant_id: self.tenant.clone(),
                table_name: self.table.clone(),
                time_range: None,
            })
            .await
            .map_err(|e| DataFusionError::Execution(e.to_string()))?
            .partitions
            .into_iter()
            .map(|p| {
                vec![PartitionedFile {
                    object_meta: ObjectMeta {
                        size: p.size as usize,
                        last_modified: Utc::now(),
                        location: Path::from(p.name),
                        e_tag: None,
                    },
                    partition_values: vec![],
                    range: None,
                    extensions: None,
                }]
            })
            .collect::<Vec<_>>();
        let config = FileScanConfig {
            object_store_url: ObjectStoreUrl::parse(url_path)?,
            file_schema: self.schema.clone(),
            file_groups: partitions,
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
            .iter()
            .fold(Expr::Literal(ScalarValue::Boolean(Some(true))), |acc, f| {
                expr_fn::and(acc, f.clone())
            });
        let predicate = create_physical_expr(&and_expr, &df_schema, self.schema.as_ref(), &props)?;
        let plan = ParquetExec::new(config, Some(predicate), None);
        Ok(Arc::new(plan))
    }
}
