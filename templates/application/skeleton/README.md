# ${{ values.appName }}

${{ values.description }}

This repository holds the code for every service in ${{ values.appName }}. It
does not hold anything about how those services are deployed — that is in
[${{ values.gitopsRepoName }}](https://github.com/${{ values.destination.owner }}/${{ values.gitopsRepoName }}),
and the split is deliberate: a change to what the code does and a change to how
much memory it gets are different changes, reviewed by different people, with
different consequences when they are wrong.

## Layout

```
platform.yaml            what this application contains; CI reads it
services/
{%- for service in values.services %}
  ${{ service.name }}/ (${{ service.language }})
{%- endfor %}
catalog-info.yaml        the System and its Components, for the portal
.github/workflows/ci.yml the pipeline; not application-specific
```

Each service is independently testable, buildable, scannable and deployable,
and produces its own container image. One repository does not mean one
artifact.

## Working on a service

Everything is local to its directory. `cd services/<name>` and follow the
README there for that language's commands — the toolchain CI uses is the
toolchain the service's own README documents, and neither invents anything the
other does not.

## What happens when you open a pull request

CI works out which services your commit actually touched — a change under
`services/auth/` tests and scans `auth` and nothing else — and runs, for each
of them: lint, unit tests, a 70% line-coverage gate, CodeQL, the language's own
security linter, and dependency scanning. Semgrep and a secrets scan run once
across the whole repository. All of it must pass before the pull request can
merge.

A change to `platform.yaml`, to `.github/workflows/`, or to `shared/` cannot be
attributed to one service, so it rebuilds all of them. That is intentional:
over-building costs a runner, under-building ships a stale image while
reporting success.

## What happens when it merges

For each affected service: the image is built, scanned by Trivy *before* it is
allowed into ECR, pushed under the commit SHA, given an SPDX SBOM, signed
keylessly, and the SBOM attested to it. Then — once, for the whole build — CI
opens a single pull request against the GitOps repository bumping the image
references for every service it just published.

Nothing here deploys. This pipeline has no cluster credential; the only thing
it can do to a running system is ask, by pull request, for the desired state to
change.

## Adding a service later

1. Create `services/<name>/` with a Dockerfile and the language's usual layout.
2. Add three lines to `platform.yaml`.
3. Add `environments/<env>/services/<name>.yaml` in the GitOps repository.
4. Add a `Component` to `catalog-info.yaml`.

No workflow changes. The pipeline is driven by the manifest, not by what it was
generated with.
