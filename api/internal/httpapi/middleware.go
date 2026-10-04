package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

// Logging wraps a handler and logs method, path, status and duration for each request.
// Middleware in Go is just a function http.Handler -> http.Handler, like a servlet Filter.
func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r) // real handler runs here
		logger.Info(
			"request", "method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start))
	})
}

// statusRecorder remembers the status code a handler wrote. http.ResponseWriter does not
// expose it. Embedding http.ResponseWriter "inherits" all its methods (composition, not
// extends), so you only override WriteHeader.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code // assign it to struct so it can log it
	s.ResponseWriter.WriteHeader(code)
}
