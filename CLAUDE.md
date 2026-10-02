# CLAUDE.md — working rules for renters-ledger

This is a **learning project**. The owner is a Java/Spring Boot developer learning Go,
TypeScript, Postgres and Next.js on a 6-day plan (see `docs/DAILY_PLAN.md`).
The goal is for the owner to write the logic and understand it. Claude helps with
structure, tests, hints and reviews.

## Rules for every session

1. **Stay inside the current day's scope.** Do not implement anything beyond the current
   day in `docs/DAILY_PLAN.md` unless the owner explicitly asks. Stubs marked
   `TODO(DayN)` belong to Day N.
2. **Always keep tests passing.** The only acceptable failures are the current day's
   unfinished exercises. Tests for future days stay behind `t.Skip("DayN")`.
   Run `make test` (or `cd api && go test ./...`) before saying something is done.
3. **Hint before the answer.** When the owner pastes an error, first give a hint
   (what the error means and where to look). Give the full fix only if they ask
   or are still stuck after the hint.
4. **Explain Go idioms with Java comparisons.** Example: "an interface is satisfied
   implicitly, like implementing a Java interface without writing `implements`."
5. **Keep dependencies minimal.** Prefer the Go standard library. Allowed: `pgx`
   (Postgres driver) and `goose` (migrations). No web frameworks, no ORMs.
   On the front end, stick to Next.js, React, TypeScript and Zod unless asked.

## Conventions

- Go module: `github.com/lilynhuynh/renters-ledger/api` (lives in `api/`).
- Money is always `int64` cents. Never `float64`.
- Tests are table-driven (`[]struct{...}` + `t.Run`).
- Package boundaries are documented in `docs/ARCHITECTURE.md`. Respect the
  "must NOT" rules (for example, `policy` never imports `store` or `httpapi`).
- Do not run `git commit` or `git push` unless asked.
- `exercises/` is throwaway practice code. Do not refactor it or wire it into `api/`.
