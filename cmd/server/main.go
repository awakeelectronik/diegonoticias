package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/awakeelectronik/diegonoticias/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config inválida", "error", err)
		os.Exit(2)
	}

	logger := newLogger(cfg)
	slog.SetDefault(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<h1>Diego Noticias</h1>"))
	})

	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	slog.Info("servidor iniciado", "addr", cfg.Listen, "env", cfg.Env)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("servidor falló", "error", err)
		os.Exit(1)
	}
	slog.Info("servidor detenido")
}

func newLogger(cfg config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	if cfg.Env == "production" {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

