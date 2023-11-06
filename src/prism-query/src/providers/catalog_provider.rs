use std::{any::Any, sync::Arc};

use datafusion::catalog::{schema::SchemaProvider, CatalogProvider};

use crate::{
    config::S3Config, meta::provider::MetaClientProvider,
    providers::schema_provider::PrismSchemaProvider,
};

pub struct PrismCatalogProvider {
    client_provider: Arc<dyn MetaClientProvider>,
    s3_config: S3Config,
}

impl PrismCatalogProvider {
    pub fn new(
        client_provider: Arc<dyn MetaClientProvider>,
        s3_config: S3Config,
    ) -> PrismCatalogProvider {
        PrismCatalogProvider {
            client_provider,
            s3_config,
        }
    }
}

impl CatalogProvider for PrismCatalogProvider {
    fn as_any(&self) -> &dyn Any {
        self
    }

    fn schema_names(&self) -> Vec<String> {
        unimplemented!("prism does not support listing schemas")
    }

    fn schema(&self, name: &str) -> Option<Arc<dyn SchemaProvider>> {
        Some(Arc::new(PrismSchemaProvider::new(
            name.to_string(),
            self.client_provider.clone(),
            self.s3_config.clone(),
        )))
    }
}
