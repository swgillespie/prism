use std::{any::Any, sync::Arc};

use async_trait::async_trait;
use datafusion::{
    arrow::datatypes::{DataType, Field, Schema},
    catalog::schema::SchemaProvider,
    datasource::TableProvider,
};

use crate::{meta::provider::MetaClientProvider, providers::table_provider::PrismTableProvider};
use prism_rpc_meta_v1::GetTableSchemaRequest;

pub struct PrismSchemaProvider {
    tenant: String,
    client_provider: Arc<dyn MetaClientProvider>,
}

impl PrismSchemaProvider {
    pub fn new(
        tenant: String,
        client_provider: Arc<dyn MetaClientProvider>,
    ) -> PrismSchemaProvider {
        PrismSchemaProvider {
            tenant,
            client_provider,
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
        /*
        let mut client = self.meta_client.lock().await;
        let resp = client
            .get_table_schema(Request::new(GetTableSchemaRequest {
                tenant_id: self.tenant.clone(),
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
        */
        let schema = {
            let client = self.client_provider.get_client().await.ok()?;
            let resp = client
                .get_table_schema(GetTableSchemaRequest {
                    tenant_id: self.tenant.clone(),
                    table_name: name.to_string(),
                })
                .await
                .ok()?;

            let mut fields = vec![];
            for field in resp.columns {
                let datatype = match field.r#type {
                    1 => DataType::Int64,
                    2 => DataType::Utf8,
                    _ => unimplemented!(),
                };

                fields.push(Field::new(field.name, datatype, true))
            }

            Arc::new(Schema::new(fields))
        };

        Some(Arc::new(PrismTableProvider::new(
            schema,
            &self.tenant,
            name,
            self.client_provider.clone(),
        )))
    }

    fn table_exist(&self, _: &str) -> bool {
        true
    }
}
