package app

import (
	"your-project/internal/domain/task"
	"your-project/internal/scheduler"
	"your-project/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RecurringConfig struct {
	TickInterval time.Duration
	Lookahead    time.Duration
	BatchLimit   int
}

type RecurringDeps struct {
	Registry   *task.GeneratorRegistry
	Worker     *scheduler.Worker
}

func WireRecurringComponents(
	db *pgxpool.Pool,
	log logger.Logger,
	cfg RecurringConfig,
) (RecurringDeps, error) {
	reg := task.NewGeneratorRegistry()
	reg.Register(task.RecurrenceTypeDaily, &task.DailyGenerator{})
	reg.Register(task.RecurrenceTypeMonthly, &task.MonthlyGenerator{})
	reg.Register(task.RecurrenceTypeSpecificDates, &task.DatesGenerator{})
	reg.Register(task.RecurrenceTypeParity, &task.ParityGenerator{})

	worker := scheduler.NewWorker(
		nil, 
		reg,
		log,
		scheduler.WorkerConfig{
			TickInterval: cfg.TickInterval,
			Lookahead:    cfg.Lookahead,
			BatchLimit:   cfg.BatchLimit,
			InstanceID:   0, 
			ClusterSize:  1, 
		},
	)

	return RecurringDeps{
		Registry: reg,
		Worker:   worker,
	}, nil
}