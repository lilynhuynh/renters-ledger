package httpapi

// httptest is the standard-library MockMvc: build a request, record the response, and
// assert on it without opening a real port. Every test gets a fresh MemoryStore, so tests
// never see each other's data (like a fresh Spring context per test, but instant).

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lilynhuynh/renters-ledger/api/internal/policy"
	"github.com/lilynhuynh/renters-ledger/api/internal/store"
)

// ---------- helpers ----------

func newTestHandler() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // silence logs in tests
	return NewHandler(store.NewMemoryStore(), logger).Routes()
}

// do sends one request through the handler and returns the recorded response.
// Pass body "" for requests without a body. Like mockMvc.perform(...).andReturn().
func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// decode parses the response body as JSON into a T and fails the test if it can't.
// [T any] makes this a generic function, like `<T> T decode(...)` in Java.
// Usage: p := decode[policy.Policy](t, rec)
func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("response body is not the expected JSON shape (%T): %v\nbody: %s", v, err, rec.Body.String())
	}
	return v
}

// wantStatus fails the test (and stops it) if the status code is wrong.
func wantStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d\nbody: %s", rec.Code, want, rec.Body.String())
	}
}

// wantErrorBody checks that an error response looks like {"error": "<something>"}.
func wantErrorBody(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	body := decode[map[string]string](t, rec)
	if body["error"] == "" {
		t.Errorf(`error response should be {"error": "..."}, got: %s`, rec.Body.String())
	}
}

const validCreateBody = `{"customer_id":"cus_1","unit":"1A","premium_cents":1250}`

// mustCreatePolicy creates a policy through the API and returns what the API sent back.
// Many tests need an existing policy first, so this is shared setup (like a @BeforeEach helper).
func mustCreatePolicy(t *testing.T, h http.Handler) policy.Policy {
	t.Helper()
	rec := do(t, h, http.MethodPost, "/policies", validCreateBody)
	wantStatus(t, rec, http.StatusCreated)
	p := decode[policy.Policy](t, rec)
	if p.ID == "" {
		t.Fatalf("POST /policies should respond with the created policy, including its new ID.\nbody: %s",
			rec.Body.String())
	}
	return p
}

// ---------- tests ----------

func TestRoutes(t *testing.T) {
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
		{"unknown path", http.MethodGet, "/nope", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, newTestHandler(), tt.method, tt.path, "")
			wantStatus(t, rec, tt.wantStatus)
		})
	}
}

func TestHealth(t *testing.T) {
	rec := do(t, newTestHandler(), http.MethodGet, "/healthz", "")
	wantStatus(t, rec, http.StatusOK)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if body := decode[map[string]string](t, rec); body["status"] != "ok" {
		t.Errorf(`body = %s, want {"status":"ok"}`, rec.Body.String())
	}
}

func TestCreatePolicy(t *testing.T) {
	t.Run("valid request returns 201 and the created policy", func(t *testing.T) {
		p := mustCreatePolicy(t, newTestHandler())

		if p.CustomerID != "cus_1" || p.Unit != "1A" || p.PremiumCents != 1250 {
			t.Errorf("fields not copied from request: %+v", p)
		}
		if p.Status != policy.StatusQuoted {
			t.Errorf("Status = %q, want %q (new policies start as quoted)", p.Status, policy.StatusQuoted)
		}
		if p.Version != 1 {
			t.Errorf("Version = %d, want 1", p.Version)
		}
		if p.CreatedAt.IsZero() {
			t.Errorf("CreatedAt is zero; the store should set it")
		}
	})

	// Every one of these must be rejected with 400 and a JSON error body.
	badRequests := []struct {
		name string
		body string
	}{
		{"malformed JSON", `{"unit":`},
		{"missing unit", `{"customer_id":"cus_1","premium_cents":1250}`},
		{"blank unit", `{"customer_id":"cus_1","unit":"   ","premium_cents":1250}`},
		{"negative premium", `{"customer_id":"cus_1","unit":"1A","premium_cents":-5}`},
		{"premium as a string", `{"customer_id":"cus_1","unit":"1A","premium_cents":"12.50"}`},
	}
	for _, tt := range badRequests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, newTestHandler(), http.MethodPost, "/policies", tt.body)
			wantStatus(t, rec, http.StatusBadRequest)
			wantErrorBody(t, rec)
		})
	}
}

