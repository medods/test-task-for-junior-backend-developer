package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	TickerRetryConnect = 2 * time.Second
	AttemptRetryCount  = 5
)

func Open(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database dsn is empty")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	var pool *pgxpool.Pool
	for i := 0; i < AttemptRetryCount; i++ {
		pool, err = pgxpool.NewWithConfig(ctx, cfg)
		if err == nil {
			err = pool.Ping(ctx)
		}

		if err == nil {
			return pool, nil
		}

		if pool != nil {
			pool.Close()
		}

		select {
		case <-time.After(TickerRetryConnect):
			continue
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("failed to connect to postgres after retries: %w", err)
}
