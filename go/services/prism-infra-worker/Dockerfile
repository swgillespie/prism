FROM rust:1.73.0-bullseye

RUN apt-get update
RUN apt-get install protobuf-compiler --yes

WORKDIR /usr/src/prism
COPY src ./src
COPY Cargo.toml .
COPY Cargo.lock .
RUN cargo build --release --bin prism-ingest
RUN cp ./target/release/prism-ingest .

COPY --from=golang:1.21.3-bullseye /usr/local/go/ /usr/local/go/
COPY go ./go
COPY go.mod .
COPY go.sum .
RUN /usr/local/go/bin/go build -o prism-infra-worker ./go/services/prism-infra-worker

CMD ["prism-infra-worker", "--config", "/etc/prism/config.yaml"]


