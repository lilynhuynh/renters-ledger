// Command api is the HTTP server entry point (Day 2), the @SpringBootApplication class.
// There is no auto-configuration. main builds every dependency by hand and wires them together.
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/lilynhuynh/renters-ledger/api/internal/httpapi"
	"github.com/lilynhuynh/renters-ledger/api/internal/store"
)

func main() {
	// slog is the standard structured logger, like SLF4J + Logback with a JSON encoder.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// TODO(Day2): pick the store. Start with store.NewMemoryStore().
	st := store.NewMemoryStore()
	h := httpapi.NewHandler(st, logger)
	server := &http.Server{Addr: ":8080", Handler: h.Routes(), ReadHeaderTimeout: 5 * time.Second}
	logger.Info("Server is listening at port 8080")
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Server stopped", "err", err)
		os.Exit(1)
	}

	// TODO(Day3): if DATABASE_URL is set, create a pgxpool.Pool and use store.NewPostgresStore(pool),
	//             like switching Spring @Profile("memory") / @Profile("postgres").
	// TODO(Day2): h := httpapi.NewHandler(st, logger)
	// TODO(Day2): srv := &http.Server{Addr: ":8080", Handler: h.Routes(),
	//             ReadHeaderTimeout: 5 * time.Second}  // always set timeouts in Go
	// TODO(Day2): logger.Info("listening", "addr", srv.Addr); if err := srv.ListenAndServe(); ...
	// TODO(Day6, optional): graceful shutdown with signal.NotifyContext + srv.Shutdown(ctx).

}
