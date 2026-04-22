package main

import (
	"context"
	"errors"
	"example.com/taskservice/internal/scheduler"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	postgresrepo "example.com/taskservice/internal/repository/postgres"
	transporthttp "example.com/taskservice/internal/transport/http"
	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
	"example.com/taskservice/internal/usecase/task"
)

const (
	InternalErrorCode      = 1
	ShotDownContextTimeOut = 10 * time.Second
	ReadHeaderTimeout      = 5 * time.Second
	SchedulerInterval      = 1 * time.Minute
)

// @title Task Service API
// @version 1.0
// @description API для управления задачами и шаблонами.
// @host localhost:8080
// @BasePath /
func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("Error loading config", "error", err)
		os.Exit(InternalErrorCode)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(InternalErrorCode)
	}
	defer pool.Close()

	taskRepo := postgresrepo.New(pool)
	taskUsecase := task.NewService(taskRepo, logger)

	taskScheduler := scheduler.NewScheduler(taskUsecase, logger, SchedulerInterval)

	go taskScheduler.Run(ctx)

	taskHandler := httphandlers.NewTaskHandler(taskUsecase)
	docsHandler := swaggerdocs.NewHandler()
	router := transporthttp.NewRouter(taskHandler, docsHandler)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: ReadHeaderTimeout,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), ShotDownContextTimeOut)
		defer cancel()

		if err = server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown http server", "error", err)
		}
	}()

	logger.Info("http server started", "addr", cfg.HTTPAddr)

	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("listen and serve", "error", err)
		os.Exit(InternalErrorCode)
	}
}

type config struct {
	HTTPAddr    string
	DatabaseDSN string
}

func loadConfig() (*config, error) {
	cfg := &config{
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseDSN: envOrDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"),
	}

	if cfg.DatabaseDSN == "" {
		return nil, errors.New("DATABASE_DSN is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
