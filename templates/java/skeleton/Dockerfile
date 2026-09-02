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
# agent - no OTel code anywhere in src/, unlike Go's manual SDK wiring.
ADD --chmod=644 https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/latest/download/opentelemetry-javaagent.jar /app/otel-javaagent.jar

# OTEL_SERVICE_NAME, the OTLP endpoint and resource attributes are injected by
# the Helm chart at deploy time rather than baked in here, so the same image
# promotes across environments unchanged.
#
# LOGBACK_MDC puts trace_id/span_id into the MDC, which logback-spring.xml
# writes into the JSON log line for the collector to promote. Logs are
# collected from stdout rather than exported by the agent (OTEL_LOGS_EXPORTER
# stays off) so that a JVM crash still produces readable output.
ENV OTEL_TRACES_EXPORTER=otlp \
    OTEL_METRICS_EXPORTER=otlp \
    OTEL_LOGS_EXPORTER=none \
    OTEL_INSTRUMENTATION_LOGBACK_MDC_ENABLED=true

EXPOSE 8080
ENTRYPOINT ["java", "-javaagent:/app/otel-javaagent.jar", "-jar", "/app/app.jar"]
