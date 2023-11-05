use std::{any::Any, sync::Arc};

use datafusion::catalog::{schema::SchemaProvider, CatalogProvider};

use crate::{meta::provider::MetaClientProvider, providers::schema_provider::PrismSchemaProvider};

pub struct PrismCatalogProvider {
    client_provider: Arc<dyn MetaClientProvider>,
}

impl PrismCatalogProvider {
    pub fn new(client_provider: Arc<dyn MetaClientProvider>) -> PrismCatalogProvider {
        PrismCatalogProvider { client_provider }
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
        Some(Arc::new(PrismSchemaProvider::new(
            name.to_string(),
            self.client_provider.clone(),
        )))
    }
}
