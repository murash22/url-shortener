package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"url-shortener/internal/config"
	http_server "url-shortener/internal/http-server"
	"url-shortener/internal/storage/sqlite"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("Starting URL Shortener", slog.String("env", cfg.Env), slog.String("addr", cfg.Addr))
	log.Debug("debug messages are enabled")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", "err", err)
		os.Exit(1)
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	srv := http_server.New(log, cfg, storage)
	go func() {
		if err = srv.Run(); err != nil {
			log.Error("failed to start server", "err", err)
			os.Exit(1)
		}
	}()
	<-s
	log.Info("shutting down the server...")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level:     slog.LevelDebug,
				AddSource: true,
			}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
