// Package httpapi translates HTTP to and from domain calls: decode request, call Store, encode JSON.
// It must NOT contain business rules (those live in policy) or SQL (that lives in store).
package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/lilynhuynh/renters-ledger/api/internal/store"
)

// Handler is like a @RestController. Its dependencies are plain struct fields filled by
// NewHandler. That is constructor injection done by hand, with no DI container.
type Handler struct {
	store  store.Store // the interface, so memory and Postgres are swappable
	logger *slog.Logger
}

func NewHandler(s store.Store, logger *slog.Logger) *Handler {
	return &Handler{store: s, logger: logger}
}

// Routes builds the router. Go 1.22's ServeMux understands methods and path variables,
// so "GET /policies/{id}" works like @GetMapping("/policies/{id}").
//
// TODO(Day2): register these and wrap the mux with Logging middleware:
//
//	mux.HandleFunc("GET /healthz", h.health)
//	mux.HandleFunc("GET /policies", h.listPolicies)
//	mux.HandleFunc("GET /policies/{id}", h.getPolicy)
//	mux.HandleFunc("POST /policies", h.createPolicy)
//	mux.HandleFunc("POST /policies/{id}/cancel", h.cancelPolicy)
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	return mux
}

// TODO(Day2): write 200 with {"status":"ok"}.
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	panic("TODO(Day2): health")
}

// TODO(Day2): h.store.List(r.Context()), then writeJSON 200.
func (h *Handler) listPolicies(w http.ResponseWriter, r *http.Request) {
	panic("TODO(Day2): listPolicies")
}

// TODO(Day2): id := r.PathValue("id") (like @PathVariable); h.store.Get;
// errors.Is(err, store.ErrNotFound) -> 404, other error -> 500, else 200 JSON.
func (h *Handler) getPolicy(w http.ResponseWriter, r *http.Request) {
	panic("TODO(Day2): getPolicy")
}

// TODO(Day2): json.NewDecoder(r.Body).Decode into a request struct (like @RequestBody),
// validate (non-empty unit, PremiumCents >= 0) -> 400 on failure, Create, 201.
func (h *Handler) createPolicy(w http.ResponseWriter, r *http.Request) {
	panic("TODO(Day2): createPolicy")
}

// TODO(Day2): read expected version from the JSON body (or an If-Match header), call Cancel;
// ErrNotFound -> 404, ErrVersionConflict / policy.ErrIllegalTransition -> 409.
func (h *Handler) cancelPolicy(w http.ResponseWriter, r *http.Request) {
	panic("TODO(Day2): cancelPolicy")
}

// writeJSON is the shared response helper (Spring does this for you via Jackson).
// TODO(Day2): set Content-Type: application/json, WriteHeader(status), json.NewEncoder(w).Encode(v).
func writeJSON(w http.ResponseWriter, status int, v any) {
	panic("TODO(Day2): writeJSON")
}
