APP_NAME := calcal
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/calcal?sslmode=disable

.PHONY: backend-generate-api
backend-generate-api:
	mkdir -p internal/http/api
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.6.0 -generate types,std-http,skip-prune -package api -o internal/http/api/api.gen.go docs/openapi/openapi.yaml

.PHONY: backend-migrate-up
backend-migrate-up:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$(DATABASE_URL)" up

.PHONY: backend-run
backend-run:
	go run ./cmd/$(APP_NAME)

.PHONY: backend-test
backend-test:
	go test ./...

.PHONY: frontend-install
frontend-install:
	cd web && npm install

.PHONY: frontend-dev
frontend-dev:
	cd web && npm run dev

.PHONY: frontend-build
frontend-build:
	cd web && npm run build

.PHONY: typespec-install
typespec-install:
	cd typespec && npm install

.PHONY: typespec-compile
typespec-compile:
	cd typespec && npm run compile