func TestGetPolicy(t *testing.T) {
	t.Run("returns a policy that was created", func(t *testing.T) {
		h := newTestHandler()
		created := mustCreatePolicy(t, h)

		rec := do(t, h, http.MethodGet, "/policies/"+created.ID, "")
		wantStatus(t, rec, http.StatusOK)

		got := decode[policy.Policy](t, rec)
		if got.ID != created.ID || got.Unit != created.Unit || got.Version != created.Version {
			t.Errorf("GET returned %+v, want the created policy %+v", got, created)
		}
	})

	t.Run("unknown id returns 404 with an error body", func(t *testing.T) {
		rec := do(t, newTestHandler(), http.MethodGet, "/policies/pol_999", "")
		wantStatus(t, rec, http.StatusNotFound)
		wantErrorBody(t, rec)
	})
}

func TestListPolicies(t *testing.T) {
	t.Run("empty store returns an empty JSON array, not null", func(t *testing.T) {
		rec := do(t, newTestHandler(), http.MethodGet, "/policies", "")
		wantStatus(t, rec, http.StatusOK)

		// The Day 5 front end expects an array. `null` or an object would break it.
		if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
			t.Errorf("body = %s, want []", got)
		}
	})

	t.Run("returns every created policy", func(t *testing.T) {
		h := newTestHandler()
		mustCreatePolicy(t, h)
		mustCreatePolicy(t, h)

		rec := do(t, h, http.MethodGet, "/policies", "")
		wantStatus(t, rec, http.StatusOK)

		list := decode[[]policy.Policy](t, rec)
		if len(list) != 2 {
			t.Errorf("got %d policies, want 2: %+v", len(list), list)
		}
	})
}

func TestCancelPolicy(t *testing.T) {
	// Each case starts from one freshly created policy (Version 1).
	// path "" means "use the created policy's /policies/{id}/cancel".
	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
	}{
		{"current version cancels", "", `{"version":1}`, http.StatusOK},
		{"stale version conflicts", "", `{"version":99}`, http.StatusConflict},
		{"missing version is a bad request", "", `{}`, http.StatusBadRequest},
		{"malformed JSON is a bad request", "", `{"version":`, http.StatusBadRequest},
		{"unknown id is not found", "/policies/pol_999/cancel", `{"version":1}`, http.StatusNotFound},
		// TODO(Day4): cancelling an already-cancelled policy should be 409 once Transition is implemented.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler()
			created := mustCreatePolicy(t, h)

			path := tt.path
			if path == "" {
				path = "/policies/" + created.ID + "/cancel"
			}

			rec := do(t, h, http.MethodPost, path, tt.body)
			wantStatus(t, rec, tt.wantStatus)

			if tt.wantStatus != http.StatusOK {
				wantErrorBody(t, rec)
				return
			}

			// Success: the response is the updated policy...
			got := decode[policy.Policy](t, rec)
			if got.Status != policy.StatusCancelled {
				t.Errorf("Status = %q, want %q", got.Status, policy.StatusCancelled)
			}
			if got.Version != created.Version+1 {
				t.Errorf("Version = %d, want %d (cancel must bump the version)", got.Version, created.Version+1)
			}

			// ...and the change was actually saved: a fresh GET sees it too.
			saved := decode[policy.Policy](t, do(t, h, http.MethodGet, "/policies/"+created.ID, ""))
			if saved.Status != policy.StatusCancelled {
				t.Errorf("after cancel, GET shows Status = %q; was the policy written back to the store?", saved.Status)
			}
		})
	}
}
