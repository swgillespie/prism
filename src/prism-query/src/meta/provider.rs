use std::{sync::Arc, time::Duration};

use async_trait::async_trait;
use tonic::transport::Endpoint;

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
        let channel = Endpoint::from_shared(self.config.endpoint.clone())?
            .connect_timeout(Duration::from_secs(self.config.connect_timeout_seconds))
            .timeout(Duration::from_secs(self.config.timeout_seconds))
            .connect()
            .await?;

        let client = MetaServiceClient::new(channel);
        Ok(Arc::new(TonicMetaClient::new(client)))
    }
}
