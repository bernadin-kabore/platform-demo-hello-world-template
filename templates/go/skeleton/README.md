# ${{ values.name }}

${{ values.description }}

A Go service inside this application's source repository. It builds its own
container image and deploys independently of its siblings — see the repository
root `README.md` for how the application fits together, and the GitOps
repository for how this service is actually running.

## Working on it

```bash
go test ./...
go run .
```

CI runs `go vet`, `go test` and golangci-lint for this directory whenever a
commit touches it, plus CodeQL, gosec, govulncheck, a Trivy filesystem scan,
and a 70% line coverage gate.
