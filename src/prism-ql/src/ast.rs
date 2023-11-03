use codespan::Span;

#[derive(Clone, Debug)]
pub struct Query {
    pub span: Span,
    pub table: Identifier,
    pub pipelines: Vec<Pipeline>,
}

#[derive(Clone, Debug)]
pub enum Pipeline {
    Count(Count),
}

#[derive(Clone, Debug)]
pub struct Count {
    pub span: Span,
}

#[derive(Clone, Debug)]
pub struct Identifier {
    pub span: Span,
    pub name: String,
}

pub(crate) fn span(l: usize, r: usize) -> Span {
    Span::new(l as u32, r as u32)
}
