package scheduler

import (
	"context"
	"log/slog"
	"time"
)

type TaskProcessor interface {
	ProcessTemplates(ctx context.Context)
}
type Scheduler struct {
	processor TaskProcessor
	logger    *slog.Logger
	period    time.Duration
}

func NewScheduler(processor TaskProcessor, logger *slog.Logger, period time.Duration) *Scheduler {
	return &Scheduler{
		processor: processor,
		logger:    logger,
		period:    period,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.period)
	defer ticker.Stop()

	s.logger.Info("scheduler started", "period", s.period)

	s.processor.ProcessTemplates(ctx)

	for {
		select {
		case <-ticker.C:
			s.processor.ProcessTemplates(ctx)
		case <-ctx.Done():
			s.logger.Info("scheduler stopped")
			return
		}
	}
}
