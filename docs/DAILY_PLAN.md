# 6-day plan (2 hours per day)

Each day ends with green tests and a git tag (`day-1`, `day-2`, …).
Spend about 20 minutes reading or doing the Go Tour in `exercises/`, then build.

---

## Day 1: Go fundamentals and the roster CLI

Goal: write real Go (structs, slices, errors, `io.Reader`, table-driven tests).

- [X] Go Tour: basics, flow control, structs, slices, maps, methods, interfaces, errors (in `exercises/`)
- [X] Read `api/internal/ingest/parser.go` and `parser_test.go` and understand each test case
- [X] Implement `RowError.Error()` → `TestRowErrorError` passes
- [X] Implement `ParseRoster` → all `TestParseRoster` subtests pass
- [ ] `make run-cli` and `make run-cli ROSTER=internal/ingest/testdata/roster_bad_rows.csv` print sensible output
- [ ] `make fmt vet test` is clean
- [ ] Notes: what surprised you compared with Java (errors as values, no exceptions, zero values)

## Day 2: HTTP API with an in-memory store

Goal: `net/http` (Go 1.22 routing), JSON, middleware, `httptest`, goroutines and mutexes.

- [ ] Implement `MemoryStore` (map + `sync.RWMutex`)
- [ ] Implement `Handler.Routes` and the handlers in `httpapi/handler.go`, plus `writeJSON`
- [ ] Implement `Logging` middleware and `statusRecorder.WriteHeader`
- [ ] Remove `t.Skip` in `handler_test.go`; add create, bad body, cancel and stale-version cases
- [ ] Wire everything in `cmd/api/main.go`; `make run-api`; try it with `curl`
- [ ] Run `go test -race ./...` once to see the race detector

## Day 3: Postgres, migrations and the Postgres store

Goal: schema design, `pgx`, `goose`, mapping rows to structs by hand.

- [ ] `make db-up`; connect with `psql` or a GUI
- [ ] Write `migrations/0001_init.sql` from the data model in `ARCHITECTURE.md` (goose Up/Down)
- [ ] `go get github.com/jackc/pgx/v5`; install goose; run the migration
- [ ] Implement `PostgresStore.List/Get/Create`
- [ ] `main.go`: use Postgres when `DATABASE_URL` is set, otherwise memory
- [ ] Integration test that skips when `DATABASE_URL` is unset
- [ ] Insert a roster through the API or CLI and query it in `psql`

## Day 4: State machine, optimistic locking and webhooks

Goal: correctness under retries and concurrency.

- [ ] Remove `t.Skip` in `policy/state_test.go`; implement `policy.Transition` until it is green
- [ ] `PostgresStore.Cancel`: one transaction, `WHERE version = $n`, append to `policy_events`
- [ ] Test: two cancels with the same version → second gets `ErrVersionConflict` → HTTP 409
- [ ] `payments.VerifySignature` with HMAC-SHA256 + `hmac.Equal`; table test with good, bad and malformed signatures
- [ ] `webhook_events.event_id UNIQUE`; `Deduper` backed by Postgres; replaying the same event is a no-op
- [ ] `POST /webhooks/payments` route that calls `payments.Process`

## Day 5: Next.js + TypeScript dashboard

Goal: TypeScript basics, App Router, server components, Zod at the boundary.

- [ ] `npx create-next-app@latest web --ts --app` (replace `web/README.md`)
- [ ] Zod schema for `Policy` that mirrors the Go struct; parse every API response
- [ ] `/policies` page: table with status badges
- [ ] `/policies/[id]` page: details + event history
- [ ] Cancel button → `POST /policies/{id}/cancel`; show a 409 conflict message nicely
- [ ] CORS or a Next.js rewrite proxy to `localhost:8080`

## Day 6: Roster upload end to end, polish and rehearsal

Goal: tie it together and be able to explain every decision.

- [ ] `POST /rosters`: hash the file (SHA-256), insert `roster_uploads` (unique hash = idempotent), parse with `ingest`, create policies in one transaction, return a RowError report
- [ ] Upload form in the dashboard showing per-row errors
- [ ] Graceful shutdown in `main.go`; config from env vars
- [ ] Write an AWS deployment sketch in `docs/` (ECS Fargate, RDS, S3 for uploads, SQS for webhooks, Secrets Manager, CloudWatch)
- [ ] Re-read `ARCHITECTURE.md` and explain each design decision out loud in under a minute
- [ ] Tag `day-6`; list what you would do next
