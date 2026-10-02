package httpapi

import (
	"log/slog"
	"net/http"
)

// Logging wraps a handler and logs method, path, status and duration for each request.
// Middleware in Go is just a function http.Handler -> http.Handler, like a servlet Filter.
//
// TODO(Day2): return http.HandlerFunc(func(w, r) { start := time.Now(); wrap w in a
// statusRecorder; next.ServeHTTP(rec, r); logger.Info("request", "method", r.Method,
// "path", r.URL.Path, "status", rec.status, "duration", time.Since(start)) }).
// For now it passes requests straight through.
func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return next
}

// statusRecorder remembers the status code a handler wrote. http.ResponseWriter does not
// expose it. Embedding http.ResponseWriter "inherits" all its methods (composition, not
// extends), so you only override WriteHeader.
//
// TODO(Day2): add `func (s *statusRecorder) WriteHeader(code int)` that saves code and then
// calls s.ResponseWriter.WriteHeader(code). Default status to 200 when you create it.
type statusRecorder struct {
	http.ResponseWriter
	status int
}
