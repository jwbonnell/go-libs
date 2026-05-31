FROM golang:1.25 AS builder
WORKDIR /app

# Copy library root (required by replace directive in sample-api/go.mod)
COPY go.mod go.sum ./
COPY pkg/ ./pkg/

# Copy sample-api module
COPY examples/observability/sample-api/ ./examples/observability/sample-api/

WORKDIR /app/examples/observability/sample-api
RUN go mod download && go build -o /sample-api .

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates wget && rm -rf /var/lib/apt/lists/*
COPY --from=builder /sample-api /sample-api
ENTRYPOINT ["/sample-api"]
