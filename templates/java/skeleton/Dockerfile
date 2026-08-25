# syntax=docker/dockerfile:1
FROM maven:3.9.8-eclipse-temurin-21 AS builder
WORKDIR /build
COPY pom.xml .
RUN mvn -B dependency:go-offline
COPY src ./src
RUN mvn -B package -DskipTests

FROM gcr.io/distroless/java21-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=builder /build/target/*.jar app.jar
# Java is the one language here with a true zero-code auto-instrumentation
# agent — no OTel code anywhere in src/, unlike Go's manual SDK wiring.
ADD --chmod=644 https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/latest/download/opentelemetry-javaagent.jar /app/otel-javaagent.jar

ENV OTEL_SERVICE_NAME="${{ values.name }}" \
    OTEL_EXPORTER_OTLP_ENDPOINT="http://otel-collector-gateway.observability.svc.cluster.local:4317" \
    OTEL_TRACES_EXPORTER=otlp \
    OTEL_METRICS_EXPORTER=none \
    OTEL_LOGS_EXPORTER=none

EXPOSE 8080
ENTRYPOINT ["java", "-javaagent:/app/otel-javaagent.jar", "-jar", "/app/app.jar"]
