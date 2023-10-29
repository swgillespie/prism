// @generated
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTableSchemaRequest {
    #[prost(string, tag="1")]
    pub tenant_id: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub table_name: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTableSchemaResponse {
    #[prost(string, tag="1")]
    pub table_name: ::prost::alloc::string::String,
    #[prost(message, repeated, tag="2")]
    pub columns: ::prost::alloc::vec::Vec<prism_common_v1::Column>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTablePartitionsRequest {
    #[prost(string, tag="1")]
    pub tenant_id: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub table_name: ::prost::alloc::string::String,
    #[prost(message, optional, tag="3")]
    pub time_range: ::core::option::Option<prism_common_v1::TimeRange>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTablePartitionsResponse {
    #[prost(string, tag="1")]
    pub tenant_id: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub table_name: ::prost::alloc::string::String,
    #[prost(message, repeated, tag="3")]
    pub partitions: ::prost::alloc::vec::Vec<prism_common_v1::Partition>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RecordNewPartitionRequest {
    #[prost(string, tag="1")]
    pub tenant_id: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub table_name: ::prost::alloc::string::String,
    #[prost(message, optional, tag="3")]
    pub partition: ::core::option::Option<prism_common_v1::Partition>,
    #[prost(message, repeated, tag="4")]
    pub columns: ::prost::alloc::vec::Vec<prism_common_v1::Column>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RecordNewPartitionResponse {
}
include!("prism.meta.v1.tonic.rs");
// @@protoc_insertion_point(module)