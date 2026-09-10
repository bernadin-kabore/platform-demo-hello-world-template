# syntax=docker/dockerfile:1
# The builder's Go patch release matters as much as the base image: the
# standard library is compiled into the binary, so Trivy reports stdlib CVEs
# against whatever toolchain built it. 1.26.4 shipped with CVE-2026-56862 and
# friends, fixed in 1.26.6 -- the image built cleanly, scanned dirty, and the
# push was refused. Keep this at a patch the current advisories are fixed in.
FROM golang:1.26.8-alpine3.24 AS builder
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
