package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lilynhuynh/renters-ledger/api/internal/httpapi"
	"github.com/lilynhuynh/renters-ledger/api/internal/store"
)

func main() {
	// slog is the standard structured logger, like SLF4J + Logback with a JSON encoder.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dbUrl := os.Getenv("DATABASE_URL")
	var st store.Store
	if dbUrl == "" {
		logger.Info("using memory store")
		st = store.NewMemoryStore() // set as memory
	} else {
		logger.Info("using postgres store")
		ctx := context.Background()
		pool, err := pgxpool.New(ctx, dbUrl)
		if err != nil {
			logger.Error("database invalid, server stopped", "err", err)
			os.Exit(1)
		}
		if err := pool.Ping(ctx); err != nil { // Check pool status
			logger.Error("database errored, server stopped", "err", err)
			os.Exit(1)
		}
		st = store.NewPostgresStore(pool)
		defer pool.Close()
	}
	h := httpapi.NewHandler(st, logger)
	server := &http.Server{Addr: ":8080", Handler: h.Routes(), ReadHeaderTimeout: 5 * time.Second}
	logger.Info("server is listening at port 8080")
	if err := server.ListenAndServe(); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}

	// TODO(Day6, optional): graceful shutdown with signal.NotifyContext + srv.Shutdown(ctx).

}
