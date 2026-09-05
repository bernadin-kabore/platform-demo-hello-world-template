package io.acme.platform.helloworld;

import java.util.Map;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class HelloController {

    @Value("${spring.application.name}")
    private String serviceName;

    @Value("${app.version:dev}")
    private String appVersion;

    @GetMapping("/")
    public Map<String, String> hello() {
        return Map.of("message", "Hello from " + serviceName + "!", "version", appVersion);
    }

    // Hand-rolled rather than pointed at Spring Actuator's /actuator/health,
    // so the shared Helm chart (common/chart) can use the exact same
    // /healthz, /readyz probe paths across every language. There is no
    // /metrics endpoint any more - metrics are pushed over OTLP by the
    // OpenTelemetry Java agent.
    @GetMapping("/healthz")
    public String healthz() {
        return "ok";
    }

    @GetMapping("/readyz")
    public String readyz() {
        return "ok";
    }
}
