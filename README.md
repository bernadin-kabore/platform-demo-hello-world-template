# platform-demo-hello-world-template

One Backstage software template that turns "we're starting a new application"
into a running, observed, security-gated, progressively-deployed set of
microservices in one form submission — in as many languages as the application
needs.

Plus a second template that is not like it at all: **Ask the Platform**, for
requests the platform has no golden path for yet. See
[Ask the Platform](#ask-the-platform) below.

## What one form produces

A developer names the application, lists its services and picks a language for
each. What comes back is two repositories:

```
checkout-platform-source        the code
├── services/
│   ├── frontend/               TypeScript
│   ├── auth/                   Python
│   ├── payments/               Go
│   └── worker/                 Java
├── platform.yaml               what this application contains; CI reads it
├── .github/workflows/ci.yml    one pipeline, N services, change-aware
└── catalog-info.yaml           one System, one Component per service

checkout-platform-gitops        how it runs
├── chart/                      one Helm chart, rendered per service
├── environments/{dev,staging,prod}/
│   ├── env-values.yaml         true of every service in that environment
│   └── services/<name>.yaml    that service's image and overrides
├── argocd/applicationset.yaml  one Application per service per environment
└── .github/workflows/          validate.yml, promote.yml
```

**Why two repositories rather than one per service.** A change to what the code
does and a change to how much memory it gets are different changes, reviewed by
different people, with different consequences when they are wrong. Keeping them
apart is what lets the source pipeline advance a deployment by *asking* — a pull
request against the GitOps repository — rather than by writing to its own
protected branch, which is what it used to do.

**Why one repository for N services rather than N repositories.** Because
`checkout-platform` is the thing a team owns and reasons about, and four
repositories with four copies of a 220-line pipeline is four places for that
pipeline to drift. Services stay independently testable, buildable, scannable,
versioned and deployable — one repository does not mean one artifact. Each
service still produces its own image and still deploys on its own.

## Layout

```
catalog-info.yaml     Location entity listing the template.yaml files — the one
                       thing registered in Backstage's app-config.yaml
.github/workflows/
├── service-validate.yml  Reusable: lint, unit tests, CodeQL, the language's own
                       security linter, dependency scanning — for ONE service
├── service-build.yml     Reusable: build, scan, push, SBOM, sign, attest.
                       THE COSIGN IDENTITY KYVERNO VERIFIES LIVES HERE — read
                       the header before renaming or moving it
└── code-coverage.yml     Reusable: each language's coverage tool, failing under
                       a 70% line-coverage threshold, for one service directory
templates/
├── application/       The golden path. One template, one application.
│   ├── template.yaml
│   ├── skeleton/                source repository: manifest, pipeline, catalog
│   ├── gitops-skeleton/         GitOps repository: chart, environments, ArgoCD
│   ├── gitops-service-values/   one file per service per environment
│   └── platform-registration/   the one pointer PR'd into platform-demo-gitops
├── nodejs/skeleton/   Express + OTel auto-instrumentation (--require)
├── python/skeleton/   FastAPI + OTel zero-code instrumentation
├── go/skeleton/       net/http + otelhttp middleware
├── java/skeleton/     Spring Boot + the OTel Java agent (-javaagent)
└── ai-platform-request/  No skeleton: a form, not a scaffolder
    └── template.yaml
```

The four language directories hold **skeletons, not templates**. They have no
`template.yaml` of their own: the application template unpacks whichever ones
the developer asked for into `services/<name>/`. Adding a fifth language is a
skeleton plus one entry in the language enum, and no new entry point in the
portal.

## The pipeline

One `ci.yml` in the source repository, driven by `platform.yaml` rather than by
what it was generated with — so adding a service later needs no workflow change.

```
detect-changes      which services did this commit actually touch?
      │
      ├── validate  per affected service: lint, tests, CodeQL, SAST, SCA
      ├── coverage  per affected service, 70% line-coverage gate
      ├── semgrep   once, whole repository
      ├── gitleaks  once, whole repository
      │
   build            per affected service: build → Trivy → push → SBOM →
      │             cosign sign → cosign attest
      │
   gitops-pr        ONE pull request against the GitOps repository, carrying
      │             every image this build published
      │
   ci-gate          the single required status check
```

**Change-aware, deliberately dumbly.** A change under `services/auth/` builds
`auth`. A change to `platform.yaml`, `.github/workflows/` or `shared/` builds
everything, because none of those can be attributed to one service. There is no
dependency graph: over-building costs a runner, under-building ships a stale
image while reporting success.

**Nothing here deploys.** The pipeline holds no cluster credential and no
ruleset bypass. Its only effect on a running system is a pull request.

## Why the required status check is called `ci-gate`

Branch protection has to name the checks it requires, and a matrix job's name
contains the service it ran for — which is not knowable when the ruleset is
written. So one job fans every stage in and reports a single context. The
scaffolder's `platform:github:branch-protection` step requires `ci-gate`,
`semgrep` and `gitleaks` on `main`, `develop` and `release/*`, along with a pull
request, an approving review, and signed commits — automatically, the moment the
application is scaffolded. No manual step, no repo list to maintain.

(The four platform repositories — this one included — are protected the same way
but via Terraform instead, since they are static and don't come and go; see
`platform-demo-terraform-modules/envs/github-repos`, which also documents why
scaffold-time automation is the personal-account answer to a problem a GitHub
Organization solves for free with one org-wide ruleset.)

Contributing to any protected repository requires commit signing configured
locally (`git config commit.gpgsign true` with a GPG key, or `gpg.format ssh`
with an SSH key — either satisfies `required_signatures`).

## Environments and promotion

Only `dev` moves automatically. A successful build opens a pull request bumping
the image references under `environments/dev/`; merging it deploys.

Staging and production move by the **promote** workflow in the GitOps
repository, which copies an image reference that is *already* the desired state
of the environment below it and opens a pull request. It cannot take a tag as
free text and cannot reach into ECR, so an image that was never merged into dev
cannot appear in production. Promotion goes one rung at a time.

## The observability contract

Generated services depend on **OpenTelemetry and nothing else**. No skeleton
imports a Prometheus client library, names a log store, or knows that Tempo
exists, so the platform can change backends without touching any service.

| Signal | How it leaves the service |
|---|---|
| Traces | OTLP, from framework-level auto-instrumentation |
| Metrics | OTLP, using OpenTelemetry semantic conventions - no `/metrics` endpoint and no hand-rolled counters |
| Logs | Structured JSON on **stdout**, carrying `trace_id` and `span_id` |

**Logs are collected, not pushed.** The platform's collector agent tails
container stdout rather than the SDK exporting log records over the network,
because records buffered inside a process are lost when it crashes - and
crash logs are the ones you most need. Each language injects trace context
its own way: a `slog.Handler` in Go, a pino `mixin()` in Node,
`OTEL_PYTHON_LOG_CORRELATION` in Python, and the OTel agent's MDC in Java.

**Nothing environment-specific is baked into images.**
`OTEL_EXPORTER_OTLP_ENDPOINT` and `OTEL_RESOURCE_ATTRIBUTES` (carrying
`deployment.environment.name`, `service.namespace` and `service.version`)
are injected by the shared Helm chart at deploy time, so one immutable
artifact promotes across environments unchanged. Kubernetes-derived
attributes - pod, namespace, node, deployment - are deliberately *not* set
here; the collector's `k8sattributes` processor adds them.

A service that needs an alert of its own declares `prometheusRules` in its own
environment values file, which the chart renders into a `PrometheusRule`. A rule
every service should have belongs in `platform-demo-gitops`, keyed by label,
where it reaches teams who never thought to ask.

See
[`platform-demo-gitops/docs/observability/README.md`](../platform-demo-gitops/docs/observability/README.md)
for why each of these choices was made and what was rejected.

## Registering these templates in Backstage

Already wired via the single `url` catalog location in
[`platform-demo-backstage/app-config.yaml`](../platform-demo-backstage/app-config.yaml)
pointing at this repo's `catalog-info.yaml`. Once merged and Backstage reloads
its catalog, **New Application** and **Ask the Platform** appear under
**Create**.

## What a developer actually experiences

1. **Create** → **New Application** → name it, add its services and pick a
   language for each, "provision an S3 bucket? yes/no" → **Create**.
2. Two repositories exist, the catalog holds one System and one Component per
   service, and a pull request is open against `platform-demo-gitops`
   registering the application.
3. Merge that one PR → ArgoCD reads the application's own GitOps repository and
   deploys every service with Istio sidecar injection, canary rollout, and full
   observability already flowing.
4. Push code → CI builds only what changed and opens a deployment pull request.
   Merge it and the new image rolls out.

Zero YAML written by the developer, zero infrastructure requests filed,
regardless of how many languages the application mixes. That gap between "one
form" and "production-shaped application" is the actual thing this whole repo set
is demonstrating.

## Adding a 5th language

1. `templates/<lang>/skeleton/` — source, `Dockerfile` (non-root, multi-stage,
   minimal base image), and OTel instrumentation wired at the framework level
   rather than as hand-written spans. **No `.github/` and no chart**: the
   pipeline is the application's, and the chart is the GitOps repository's.
2. Add the language to two places in
   [`.github/workflows/service-validate.yml`](.github/workflows/service-validate.yml)
   (its `test` and `sca` jobs) and to
   [`code-coverage.yml`](.github/workflows/code-coverage.yml). Nothing goes in
   `service-build.yml` — building a container is language-agnostic, which is
   most of why the split is where it is.
3. Add it to the `language` enum in
   [`templates/application/template.yaml`](templates/application/template.yaml).

Nothing in `templates/application/gitops-skeleton/` or in
`platform-demo-backstage` changes.

## Ask the Platform

The application template covers the case the platform already has a golden path
for. `templates/ai-platform-request/` covers the case it does not.

It scaffolds nothing. There is no skeleton directory, no `publish:github`, no
`catalog:register`, no GitOps pull request step. It is a textarea and one
action:

```
Developer describes what they want
   ↓
platform:ai:request  (platform-demo-backstage/packages/backend/src/modules/ai-platform-request)
   ↓
AI Platform Agent    (platform-demo-ai-agent, running in the cluster)
   ↓
Terraform / Application / Security / Observability specialists
   ↓
Automated evals — deterministic checks, then a review model
   ↓
Pull requests, labelled ai-generated and needs-human-approval
```

From the pull request onwards it is the same pipeline a hand-written change
travels: the `ci-gate` stages, Checkov and `terraform plan` on infrastructure
changes, an approving human review required by branch protection, then ArgoCD.

**Why it lives here rather than in the portal repository.** The same reason the
application template does: templates are registered from this repo's root
`catalog-info.yaml`, so adding one is a directory plus one line there, and the
Backstage repository needs no change at all.

**Why the form asks for intent rather than a specification.** "We have no idea
when a service starts crash-looping" produces a better change than "add a
PrometheusRule with expr X". Working out the *how* is what the specialists are
for, and they know this platform's conventions — that Prometheus already
discovers rules in every namespace, that an alert without a documented response
is noise — better than a form field can capture.

**The agent is scoped to a service, not a repository.** Selecting a Component
authorizes it for that service's directory in its application's source
repository and that service's file in each environment of the GitOps repository
— not for a sibling service sharing the same repository, and not for anything
the application's services share. See
[`platform-demo-ai-agent/docs/trust-boundaries.md`](../platform-demo-ai-agent/docs/trust-boundaries.md).

**The agent cannot merge anything.** Its GitHub App holds no bypass on any
ruleset in the platform, including on its own repository.
