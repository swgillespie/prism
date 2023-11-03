use std::sync::Arc;

use codespan::FileId;
use codespan_reporting::diagnostic::{Diagnostic, Label};
use datafusion::{
    logical_expr::{expr_fn, Expr, LogicalPlan, LogicalPlanBuilder, TableSource},
    sql::TableReference,
};
use maplit::hashmap;
use thiserror::Error;

use crate::{
    ast::{ColumnExpression, Count, Expression, Pipeline, Query},
    diagnostics,
};

#[derive(Debug, Error)]
pub enum LowerError {
    #[error("internal datafusion error")]
    DataFusionError(#[from] datafusion::error::DataFusionError),
    #[error("query is invalid")]
    InvalidQuery,
}

pub type LowerResult<T> = Result<T, LowerError>;

pub trait QueryContext {
    fn get_tenant_id(&self) -> &str;
}

pub struct Lowerer {
    ctx: Arc<dyn QueryContext>,
    file_id: FileId,
    table_source: Arc<dyn TableSource>,
    diagnostics: Vec<Diagnostic<FileId>>,
    table_name: String,
}

impl Lowerer {
    pub fn new(
        ctx: Arc<dyn QueryContext>,
        table_source: Arc<dyn TableSource>,
        file_id: FileId,
        table_name: String,
    ) -> Lowerer {
        Lowerer {
            ctx,
            file_id,
            table_source,
            table_name,
            diagnostics: vec![],
        }
    }

    pub fn diagnostics(&self) -> &[Diagnostic<FileId>] {
        &self.diagnostics
    }

    pub fn lower(&mut self, query: Query) -> LowerResult<LogicalPlan> {
        let table_ref = TableReference::Full {
            catalog: "prism".into(),
            schema: self.ctx.get_tenant_id().to_string().into(),
            table: query.table.name.into(),
        };
        let mut plan = LogicalPlanBuilder::scan(table_ref, self.table_source.clone(), None)?;
        for pipeline in query.pipelines {
            plan = self.lower_pipeline(plan, pipeline)?;
        }

        Ok(plan.build()?)
    }

    fn lower_pipeline(
        &mut self,
        builder: LogicalPlanBuilder,
        pipeline: Pipeline,
    ) -> LowerResult<LogicalPlanBuilder> {
        match pipeline {
            Pipeline::Count(count) => self.lower_count(builder, count),
        }
    }

    fn lower_count(
        &mut self,
        builder: LogicalPlanBuilder,
        count: Count,
    ) -> LowerResult<LogicalPlanBuilder> {
        let aggr_expr = expr_fn::count(Expr::Wildcard);
        let group_by: Vec<Expr> = if let Some(by) = count.by {
            vec![self.lower_expr(by)?]
        } else {
            vec![]
        };
        Ok(builder.aggregate(group_by, vec![aggr_expr])?)
    }

    fn lower_expr(&mut self, expr: Expression) -> LowerResult<Expr> {
        match expr {
            Expression::Column(column) => self.lower_column(column),
        }
    }

    fn lower_column(&mut self, column: ColumnExpression) -> LowerResult<Expr> {
        let schema = self.table_source.schema();
        if schema.field_with_name(&column.name.name).is_err() {
            self.diagnostics.push(
                diagnostics::column_does_not_exist(hashmap! {
                    "column" => column.name.name.clone(),
                    "table" => self.table_name.clone(),
                })
                .with_labels(vec![Label::primary(self.file_id, column.name.span)]),
            );

            return Err(LowerError::InvalidQuery);
        }

        Ok(expr_fn::col(column.name.name))
    }
}

#[cfg(test)]
mod tests {
    use std::sync::Arc;

    use codespan::Files;
    use datafusion::{
        arrow::datatypes::{DataType, Schema},
        logical_expr::builder::LogicalTableSource,
    };
    use expect_test::{expect, Expect};

    use super::{Lowerer, QueryContext};
    use crate::parser::parse;

    macro_rules! schema {
        ($($key:expr => $value:expr),*) => {
            {
                use datafusion::arrow::datatypes::{Schema, Field};

                let fields = vec![
                    $(
                        Field::new($key, $value, true),
                    )*
                ];
                Schema::new(fields)
            }
        };
    }

    struct TestQueryContext {
        pub tenant_id: String,
    }

    impl TestQueryContext {
        pub fn new(tenant_id: impl AsRef<str>) -> TestQueryContext {
            TestQueryContext {
                tenant_id: tenant_id.as_ref().to_string(),
            }
        }
    }

    impl QueryContext for TestQueryContext {
        fn get_tenant_id(&self) -> &str {
            &self.tenant_id
        }
    }

    fn check(ctx: TestQueryContext, schema: Schema, input: &str, expect: Expect) {
        let source = LogicalTableSource::new(Arc::new(schema));
        let mut files: Files<String> = Files::new();
        let fileid = files.add("query", input.to_string());
        let query = parse(input).unwrap();
        let mut lowerer = Lowerer::new(
            Arc::new(ctx),
            Arc::new(source),
            fileid,
            "http_requests".to_string(),
        );
        let plan = lowerer.lower(query).unwrap();
        expect.assert_debug_eq(&plan);
    }

    fn check_err(ctx: TestQueryContext, schema: Schema, input: &str, expect: Expect) {
        let source = LogicalTableSource::new(Arc::new(schema));
        let mut files: Files<String> = Files::new();
        let fileid = files.add("query", input.to_string());
        let query = parse(input).unwrap();
        let mut lowerer = Lowerer::new(
            Arc::new(ctx),
            Arc::new(source),
            fileid,
            "http_requests".to_string(),
        );
        let _ = lowerer.lower(query).unwrap_err();
        let diags = lowerer.diagnostics();
        expect.assert_debug_eq(&diags);
    }

    #[test]
    fn basic_count() {
        let ctx = TestQueryContext::new("tenant");
        let schema = schema! {
            "bytes" => DataType::UInt64,
            "method" => DataType::Utf8
        };

        let e = expect![[r#"
            Aggregate: groupBy=[[]], aggr=[[COUNT(*)]]
              TableScan: prism.tenant.http_requests
        "#]];
        check(ctx, schema, "http_requests | count", e);
    }

    #[test]
    fn basic_count_by() {
        let ctx = TestQueryContext::new("tenant");
        let schema = schema! {
            "bytes" => DataType::UInt64,
            "method" => DataType::Utf8
        };

        let e = expect![[r#"
            Aggregate: groupBy=[[prism.tenant.http_requests.method]], aggr=[[COUNT(*)]]
              TableScan: prism.tenant.http_requests
        "#]];
        check(ctx, schema, "http_requests | count by method", e);
    }

    #[test]
    fn count_by_invalid_column() {
        let ctx = TestQueryContext::new("tenant");
        let schema = schema! {
            "bytes" => DataType::UInt64,
            "method" => DataType::Utf8
        };

        let e = expect![[r#"
            [
                Diagnostic {
                    severity: Error,
                    code: Some(
                        "E0001",
                    ),
                    message: "column `something` does not exist on table `http_requests`",
                    labels: [
                        Label {
                            style: Primary,
                            file_id: FileId(
                                1,
                            ),
                            range: 25..34,
                            message: "",
                        },
                    ],
                    notes: [],
                },
            ]
        "#]];
        check_err(ctx, schema, "http_requests | count by something", e);
    }
}
