use codespan::{FileId, Files};
use datafusion::logical_expr::LogicalPlan;
use std::sync::Arc;

mod ast;
mod diagnostics;
mod lower;
mod parser;

pub use lower::QueryContext;

pub struct PlanResult {
    pub logical_plan: Option<LogicalPlan>,
    pub diagnostics: Vec<codespan_reporting::diagnostic::Diagnostic<FileId>>,
}

pub fn plan(ctx: Arc<dyn QueryContext>, input: &str) -> PlanResult {
    let mut files: Files<String> = Files::new();
    let file_id = files.add("query", input.to_string());
    let query = parser::parse(input).unwrap();
    let mut lowerer = lower::Lowerer::new(ctx, file_id);
    let result = lowerer.lower(query);
    // not ok implies there is at least one diagnostic
    assert!(result.is_ok() || !lowerer.diagnostics().is_empty());
    PlanResult {
        logical_plan: result.ok(),
        diagnostics: lowerer.diagnostics().to_vec(),
    }
}
