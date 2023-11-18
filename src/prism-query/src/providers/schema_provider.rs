use std::{any::Any, sync::Arc};

use async_trait::async_trait;
use datafusion::{
    arrow::datatypes::{DataType, Field, Schema, TimeUnit},
    catalog::schema::SchemaProvider,
    datasource::TableProvider,
};

use crate::{
    config::S3Config, meta::provider::MetaClientProvider,
    providers::table_provider::PrismTableProvider,
};
use prism_common_v1::ColumnType;
use prism_rpc_meta_v1::GetTableSchemaRequest;

pub struct PrismSchemaProvider {
    tenant: String,
    client_provider: Arc<dyn MetaClientProvider>,
    s3_config: S3Config,
}

impl PrismSchemaProvider {
    pub fn new(
        tenant: String,
        client_provider: Arc<dyn MetaClientProvider>,
        s3_config: S3Config,
    ) -> PrismSchemaProvider {
        PrismSchemaProvider {
            tenant,
            client_provider,
            s3_config,
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
        let schema = {
            let client = match self.client_provider.get_client().await {
                Ok(client) => client,
                Err(e) => {
                    tracing::error!(err = ?e, "failed to get meta client");
                    return None;
                }
            };

            let resp = {
                match client
                    .get_table_schema(GetTableSchemaRequest {
                        tenant_id: self.tenant.clone(),
                        table_name: name.to_string(),
                    })
                    .await
                {
                    Ok(resp) => resp,
                    Err(e) => {
                        tracing::error!(err = ?e, "failed to get table schema");
                        return None;
                    }
                }
            };

            let mut fields = vec![];
            for field in resp.columns {
                let datatype = column_type_to_datafusion(field.r#type());
                fields.push(Field::new(field.name, datatype, true))
            }

            Arc::new(Schema::new(fields))
        };

        Some(Arc::new(PrismTableProvider::new(
            schema,
            &self.tenant,
            name,
            self.client_provider.clone(),
            self.s3_config.clone(),
        )))
    }

    fn table_exist(&self, _: &str) -> bool {
        true
    }
}

fn column_type_to_datafusion(ty: ColumnType) -> DataType {
    match ty {
        ColumnType::Int64 => DataType::Int64,
        ColumnType::Utf8 => DataType::Utf8,
        ColumnType::Timestamp => DataType::Timestamp(TimeUnit::Millisecond, None),
        ColumnType::Int16 => DataType::Int16,
        ColumnType::Int32 => DataType::Int32,
        ColumnType::Uint16 => DataType::UInt16,
        ColumnType::Binary => DataType::Binary,
        ColumnType::Unspecified => unimplemented!(),
    }
}
