package scheduler

import (
	"context"
	"fmt"
	"time"

	"your-project/internal/domain/task"
	"your-project/internal/logger"
)

type WorkerConfig struct {
	TickInterval time.Duration 
	Lookahead    time.Duration 
	BatchLimit   int          
	InstanceID   int           
	ClusterSize  int         
}

type Worker struct {
	repo task.TaskRepository      
	reg  *task.GeneratorRegistry
	log  logger.Logger           
	cfg  WorkerConfig             
}

func NewWorker(repo task.TaskRepository, reg *task.GeneratorRegistry, log logger.Logger, cfg WorkerConfig) *Worker {
	return &Worker{repo: repo, reg: reg, log: log, cfg: cfg}
}

func (w *Worker) Run(ctx context.Context) error {
	w.log.Info("recurring worker started",
		"tick_interval", w.cfg.TickInterval,
		"lookahead", w.cfg.Lookahead,
		"instance_id", w.cfg.InstanceID,
	)

	ticker := time.NewTicker(w.cfg.TickInterval)
	defer ticker.Stop()

	if err := w.processTick(ctx); err != nil {
		w.log.Error("initial tick failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			w.log.Info("recurring worker graceful shutdown")
			return ctx.Err()
		case <-ticker.C:
			if err := w.processTick(ctx); err != nil {
				w.log.Error("tick failed", "error", err)
			}
		}
	}
}

func (w *Worker) processTick(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	now := time.Now().UTC()
	windowEnd := now.Add(w.cfg.Lookahead)

	templates, err := w.repo.GetActiveTemplatesByDate(ctx, now, w.cfg.BatchLimit, w.cfg.InstanceID, w.cfg.ClusterSize)
	if err != nil {
		return fmt.Errorf("fetch templates: %w", err)
	}

	if len(templates) == 0 {
		return nil
	}

	w.log.Info("processing tick", "templates_count", len(templates), "window_end", windowEnd)

	for _, tmpl := range templates {
		if err := w.generateAndInsert(ctx, tmpl, now, windowEnd); err != nil {
			w.log.Error("template generation failed",
				"template_id", tmpl.ID,
				"type", tmpl.RecurringConfig.Type,
				"error", err,
			)
		}
	}

	return nil
}

func (w *Worker) generateAndInsert(ctx context.Context, tmpl *task.Task, windowStart, windowEnd time.Time) error {
	if tmpl.RecurringConfig == nil {
		return nil 
	}

	gen, err := w.reg.Get(tmpl.RecurringConfig.Type)
	if err != nil {
		return fmt.Errorf("get generator: %w", err)
	}

	allDates, err := gen.Generate(*tmpl.RecurringConfig)
	if err != nil {
		return fmt.Errorf("generate dates: %w", err)
	}

	var windowDates []time.Time
	for _, d := range allDates {
		if !d.Before(windowStart) && !d.After(windowEnd) {
			windowDates = append(windowDates, d)
		}
	}

	if len(windowDates) == 0 {
		return nil
	}

	inserted, err := w.repo.InsertInstancesIdempotent(ctx, tmpl.ID, windowDates)
	if err != nil {
		return fmt.Errorf("insert instances: %w", err)
	}

	if inserted > 0 {
		w.log.Info("instances created",
			"template_id", tmpl.ID,
			"count", inserted,
			"window_start", windowStart,
			"window_end", windowEnd,
		)
	}

	return nil
}