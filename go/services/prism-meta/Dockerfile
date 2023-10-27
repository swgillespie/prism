FROM golang:1.21.3-bullseye

WORKDIR /usr/src/prism
COPY go ./go
COPY go.mod .
COPY go.sum .
RUN /usr/local/go/bin/go build -o prism-meta ./go/services/prism-meta

CMD ["/usr/src/prism/prism-meta", "server"]

