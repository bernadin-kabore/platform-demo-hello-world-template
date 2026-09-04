# ${{ values.name }}

${{ values.description }}

A Python (FastAPI) service inside this application's source repository. It
builds its own container image and deploys independently of its siblings — see
the repository root `README.md` for how the application fits together, and the
GitOps repository for how this service is actually running.

## Working on it

```bash
pip install -r requirements-dev.txt
pytest
uvicorn app.main:app --reload
```

CI runs `ruff check app` and `pytest` for this directory whenever a commit
touches it, plus CodeQL, Bandit, `pip-audit`, a Trivy filesystem scan, and a
70% line coverage gate.
