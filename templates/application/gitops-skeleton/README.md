# ${{ values.gitopsRepoName }}

This repository decides what is running. It holds no application code — that
lives in
[${{ values.destination.repo }}](https://github.com/${{ values.destination.owner }}/${{ values.destination.repo }}) —
and it holds no scripts that deploy anything. It is a description of the
desired state of ${{ values.appName }} in each environment, and ArgoCD's job is
to make the cluster match it.

The practical consequence: **merging a pull request here is deploying.** There
is no separate deploy step to run afterwards and no way to deploy without one.

## What is in here

```
chart/                          one Helm chart, rendered once per service
environments/
  dev/
    env-values.yaml             true of every service in dev
    services/<service>.yaml     that service's image and overrides in dev
  staging/
  prod/
argocd/applicationset.yaml      how ArgoCD turns the above into deployments
.github/workflows/validate.yml  renders every combination before a merge
.github/workflows/promote.yml   moves an image up one environment, by PR
```

A service exists in an environment exactly when it has a values file there.
Nothing needs to be registered anywhere else.

## How an image gets here

A merge to `main` in the source repository builds only the services that
commit touched, and — after they pass their tests, scans, signing and
attestation — opens **one** pull request here bumping `image.tag` and
`image.digest` for each of them under `environments/dev/`. That pull request is
the only thing the source pipeline can do to this repository. It has no cluster
credential and no bypass on this repository's branch rules.

## How an image gets to staging and production

Not by being built. Run the **promote** workflow (Actions → promote), pick the
environments and the services, and it opens a pull request copying the image
reference that is *already* the desired state of the lower environment into the
higher one. It cannot take a tag as free text and cannot reach into ECR, so an
image that was never merged into dev cannot appear in production — which is the
property the whole arrangement exists to guarantee.

Promotion goes one rung at a time: `dev → staging → prod`.

## Changing how a service runs

Memory limits, replica counts, probe timings, canary steps: edit the service's
values file for the environment you mean, and open a pull request. If the
change is true of every service in an environment, `env-values.yaml` is the
better place. If it is true of every service everywhere, it probably belongs in
`chart/values.yaml` — or, if it is true of every application on the platform,
in the platform chart rather than here.

## What still stands between a merge and a running pod

`validate.yml` renders every service in every environment and refuses a
malformed or mutable image reference. After that, Kyverno at admission checks
that the image is signed by the platform's build workflow, carries an SBOM
attestation, comes from ECR, is not `:latest`, declares probes and resource
limits, and runs under a restricted pod-security profile. A pull request that
merges is not yet a pod that runs.
