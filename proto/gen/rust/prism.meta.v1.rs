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
    pub columns: ::prost::alloc::vec::Vec<TableColumn>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TableColumn {
    #[prost(string, tag="1")]
    pub name: ::prost::alloc::string::String,
    #[prost(enumeration="ColumnType", tag="2")]
    pub r#type: i32,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTablePartitionsRequest {
    #[prost(string, tag="1")]
    pub tenant_id: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub table_name: ::prost::alloc::string::String,
    #[prost(message, optional, tag="3")]
    pub time_range: ::core::option::Option<TimeRange>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TimeRange {
    #[prost(int64, tag="1")]
    pub start_time: i64,
    #[prost(int64, tag="2")]
    pub end_time: i64,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetTablePartitionsResponse {
    #[prost(string, tag="1")]
    pub tenant_id: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub table_name: ::prost::alloc::string::String,
    #[prost(message, repeated, tag="3")]
    pub partitions: ::prost::alloc::vec::Vec<Partition>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Partition {
    #[prost(string, tag="1")]
    pub name: ::prost::alloc::string::String,
    #[prost(int64, tag="2")]
    pub size: i64,
    #[prost(message, optional, tag="3")]
    pub time_range: ::core::option::Option<TimeRange>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RecordNewPartitionRequest {
    #[prost(string, tag="1")]
    pub tenant_id: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub table_name: ::prost::alloc::string::String,
    #[prost(message, optional, tag="3")]
    pub partition: ::core::option::Option<Partition>,
    #[prost(message, repeated, tag="4")]
    pub columns: ::prost::alloc::vec::Vec<TableColumn>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RecordNewPartitionResponse {
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum ColumnType {
    Unspecified = 0,
    Int64 = 1,
    Utf8 = 2,
    Timestamp = 3,
}
impl ColumnType {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            ColumnType::Unspecified => "COLUMN_TYPE_UNSPECIFIED",
            ColumnType::Int64 => "COLUMN_TYPE_INT64",
            ColumnType::Utf8 => "COLUMN_TYPE_UTF8",
            ColumnType::Timestamp => "COLUMN_TYPE_TIMESTAMP",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "COLUMN_TYPE_UNSPECIFIED" => Some(Self::Unspecified),
            "COLUMN_TYPE_INT64" => Some(Self::Int64),
            "COLUMN_TYPE_UTF8" => Some(Self::Utf8),
            "COLUMN_TYPE_TIMESTAMP" => Some(Self::Timestamp),
            _ => None,
        }
    }
}
include!("prism.meta.v1.tonic.rs");
// @@protoc_insertion_point(module)