# Architecture

## 1. Components and data flow

```
                       ┌──────────────────────────────┐
   Staff browser ─────▶│  Next.js app (web/)          │
                       │  pages + Zod-validated fetch │
                       └──────────────┬───────────────┘
                                      │ JSON over HTTP
                                      ▼
  Property manager      ┌──────────────────────────────────────────────────────┐
  roster CSV upload ───▶│  Go API (api/cmd/api)                                │
  (or rostercli)        │                                                      │
                        │  httpapi ──▶ ingest   (parse + validate CSV rows)    │
  Payment processor     │     │                                                │
  webhooks (signed) ───▶│     ├────▶ payments (verify HMAC, dedupe event_id)   │
                        │     │            │                                   │
                        │     ▼            ▼                                   │
                        │   policy.Transition (pure lifecycle rules)           │
                        │     │                                                │
                        │     ▼                                                │
                        │   store.Store  ── MemoryStore (Day 2, tests)         │
                        │                └─ PostgresStore (Day 3+)             │
                        └──────────────────────────┬───────────────────────────┘
                                                   │ pgx (SQL)
                                                   ▼
                                    ┌───────────────────────────┐
                                    │ Postgres 16               │
                                    │ customers, policies,      │
                                    │ policy_events, roster_    │
                                    │ uploads, webhook_events   │
                                    └───────────────────────────┘
```

Dependency direction always points **inward** toward `policy`:
`cmd → httpapi → {ingest, payments, store} → policy`. Nothing imports `cmd` or `httpapi`.

## 2. Packages: responsibility and boundaries

| Package | Single responsibility | Must NOT |
|---|---|---|
| `cmd/rostercli` | CLI: read a file path, call `ingest.ParseRoster`, print a summary. | Contain parsing or validation logic. |
| `cmd/api` | Composition root: read config, build the store, logger and handler, start the server. | Contain business logic or SQL. It only wires things together. |
| `internal/ingest` | Turn CSV bytes into `[]Tenant` plus `[]RowError`. | Touch the database, HTTP, or policy status. It must not import `store` or `httpapi`. |
| `internal/policy` | Domain types (`Policy`, `Status`, `Event`) and the pure `Transition` function. | Import `store`, `httpapi`, `payments`, `database/sql`, `net/http`, or read the clock. |
| `internal/store` | Persist and load policies behind the `Store` interface; enforce version checks. | Know about HTTP (no status codes, no `net/http`). It must not decide which transitions are legal; it calls `policy.Transition` for that. |
| `internal/httpapi` | Routing, JSON decode and encode, mapping errors to status codes, middleware. | Contain SQL or lifecycle rules. It only talks to `store.Store` (the interface), never a concrete store. |
| `internal/payments` | Verify webhook signatures, dedupe by event ID, translate to `policy.Event`. | Import `httpapi`, or set `Policy.Status` directly without `policy.Transition`. |

`internal/` is enforced by the Go compiler: code outside `api/` cannot import these packages.

## 3. Request flow: `GET /policies/{id}`

1. **net/http server** accepts the connection and starts a goroutine for the request
   (similar to a Tomcat worker thread, but much cheaper).
2. **`Logging` middleware** records the start time and wraps the `ResponseWriter` in a
   `statusRecorder`.
3. **`http.ServeMux`** matches the pattern `"GET /policies/{id}"` and calls `Handler.getPolicy`.
4. **`getPolicy`** reads `id := r.PathValue("id")` and calls
   `h.store.Get(r.Context(), id)`. `r.Context()` is cancelled if the client disconnects.
5. **`store.Store.Get`** dispatches to whichever implementation `main` injected:
   - `MemoryStore`: read-lock, map lookup.
   - `PostgresStore`: `SELECT ... FROM policies WHERE id = $1`, `Scan` into a `policy.Policy`;
     `pgx.ErrNoRows` becomes `store.ErrNotFound`.
6. **Back in `getPolicy`**:
   `errors.Is(err, store.ErrNotFound)` → `404`; any other error → log it and return `500`;
   success → `writeJSON(w, 200, p)`.
7. **`Logging` middleware** logs `method=GET path=/policies/abc status=200 duration=…`.

## 4. Data model (Day 3 plan, no SQL yet)

**customers**
- `id` uuid PK
- `name` text not null
- `email` text not null, unique (case-insensitive)
- `created_at` timestamptz not null default now()

**policies**
- `id` uuid PK
- `customer_id` uuid FK → customers, not null
- `unit` text not null
- `status` text not null, CHECK in (quoted, pending, active, past_due, cancelled, lapsed, reinstated)
- `premium_cents` bigint not null, CHECK ≥ 0
- `effective_date` date not null
- `version` bigint not null default 1 (optimistic lock)
- `created_at`, `updated_at` timestamptz not null
- unique (customer_id, unit, effective_date), so re-ingesting a roster cannot duplicate a policy

