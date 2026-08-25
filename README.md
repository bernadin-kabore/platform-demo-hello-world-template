# platform-demo-hello-world-template

Four Backstage software templates — one per language — that each turn "I
need a new service" into a running, observed, security-gated,
progressively-deployed application in one form submission.

## Layout

```
catalog-info.yaml     Location entity listing all four template.yaml files —
                       the one thing registered in Backstage's app-config.yaml
.github/workflows/
└── code-coverage.yml  Reusable workflow (workflow_call): runs each language's
                       coverage tool, fails under a 70% line-coverage threshold.
                       Called by every scaffolded service's own ci.yml as its
                       "coverage" job — see "Coverage gate" below.
common/                Shared across every language: the Helm chart (Argo
                       Rollout + Istio + ServiceMonitor + optional Crossplane
                       S3Bucket claim), catalog-info.yaml, and docs/mkdocs.yml
gitops-pr/             Shared: renders the one file (services/<name>/config.json)
                       PR'd into platform-demo-gitops
templates/
├── nodejs/            Express + OTel auto-instrumentation (--require)
├── python/            FastAPI + OTel zero-code instrumentation (opentelemetry-instrument)
├── go/                 net/http + otelhttp middleware
└── java/               Spring Boot + the OTel Java agent (-javaagent)
    └── template.yaml  + skeleton/  (language-specific: source, Dockerfile, CI)
```

## Coverage gate

Every skeleton's `ci.yml` has a `coverage` job that calls
[`code-coverage.yml`](.github/workflows/code-coverage.yml) with its
language (`c8` for Node.js, `pytest-cov` for Python, `go test -cover` for
Go, JaCoCo for Java) and fails the run if line coverage is under 70%. That
alone doesn't block a merge — it's just a status check. What actually
blocks it is the `branch-protection` step every `template.yaml` runs right
after `publish:github` creates the new repo: it calls the custom
`platform:github:branch-protection` scaffolder action
(`platform-demo-backstage/packages/backend/src/modules/branch-protection`),
which creates a GitHub ruleset on that repo making `coverage / check`
(along with `test`, `sast`, `sca`) a **required** status check on `main`,
`develop`, and `release/*`, requiring a PR (no direct pushes), and
requiring signed commits — automatically, for every service, the moment
it's scaffolded. No manual step, no repo list to maintain. (The 4 platform
repos — this one included — are protected the same way but via Terraform
instead, since they're static and don't come and go; see
`platform-demo-terraform-modules/envs/github-repos`, which also documents
why scaffold-time automation is the personal-account answer to a problem a
GitHub Organization solves for free with one org-wide ruleset.)

Contributing to any protected repo requires commit signing configured
locally (`git config commit.gpgsign true` with a GPG key, or
`gpg.format ssh` with an SSH key — either satisfies `required_signatures`).

Each `template.yaml`'s scaffolder steps fetch **both** `../../common` and
its own `./skeleton` into the same new repo, so the Helm chart, docs, and
`catalog-info.yaml` are never duplicated four times — only source code,
`Dockerfile`, and the language-specific parts of `.github/workflows/ci.yml`
differ between languages. Every skeleton exposes the same three endpoints
(`/healthz`, `/readyz`, `/metrics`) and the same OTLP export target, which
is what lets one shared Helm chart deploy any of them unmodified.

## Registering these templates in Backstage

Already wired via the single `url` catalog location in
[`platform-demo-backstage/app-config.yaml`](../platform-demo-backstage/app-config.yaml)
pointing at this repo's `catalog-info.yaml`. Once merged and Backstage
reloads its catalog, **Node.js Service**, **Python Service**, **Go
Service**, and **Java Service** all appear under **Create**.

## What a developer actually experiences

1. **Create** → pick a language tile → fill in name, owner, "provision an
   S3 bucket? yes/no" → **Create**.
2. ~30 seconds later: a new GitHub repo exists, CI has already run once and
   pushed a signed, SBOM'd image, and a PR is open against
   `platform-demo-gitops`.
3. Merge that one PR → ArgoCD deploys the service with Istio sidecar
   injection, canary rollout, and full observability already flowing —
   zero YAML written by the developer, zero infrastructure requests filed,
   regardless of which language they picked.

That gap between "one form" and "production-shaped service" — reproduced
identically across four different language ecosystems — is the actual
thing this whole repo set is demonstrating.

## Adding a 5th language

1. `templates/<lang>/skeleton/` — source, `Dockerfile` (non-root, multi-stage,
   minimal base image), OTel instrumentation wired at the framework level
   (not hand-written spans), and `.github/workflows/ci.yml` following the
   same five-job shape every other language uses: `test` → `sast` → `sca` →
   `build-scan-sign` → `update-manifests`.
2. `templates/<lang>/template.yaml` — copy an existing one, change
   `metadata.name`/`title`/`tags`, the `localDevCommands` value, and the two
   `fetch:template` URLs stay `../../common` and `./skeleton`.
3. Add one line to `catalog-info.yaml`'s `spec.targets`. Nothing else in
   `common/`, `gitops-pr/`, or `platform-demo-backstage` changes.
