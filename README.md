# Renters Ledger Project
Sample renters insurance backend and staff dashboard as a personal project.

Property managers upload tenant rosters -> system creates the policies -> move in defined lifecycle and policies can be update via payments recorded with audit trail.

## The problem

Renters insurance sold through property managers involves messy, high-stakes data flows:

- **Data arrives in formats you don't control.** Property managers send tenant rosters as files that have inconsistent dates, duplicate rows, and missing fields. Uploading the same file twice must not create duplicate policies.
- **Policies have a lifecycle with consequences.** A policy can be pending, active, past due, cancelled, lapsed, or reinstated. If the system allows a wrong transition, someone could be uninsured or billed incorrectly.
- **Payments are asynchronous and unreliable.** Payment processors send webhook events that can arrive twice, out of order, or days later. An ACH debit can look successful and then be returned.
- **Regulators and customers expect a record.** Staff need to answer "why is this policy in this state?" with a full history.

## What it solves

- **Safe ingestion.** Roster files are hashed and staged, validated row by row, and promoted to real records in one transaction, with a per-row error report. Re-uploading a file changes nothing.
- **A single source of truth for state.** One function decides which transitions are legal. Each transition is saved with a version check, so concurrent updates cannot overwrite each other, and an append-only event log records every change.
- **Idempotent payment handling.** Webhooks are signature-verified and deduplicated by event ID with a database unique constraint, so a repeated event has no effect.
- **Visibility for staff.** A dashboard lists policies, shows each one's event history, and lets staff cancel.

## Tech stack

| Layer | Technology |
|---|---|
| API | Go (standard library `net/http`, `log/slog`) |
| Database | PostgreSQL, with `pgx` and SQL migrations |
| Front end | Next.js (App Router), React, TypeScript, Zod for response validation |
| Local dev | Docker Compose for Postgres |
| Cloud concepts (documented, optional to deploy) | AWS ECS on Fargate, S3, SQS, Secrets Manager, CloudWatch |

## Key design decisions

- Money is stored as integer cents, and timestamps use `timestamptz`.
- The state machine is a pure function with table-driven tests for every legal and illegal transition.
- Idempotency is enforced by database constraints instead of application-level checks.
- Validation runs at the boundary: Zod on the front end and explicit checks in Go.

## Repository layout

```
api/        Go module (cmd/ = binaries, internal/ = packages, migrations/ = SQL)
web/        Next.js dashboard (created on Day 5)
docs/       ARCHITECTURE.md (design), DAILY_PLAN.md (6-day checklist)
exercises/  throwaway practice code
```

## How to run

Requires Go 1.22+, Docker (from Day 3) and Node 20+ (from Day 5).

```sh
make test                      # run all Go tests
make run-cli                   # Day 1: parse api/internal/ingest/testdata/roster_valid.csv
make run-cli ROSTER=internal/ingest/testdata/roster_bad_rows.csv
make run-api                   # Day 2+: HTTP API on :8080
make db-up / make db-down      # Day 3+: Postgres 16 in Docker
make fmt vet                   # format and static checks
```

## Design decisions (one line each)

- Money is `int64` cents, because floats cannot represent cents exactly.
- Timestamps are `timestamptz`, so every instant is unambiguous.
- Idempotency comes from unique constraints (file hash, webhook `event_id`), not from check-then-insert.
- `policy_events` is append-only, which gives a full audit trail.
- Optimistic locking uses a `version` column, so concurrent edits get a 409 instead of a silent overwrite.
- The `Store` interface makes the in-memory and Postgres stores swappable.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

## Status and scope

This is an educational project and not connected to any real payment processor or insurer. Webhooks are tested with signed sample payloads. The repository is organized as a monorepo with `api/`, `web/`, and `exercises/`, with a git tag for each day of the build.
