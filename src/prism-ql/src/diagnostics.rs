use std::collections::HashMap;

macro_rules! define_error {
    ($name:ident, $code:ident, $message:expr) => {
        pub fn $name(
            args: HashMap<&str, String>,
        ) -> ::codespan_reporting::diagnostic::Diagnostic<::codespan::FileId> {
            let map: HashMap<String, String> = args
                .into_iter()
                .map(|(k, v)| (k.to_string(), v.into()))
                .collect();
            let msg = strfmt::strfmt($message, &map).expect("invalid format string for diagnostic");
            ::codespan_reporting::diagnostic::Diagnostic::error()
                .with_code(stringify!($code))
                .with_message(msg)
        }
    };
}

define_error!(
    table_does_not_exist,
    E0001,
    "table `{table}` does not exist"
);
