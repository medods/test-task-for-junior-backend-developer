package main

import (
	"context"
	"fmt"
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
	recurrenceusecase "example.com/taskservice/internal/usecase/recurrence"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	taskRepo := postgresrepo.New(pool)
	recurrenceRepo := postgresrepo.NewRecurrenceRepository(pool)
	taskService := taskusecase.NewService(taskRepo)
	recurrenceService := recurrenceusecase.NewService(recurrenceRepo)

	if err := recurrenceService.MaterializeActive(ctx); err != nil {
		logger.Error("initial materialization failed", "error", err)
		os.Exit(1)
	}

	go runMaterializer(ctx, logger, recurrenceService, cfg.MaterializationInterval)

	taskHandler := httphandlers.NewTaskHandler(taskService)
	recurrenceHandler := httphandlers.NewRecurrenceHandler(recurrenceService)
	docsHandler := swaggerdocs.NewHandler()
	router := transporthttp.NewRouter(taskHandler, recurrenceHandler, docsHandler)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown http server", "error", err)
		}
	}()

	logger.Info("http server started", "addr", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("listen and serve", "error", err)
		os.Exit(1)
	}
}

type materializer interface {
	MaterializeActive(ctx context.Context) error
}

func runMaterializer(ctx context.Context, logger *slog.Logger, service materializer, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := service.MaterializeActive(ctx); err != nil {
				logger.Error("periodic materialization failed", "error", err)
			}
		}
	}
}

type config struct {
	HTTPAddr                string
	DatabaseDSN             string
	MaterializationInterval time.Duration
}

func loadConfig() config {
	cfg := config{
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseDSN: envOrDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"),
	}

	if cfg.DatabaseDSN == "" {
		panic(fmt.Errorf("DATABASE_DSN is required"))
	}

	interval, err := time.ParseDuration(envOrDefault("MATERIALIZATION_INTERVAL", "1m"))
	if err != nil {
		panic(fmt.Errorf("invalid MATERIALIZATION_INTERVAL: %w", err))
	}
	cfg.MaterializationInterval = interval

	return cfg
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
