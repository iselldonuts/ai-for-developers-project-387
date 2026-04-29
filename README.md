### Hexlet tests and linter status:
[![Actions Status](https://github.com/iselldonuts/ai-for-developers-project-387/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/iselldonuts/ai-for-developers-project-387/actions)

## Lighthouse report

The `lighthouse` GitHub Actions workflow runs every day at `23:50 MSK` (`20:50 UTC`) and can also be started manually from the Actions tab with `workflow_dispatch`.

The workflow builds the frontend, starts the Go app with in-memory storage, audits `/` and `/owner`, and uploads `lighthouse-reports` as an artifact. The job summary includes scores and a `Нужны правки` section with Lighthouse audits that need attention.

To run the same check locally:

```sh
make frontend-build
APP_STORAGE=memory APP_ADDR=:8080 go run ./cmd/calcal
(cd web && npm run lighthouse)
node scripts/summarize-lighthouse.mjs
```
