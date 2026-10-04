// Package httpapi translates HTTP to and from domain calls: decode request, call Store, encode JSON.
// It must NOT contain business rules (those live in policy) or SQL (that lives in store).
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/lilynhuynh/renters-ledger/api/internal/policy"
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
//	mux.HandleFunc("GET /healthz", h.health)
//	mux.HandleFunc("GET /policies", h.listPolicies)
//	mux.HandleFunc("GET /policies/{id}", h.getPolicy)
//	mux.HandleFunc("POST /policies", h.createPolicy)
//	mux.HandleFunc("POST /policies/{id}/cancel", h.cancelPolicy)
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /policies", h.listPolicies)
	mux.HandleFunc("GET /policies/{id}", h.getPolicy)
	mux.HandleFunc("POST /policies", h.createPolicy)
	mux.HandleFunc("POST /policies/{id}/cancel", h.cancelPolicy)
	return Logging(h.logger, mux)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := h.store.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error in return list of polcies")
		return
	}
	writeJSON(w, http.StatusOK, policies)
}

func (h *Handler) getPolicy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// Get context
	p, err := h.store.Get(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "policy not found")
		return
	}
	if err != nil {
		h.logger.Error("get policy failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) createPolicy(w http.ResponseWriter, r *http.Request) {
	// Decode the body to get the request struct
	type createPolicyRequest struct {
		CustomerID   string `json:"customer_id"` // like @JsonProperty
		Unit         string `json:"unit"`
		PremiumCents int64  `json:"premium_cents"`
	}
	var request createPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Warn("create policy failed", "err", err)
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	// Validate the body
	if strings.TrimSpace(request.Unit) == "" { //empty unit
		writeError(w, http.StatusBadRequest, "unit is required")
		return
	}
	if request.PremiumCents < 0 { // not >= 0
		writeError(w, http.StatusBadRequest, "premium must not be negative")
		return
	}
	timestamp := time.Now()
	p := policy.Policy{
		CustomerID:    request.CustomerID,
		Unit:          request.Unit,
		PremiumCents:  request.PremiumCents,
		Status:        policy.StatusQuoted,
		EffectiveDate: timestamp,
	}
	newPol, err := h.store.Create(r.Context(), p)
	if err != nil {
		h.logger.Error(fmt.Sprintf("error creating policy %q", newPol), "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, newPol)
}

func (h *Handler) cancelPolicy(w http.ResponseWriter, r *http.Request) {
	type cancelPolicyRequest struct {
		Version int64 `json:"version"`
	}
	var request cancelPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Warn("cancel policy failed", "err", err)
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	id := r.PathValue("id")
	if request.Version == 0 {
		writeError(w, http.StatusBadRequest, "policy was modified, reload and retry")
		return
	}
	p, err := h.store.Cancel(r.Context(), id, request.Version)
	if errors.Is(err, store.ErrVersionConflict) {
		writeError(w, http.StatusConflict, "policy was modified, reload and retry")
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "policy was modified, reload and retry")
		return
	}
	if errors.Is(err, policy.ErrIllegalTransition) {
		writeError(w, http.StatusConflict, "illegal transition")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// writeJSON is the shared response helper (Spring does this for you via Jackson).
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
