use async_trait::async_trait;
use tokio::sync::Mutex;
use tonic::transport::Channel;

use prism_rpc_meta_v1::{
    meta_service_client::MetaServiceClient, GetTablePartitionsRequest, GetTablePartitionsResponse,
    GetTableSchemaRequest, GetTableSchemaResponse,
};

#[async_trait]
pub trait MetaClient: Send + Sync + 'static {
    async fn get_table_schema(
        &self,
        request: GetTableSchemaRequest,
    ) -> anyhow::Result<GetTableSchemaResponse>;

    async fn get_table_partitions(
        &self,
        request: GetTablePartitionsRequest,
    ) -> anyhow::Result<GetTablePartitionsResponse>;
}

pub(in crate::meta) struct TonicMetaClient {
    client: Mutex<MetaServiceClient<Channel>>,
}

impl TonicMetaClient {
    pub fn new(client: MetaServiceClient<Channel>) -> Self {
        Self {
            client: Mutex::new(client),
        }
    }
}

#[async_trait]
impl MetaClient for TonicMetaClient {
    async fn get_table_schema(
        &self,
        request: GetTableSchemaRequest,
    ) -> anyhow::Result<GetTableSchemaResponse> {
        let mut client = self.client.lock().await;
        let resp = client.get_table_schema(request).await?;
        Ok(resp.into_inner())
    }

    async fn get_table_partitions(
        &self,
        request: GetTablePartitionsRequest,
    ) -> anyhow::Result<GetTablePartitionsResponse> {
        let mut client = self.client.lock().await;
        let resp = client.get_table_partitions(request).await?;
        Ok(resp.into_inner())
    }
}
