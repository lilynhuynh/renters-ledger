package httpapi

// httptest is the standard-library MockMvc: build a request, record the response, and
// assert on it without opening a real port.

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lilynhuynh/renters-ledger/api/internal/store"
)

func newTestHandler() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // silence logs in tests
	return NewHandler(store.NewMemoryStore(), logger).Routes()
}

func TestRoutes(t *testing.T) {
	t.Skip("Day2: remove this line once Routes and MemoryStore are implemented")

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"health", http.MethodGet, "/healthz", http.StatusOK},
		{"list empty", http.MethodGet, "/policies", http.StatusOK},
		{"get missing", http.MethodGet, "/policies/does-not-exist", http.StatusNotFound},
		{"wrong method", http.MethodDelete, "/policies", http.StatusMethodNotAllowed},
		// TODO(Day2): add create (201), create with bad body (400), get after create (200),
		// cancel (200), cancel with stale version (409). For POSTs use strings.NewReader(`{...}`).
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
