# ${{ values.name }}

${{ values.description }}

A Node.js (Express) service inside this application's source repository. It
builds its own container image and deploys independently of its siblings — see
the repository root `README.md` for how the application fits together, and the
GitOps repository for how this service is actually running.

## Working on it

```bash
npm ci
npm test
npm start
```

CI runs `npm run lint` and `npm test` for this directory whenever a commit
touches it, plus CodeQL, `npm audit`, a Trivy filesystem scan, and a 70% line
coverage gate.

**First-time setup**: run `npm install` once locally and commit the resulting
`package-lock.json` — `npm ci` in CI requires one and the scaffolder does not
generate it.
