FROM rust:1.73.0-bullseye

WORKDIR /usr/src/prism
COPY src ./src
COPY Cargo.toml .
COPY Cargo.lock .

RUN apt-get update
RUN apt-get install protobuf-compiler --yes
RUN cargo build --release --bin prism-ingest
COPY ./target/release/prism-ingest .
CMD ["prism-ingest"]