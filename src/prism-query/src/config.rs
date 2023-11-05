//! Configuration for prism-query.
//!
//! The querier component needs to be aware of configuration for the meta service that it talks to and the S3 buckets
//! that contain the data that the querier intends to reference.

use config::{Config, File};
use serde::{Deserialize, Serialize};

/// Configuration pertaining to the meta service.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MetaConfig {
    /// The endpoint that the meta service is listening on.
    pub endpoint: String,
}

/// Configuration pertaining to S3.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct S3Config {
    /// An optional S3 endpoint to use. Can be overridden when testing with LocalStack, otherwise defaults to whatever
    /// AWS s3 endpoint is appropriate for the given region.
    pub endpoint: Option<String>,

    /// The name of the bucket containing the data to be queried.
    pub bucket_name: String,
}

/// Configuration for the querier.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct QueryConfig {
    /// Meta configuration.
    pub meta: MetaConfig,
    /// S3 configuration.
    pub s3: S3Config,
}

/// Reads configuration from the given path.
pub fn get_config(config_path: &str) -> anyhow::Result<QueryConfig> {
    let settings = Config::builder()
        .add_source(File::with_name(config_path))
        .build()?;
    let query_config = settings.try_deserialize()?;
    Ok(query_config)
}
