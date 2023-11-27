use lalrpop_util::lalrpop_mod;

use crate::ast;

lalrpop_mod!(pql);

pub fn parse(
    input: &str,
) -> Result<ast::Query, lalrpop_util::ParseError<usize, pql::Token<'_>, &'static str>> {
    let parser = pql::QueryParser::new();
    parser.parse(input)
}

#[cfg(test)]
mod tests {
    use expect_test::{expect, Expect};

    fn check(input: &str, expect: Expect) {
        let query = super::parse(input).unwrap();
        expect.assert_debug_eq(&query);
    }

    #[test]
    fn basic_count() {
        let e = expect![[r#"
            Query {
                span: Span {
                    start: ByteIndex(0),
                    end: ByteIndex(21),
                },
                table: Identifier {
                    span: Span {
                        start: ByteIndex(0),
                        end: ByteIndex(13),
                    },
                    name: "http_requests",
                },
                pipelines: [
                    Count(
                        Count {
                            span: Span {
                                start: ByteIndex(16),
                                end: ByteIndex(21),
                            },
                            by: None,
                        },
                    ),
                ],
            }
        "#]];
        check("http_requests | count", e);
    }
}
