fn main() {
    tonic_build::compile_protos("proto/meta.proto").unwrap();
}
