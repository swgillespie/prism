use std::sync::Arc;

use codespan::{FileId, Files};
use codespan_reporting::diagnostic::Diagnostic;
use datafusion::logical_expr::{LogicalPlan, TableSource};
use either::Either;

pub mod ast;
mod diagnostics;
mod lower;
mod parser;

pub use lower::QueryContext;

use ast::Query;

pub fn parse(input: &str) -> anyhow::Result<Query> {
    let query = parser::parse(input).map_err(|e| anyhow::anyhow!("parse error: {}", e))?;
    Ok(query)
}

pub fn lower(
    query: Query,
    ctx: Arc<dyn QueryContext>,
    table_source: Arc<dyn TableSource>,
    input: &str,
) -> either::Either<LogicalPlan, Vec<Diagnostic<FileId>>> {
    let mut files: Files<String> = Files::new();
    let fileid = files.add("query", input.to_string());
    let mut lowerer = lower::Lowerer::new(ctx, table_source, fileid, query.table.name.clone());
    match lowerer.lower(query) {
        Ok(plan) => Either::Left(plan),
        Err(_) => Either::Right(lowerer.diagnostics().to_vec()),
    }
}