**policy_events** (append-only, never UPDATE or DELETE)
- `id` bigserial PK
- `policy_id` uuid FK → policies, not null
- `event` text not null (submit, payment_succeeded, …)
- `from_status`, `to_status` text not null
- `actor` text (staff user, "system", "webhook:<event_id>")
- `occurred_at` timestamptz not null default now()
- index on (policy_id, occurred_at)

**roster_uploads**
- `id` uuid PK
- `file_sha256` text not null, **unique**, so uploading the same file twice is a no-op
- `filename` text, `row_count`, `valid_count`, `error_count` int
- `errors` jsonb (the RowError list)
- `uploaded_at` timestamptz not null

**webhook_events**
- `id` bigserial PK
- `event_id` text not null, **unique** (the dedupe key)
- `type` text not null
- `payload` jsonb not null (raw body, for audit and replay)
- `received_at` timestamptz not null, `processed_at` timestamptz null

## 5. Policy lifecycle

```
 quoted ──submit──▶ pending ──payment_succeeded──▶ active
                                                    │  ▲
                                     payment_failed │  │ payment_succeeded
                                                    ▼  │
                                                   past_due ◀────────────────┐
                                                    │         payment_failed │
                               grace_period_expired │                        │
                                                    ▼                        │
                                                   lapsed ──reinstate──▶ reinstated
                                                                             │
                                                                             └──payment_succeeded──▶ active

 cancel:  quoted | pending | active | past_due | reinstated  ──▶  cancelled  (terminal)
 lapsed accepts only reinstate. cancelled accepts nothing.
```

Transition table (the authoritative version is the doc comment on `policy.Transition`):

| From | Event | To |
|---|---|---|
| quoted | submit | pending |
| quoted | cancel | cancelled |
| pending | payment_succeeded | active |
| pending | cancel | cancelled |
| active | payment_failed | past_due |
| active | cancel | cancelled |
| past_due | payment_succeeded | active |
| past_due | grace_period_expired | lapsed |
| past_due | cancel | cancelled |
| lapsed | reinstate | reinstated |
| reinstated | payment_succeeded | active |
| reinstated | payment_failed | past_due |
| reinstated | cancel | cancelled |

Any other (status, event) pair is illegal and returns `ErrIllegalTransition`.

## 6. Go ↔ Spring Boot mapping

| Go here | Spring Boot equivalent |
|---|---|
| `go.mod` | `pom.xml` / `build.gradle` |
| `cmd/api/main.go` (manual wiring) | `@SpringBootApplication` + DI container |
| `NewHandler(store, logger)` | Constructor injection |
| `httpapi.Handler` | `@RestController` |
| `mux.HandleFunc("GET /policies/{id}", …)` | `@GetMapping("/policies/{id}")` |
| `r.PathValue("id")` | `@PathVariable` |
| `json.NewDecoder(r.Body).Decode(&req)` | `@RequestBody` + Jackson |
| `httpapi.Logging` middleware | Servlet `Filter` / `HandlerInterceptor` |
| `store.Store` interface | Spring Data `Repository` interface |
| `MemoryStore` / `PostgresStore` | Two `@Repository` beans selected by `@Profile` |
| `pgxpool.Pool` | HikariCP `DataSource` + `JdbcTemplate` |
| `Policy.Version` + `WHERE version = $n` | JPA `@Version` / `OptimisticLockException` |
| `goose` migrations | Flyway |
| `policy.Status` constants | Java `enum` |
| `policy.Transition` | Domain service / Spring State Machine (but just a function) |
| `error` return values, `errors.Is` | Checked exceptions, `catch (SpecificException e)` |
| `context.Context` | Request-scoped timeout and cancellation (no direct equivalent) |
| `log/slog` | SLF4J + Logback |
| Table-driven `go test` | JUnit 5 `@ParameterizedTest` |
| `net/http/httptest` | `MockMvc` |
| `internal/` directory | Package-private visibility / JPMS module boundary |
| Struct tags `` `json:"id"` `` | `@JsonProperty("id")` |

## 7. Design decisions

- **Money in integer cents (`int64` / `bigint`).** Floats cannot represent most decimal
  amounts exactly, and the errors compound. Format as dollars only at the edges (CLI, UI).
- **`timestamptz` for every timestamp.** Postgres stores an absolute instant, so there is no
  "which timezone was this?" ambiguity. Go code works in UTC. Calendar dates such as
  move-in date and effective date use `date`.
- **Idempotency via unique constraints.** `roster_uploads.file_sha256` and
  `webhook_events.event_id` are UNIQUE. The insert itself is the check, which is race-free,
  unlike "SELECT then INSERT".
- **Append-only events.** `policy_events` rows are never updated or deleted, which gives
  a full audit trail that answers "why is this policy in this state?"
- **Optimistic locking with a `version` column.** Updates use `WHERE id = $1 AND version = $2`.
  Zero rows affected means someone else changed it first, so the API returns 409.
- **`Store` interface so memory and Postgres are swappable.** Day 2 builds the HTTP layer
  with no database, tests stay fast, and Day 3 swaps in Postgres without touching `httpapi`.
