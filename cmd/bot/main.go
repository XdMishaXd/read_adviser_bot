package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	tgClient "read_adviser_bot/internal/clients/telegram"
	"read_adviser_bot/internal/config"
	event_consumer "read_adviser_bot/internal/consumer/event-consumer"
	telegram "read_adviser_bot/internal/events/telegram"
	"read_adviser_bot/internal/lib/logger/sl"
	postgresql "read_adviser_bot/internal/storage/PostgreSQL"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting telegram bot", slog.String("env", cfg.Env))

	tgClient := tgClient.New(cfg.Host, cfg.BotToken)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo, err := postgresql.Connect(ctx, cfg, log)
	if err != nil {
		log.Error("failed to connect to database", sl.Err(err))
		os.Exit(1)
	}
	defer repo.Close()

	eventsProcessor := telegram.New(&tgClient, repo)

	log.Info("service is running")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, cfg.BatchSize)

	if err := consumer.Start(log); err != nil {
		os.Exit(1)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
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
