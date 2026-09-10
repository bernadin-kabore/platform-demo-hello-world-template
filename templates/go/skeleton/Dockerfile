# syntax=docker/dockerfile:1
FROM golang:1.26.4-alpine3.23 AS builder
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app ./...

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=builder /app /app
# OTEL_* configuration is injected by the Helm chart at deploy time, not
# baked in here, so the same image promotes across environments unchanged.
EXPOSE 8080
ENTRYPOINT ["/app"]
