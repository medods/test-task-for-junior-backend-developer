package app

import (
	"fmt"
	"os"
	"strconv"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
)

const defaultDatabaseDSN = "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"

type Config struct {
	HTTPAddr                  string
	DatabaseDSN               string
	AppTimezone               string
	Location                  *time.Location
	MaterializationWindowDays int
	SchedulerCron             string
}

func LoadConfig() (Config, error) {
	locationName := envOrDefault("APP_TIMEZONE", "UTC")
	location, err := time.LoadLocation(locationName)
	if err != nil {
		return Config{}, fmt.Errorf("load APP_TIMEZONE: %w", err)
	}

	windowDays, err := intEnvOrDefault("MATERIALIZATION_WINDOW_DAYS", 30)
	if err != nil {
		return Config{}, fmt.Errorf("parse MATERIALIZATION_WINDOW_DAYS: %w", err)
	}

	if windowDays <= 0 {
		return Config{}, fmt.Errorf("MATERIALIZATION_WINDOW_DAYS must be greater than 0")
	}

	cfg := Config{
		HTTPAddr:                  envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseDSN:               envOrDefault("DATABASE_DSN", defaultDatabaseDSN),
		AppTimezone:               locationName,
		Location:                  location,
		MaterializationWindowDays: windowDays,
		SchedulerCron:             envOrDefault("SCHEDULER_CRON", "5 0 * * *"),
	}

	if cfg.DatabaseDSN == "" {
		return Config{}, fmt.Errorf("DATABASE_DSN is required")
	}

	return cfg, nil
}

func (c Config) MaterializationWindow(now time.Time) (time.Time, time.Time) {
	from := scheduledomain.DateFromLocation(now, c.Location)
	to := from.AddDate(0, 0, c.MaterializationWindowDays)
	return from, to
}

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func intEnvOrDefault(key string, fallback int) (int, error) {
	raw := envOrDefault(key, strconv.Itoa(fallback))
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}

	return value, nil
}
