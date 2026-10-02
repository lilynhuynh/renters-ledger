# Makefile — shortcuts, roughly like Maven/Gradle tasks.
# All Go commands run inside api/ because that is where go.mod lives (like a pom.xml).

ROSTER ?= internal/ingest/testdata/roster_valid.csv

.PHONY: test run-cli run-api db-up db-down fmt vet

test:            ## Run all Go tests (like `mvn test`)
	cd api && go test ./...

run-cli:         ## Day 1: parse a roster CSV. Override with: make run-cli ROSTER=path/to/file.csv
	cd api && go run ./cmd/rostercli $(ROSTER)

run-api:         ## Day 2+: start the HTTP API on :8080
	cd api && go run ./cmd/api

db-up:           ## Day 3+: start Postgres 16 in Docker
	docker compose up -d db

db-down:         ## Stop Postgres (data stays in the named volume)
	docker compose down

fmt:             ## Format all Go code (gofmt is the one true style; no config)
	cd api && gofmt -l -w .

vet:             ## Static checks that catch common mistakes (like SpotBugs, lighter)
	cd api && go vet ./...
