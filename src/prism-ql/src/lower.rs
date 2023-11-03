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
    ast::{Count, Identifier, Pipeline, Query},
    diagnostics,
};

#[derive(Debug, Error)]
pub enum LowerError {
    #[error("failed to get table schema")]
    GetTableSchemaError(anyhow::Error),
    #[error("internal datafusion error")]
    DataFusionError(#[from] datafusion::error::DataFusionError),
}

pub type LowerResult<T> = Result<T, LowerError>;

pub trait QueryContext {
    fn get_tenant_id(&self) -> &str;

    fn get_table_source(&self, table: &str) -> anyhow::Result<Arc<dyn TableSource>>;
}

pub struct Lowerer {
    ctx: Arc<dyn QueryContext>,
    file_id: FileId,
    diagnostics: Vec<Diagnostic<FileId>>,
}

impl Lowerer {
    pub fn new(ctx: Arc<dyn QueryContext>, file_id: FileId) -> Lowerer {
        Lowerer {
            ctx,
            file_id,
            diagnostics: vec![],
        }
    }

    pub fn diagnostics(&self) -> &[Diagnostic<FileId>] {
        &self.diagnostics
    }

    pub fn lower(&mut self, query: Query) -> LowerResult<LogicalPlan> {
        let source = self.get_table_source(&query.table)?;
        let table_ref = TableReference::Full {
            catalog: "prism".into(),
            schema: self.ctx.get_tenant_id().to_string().into(),
            table: query.table.name.into(),
        };
        let mut plan = LogicalPlanBuilder::scan(table_ref, source, None)?;
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
        _: Count,
    ) -> LowerResult<LogicalPlanBuilder> {
        let aggr_expr = expr_fn::count(Expr::Wildcard);
        let group_by: Vec<Expr> = vec![];
        Ok(builder.aggregate(group_by, vec![aggr_expr])?)
    }

    fn get_table_source(&mut self, table: &Identifier) -> LowerResult<Arc<dyn TableSource>> {
        match self.ctx.get_table_source(&table.name) {
            Ok(schema) => Ok(schema),
            Err(e) => {
                self.diagnostics.push(
                    diagnostics::table_does_not_exist(hashmap! {
                        "table" => table.name.to_string(),
                    })
                    .with_labels(vec![Label::primary(self.file_id, table.span)]),
                );
                Err(LowerError::GetTableSchemaError(e.into()))
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use std::{collections::HashMap, sync::Arc};

    use codespan::Files;
    use datafusion::{
        arrow::datatypes::{DataType, SchemaRef},
        logical_expr::{builder::LogicalTableSource, TableSource},
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

    macro_rules! test_context {
        ($tenant_id:expr, $name:expr => $schema:expr) => {{
            use std::collections::HashMap;

            let mut sources = HashMap::new();
            sources.insert($name.to_string(), Arc::new($schema));
            TestQueryContext {
                tenant_id: $tenant_id.to_string(),
                sources,
            }
        }};
    }

    struct TestQueryContext {
        pub tenant_id: String,
        pub sources: HashMap<String, SchemaRef>,
    }

    impl QueryContext for TestQueryContext {
        fn get_tenant_id(&self) -> &str {
            &self.tenant_id
        }

        fn get_table_source(&self, table: &str) -> anyhow::Result<Arc<dyn TableSource>> {
            if let Some(schema) = self.sources.get(table) {
                let source = LogicalTableSource::new(schema.clone());
                Ok(Arc::new(source))
            } else {
                return Err(anyhow::anyhow!("table does not exist"));
            }
        }
    }

    fn check(ctx: TestQueryContext, input: &str, expect: Expect) {
        let mut files: Files<String> = Files::new();
        let fileid = files.add("query", input.to_string());
        let query = parse(input).unwrap();
        let mut lowerer = Lowerer::new(Arc::new(ctx), fileid);
        let plan = lowerer.lower(query).unwrap();
        expect.assert_debug_eq(&plan);
    }

    fn check_err(ctx: TestQueryContext, input: &str, expect: Expect) {
        let mut files: Files<String> = Files::new();
        let fileid = files.add("query", input.to_string());
        let query = parse(input).unwrap();
        let mut lowerer = Lowerer::new(Arc::new(ctx), fileid);
        let _ = lowerer.lower(query).unwrap_err();
        let diags = lowerer.diagnostics();
        expect.assert_debug_eq(&diags);
    }

    #[test]
    fn basic_count() {
        let ctx = test_context!("tenant", "http_requests" => schema! {
            "bytes" => DataType::UInt64,
            "method" => DataType::Utf8
        });

        let e = expect![[r#"
            Aggregate: groupBy=[[]], aggr=[[COUNT(*)]]
              TableScan: prism.tenant.http_requests
        "#]];
        check(ctx, "http_requests | count", e);
    }

    #[test]
    fn table_not_found() {
        let ctx = test_context!("tenant", "http_requests" => schema! {
            "bytes" => DataType::UInt64,
            "method" => DataType::Utf8
        });

        let e: Expect = expect![[r#"
            [
                Diagnostic {
                    severity: Error,
                    code: Some(
                        "E0001",
                    ),
                    message: "table `http_requests2` does not exist",
                    labels: [
                        Label {
                            style: Primary,
                            file_id: FileId(
                                1,
                            ),
                            range: 0..14,
                            message: "",
                        },
                    ],
                    notes: [],
                },
            ]
        "#]];
        check_err(ctx, "http_requests2 | count", e);
    }
}
