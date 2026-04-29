# Repository Guidelines

## Project Structure & Module Organization
The backend entry point lives in `cmd/calcal`. Core Go code is under `internal`: `app` wires dependencies, `config` loads environment settings, `service` holds business logic, `store` contains persistence interfaces and PostgreSQL implementations, and `http` owns routers, handlers, middleware, and transport mapping. Shared domain types live in `internal/model`, while service-facing payloads belong in `internal/dto`.

The TypeSpec API contract lives in `typespec/`, with `typespec/main.tsp` as the source of truth and generated OpenAPI emitted to `docs/openapi/openapi.yaml`. The frontend lives in `web/src`. Use `app` for bootstrapping, `pages` for route-level screens, `components` for reusable UI, `hooks` for shared React hooks, and `api` for HTTP client code. Infra assets live in `compose.yml`, `migrations/`, `scripts/`, and `docs/`.

## Build, Test, and Development Commands
Run backend and frontend tasks through `make`:

- `make backend-run` starts the Go API from `./cmd/calcal`.
- `make backend-test` runs `go test ./...` across all Go packages.
- `make frontend-install` installs `web` dependencies.
- `make frontend-dev` starts the Vite dev server.
- `make frontend-build` runs TypeScript compilation and creates a production bundle.
- `make typespec-install` installs the TypeSpec toolchain from `typespec/package.json`.
- `make typespec-compile` generates `docs/openapi/openapi.yaml` from `typespec/main.tsp`.

For local services, start PostgreSQL with `docker compose -f compose.yml up -d`.

## Coding Style & Naming Conventions
Target Go `1.26+` and write modern Go code from the start. Use `go fix ./...` when updating code to current language and toolchain conventions, but keep `gofmt` as the formatting standard because `go fix` does not replace it. Keep packages lowercase and focused, and use descriptive exported names like `BookingPage` or `New`. In React and TypeScript, use 2-space indentation, PascalCase for components, and kebab-case directories such as `pages/booking-page`. Keep transport concerns inside `internal/http` and avoid leaking HTTP types into `service` or `model`.

API-first changes should start in TypeSpec, not in manually written Go handlers. Update the contract in `typespec/`, regenerate OpenAPI, and treat the generated schema as the source for future code generation of the API boundary. Do not hand-maintain duplicate request/response contracts in Go when they can be generated from the spec.

## Testing Guidelines
There are no committed test files yet, so new features should add tests with the change. Place Go tests next to the package they cover in `*_test.go` files and favor table-driven tests for service and store logic. Run `make backend-test` before opening a PR. Frontend tests are not configured yet; if you add them, keep them under `web/src` and document the command in `web/package.json`.

## Commit & Pull Request Guidelines
Recent commits use short, imperative subjects such as `project structure` and `Add README.md`. Keep commit messages concise, scoped to one change, and written in the imperative voice. PRs should include a brief summary, note any config or schema changes, link the relevant issue, and attach screenshots for visible frontend updates.

## Configuration Tips
Copy values from `.env.example` and keep secrets out of git. The app expects `DATABASE_URL`, `APP_ADDR`, and `WEB_PORT`; the default local database is PostgreSQL on `localhost:5432`.
