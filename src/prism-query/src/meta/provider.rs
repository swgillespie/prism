use std::sync::Arc;

use async_trait::async_trait;

use crate::{
    config::MetaConfig,
    meta::client::{MetaClient, TonicMetaClient},
};
use prism_rpc_meta_v1::meta_service_client::MetaServiceClient;

#[async_trait]
pub trait MetaClientProvider: Send + Sync + 'static {
    async fn get_client(&self) -> anyhow::Result<Arc<dyn MetaClient>>;
}

pub struct DirectMetaClientProvider {
    config: MetaConfig,
}

impl DirectMetaClientProvider {
    pub fn new(config: MetaConfig) -> Self {
        Self { config }
    }
}

#[async_trait]
impl MetaClientProvider for DirectMetaClientProvider {
    async fn get_client(&self) -> anyhow::Result<Arc<dyn MetaClient>> {
        let client = MetaServiceClient::connect(self.config.endpoint.clone()).await?;
        Ok(Arc::new(TonicMetaClient::new(client)))
    }
}
