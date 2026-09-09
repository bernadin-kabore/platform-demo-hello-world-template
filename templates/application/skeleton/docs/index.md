# ${{ values.appName }}

${{ values.description }}

Owned by `${{ values.owner }}`. Created by the **New Application** Backstage
template, which produced this repository and its GitOps counterpart,
`${{ values.gitopsRepoName }}`.

## The shape of it

One application, ${{ values.services | length }} independently deployable
service(s):

{% for service in values.services %}
- **${{ service.name }}** (${{ service.language }}) — ${{ service.description }}
{%- endfor %}

Each has its own directory under `services/`, its own container image, its own
tests and scans, and its own deployment state in each environment. They share a
repository and a pipeline; they do not share a release.

## What's already wired up

- **Observability**: OTel traces exported on every request, Prometheus metrics
  at `/metrics`, structured logs to stdout shipped to Elasticsearch. Services
  depend on the OpenTelemetry contract only, never on a specific backend.
- **Progressive delivery**: each service deploys via an Argo Rollout canary,
  gated by a Prometheus success-rate `AnalysisTemplate` — a bad deploy aborts
  itself.
- **Service mesh**: Istio sidecar injected automatically, mTLS enforced.
- **Security**: every image is scanned before it reaches the registry, SBOM'd,
  and cosign-signed in CI, then verified again at admission by Kyverno before
  it is allowed to run.
- **Change-aware CI**: only the services a commit touched are tested, scanned
  and rebuilt.

## Local development

`cd` into the service you are working on and follow its own README — each
language's toolchain is documented where that language lives, not here.

## Deploying

Nothing to run. Merging to `main` builds the affected services and opens one
pull request against
[`${{ values.gitopsRepoName }}`](https://github.com/${{ values.destination.owner }}/${{ values.gitopsRepoName }})
bumping their image references in `dev`. Merging *that* is the deployment.

Staging and production are reached by the **promote** workflow in the GitOps
repository, which moves an image that is already running in a lower
environment. An image cannot reach production by being built — only by being
promoted through every rung below it.
