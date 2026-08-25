# syntax=docker/dockerfile:1
FROM golang:1.22.5-alpine3.20 AS builder
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app ./...

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=builder /app /app
ENV OTEL_EXPORTER_OTLP_ENDPOINT="otel-collector-gateway.observability.svc.cluster.local:4317"
EXPOSE 8080
ENTRYPOINT ["/app"]
