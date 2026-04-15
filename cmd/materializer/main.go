package main

import (
	"context"
	"log/slog"
	"os"

	"example.com/taskservice/internal/app"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := app.LoadConfig()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	container, err := app.NewContainer(context.Background(), cfg)
	if err != nil {
		logger.Error("bootstrap application", "error", err)
		os.Exit(1)
	}
	defer container.Pool.Close()

	if err := container.Materializer.MaterializeDefaultWindow(context.Background()); err != nil {
		logger.Error("materialize task schedules", "error", err)
		os.Exit(1)
	}

	logger.Info("task schedules materialized")
}
