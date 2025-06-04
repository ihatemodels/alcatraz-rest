package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	api "github.com/ihatemodels/alcatraz-rest/internal/api/v1"
	"github.com/ihatemodels/alcatraz-rest/internal/config"
	"github.com/ihatemodels/alcatraz-rest/internal/observability"
)

var version string

func main() {
	cfg, err := config.LoadConfig()

	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger := observability.InitLogger(cfg.Observability)

	logger.Info("starting...", "application", "alcatraz-rest", "version", version)

	srv := &http.Server{
		Addr:    cfg.GetServerAddress(),
		Handler: nil,
	}

	http.HandleFunc("/api/ping", api.PingHandler)

	term := make(chan os.Signal, 1)
	srvClose := make(chan struct{})

	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			close(srvClose)
		}
	}()

	for {
		select {
		case <-term:
			// TODO: notify any dependencies that the server is shutting down
			logger.Info("Received SIGTERM, exiting gracefully...")
			os.Exit(0)
		case <-srvClose:
			os.Exit(1)
		}
	}
}
