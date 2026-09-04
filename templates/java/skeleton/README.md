# ${{ values.name }}

${{ values.description }}

A Java (Spring Boot) service inside this application's source repository. It
builds its own container image and deploys independently of its siblings — see
the repository root `README.md` for how the application fits together, and the
GitOps repository for how this service is actually running.

## Working on it

```bash
mvn -B test
mvn spring-boot:run
```

CI runs `mvn -B test` for this directory whenever a commit touches it, plus
CodeQL, OWASP Dependency-Check, a Trivy filesystem scan, and a 70% line
coverage gate (JaCoCo).
