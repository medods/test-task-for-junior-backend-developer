package worker

import (
    "context"
    "log"
    "time"

    "example.com/taskservice/internal/usecase/task"
)

type PeriodicWorker struct {
    usecase  *task.ProcessRecurringTasksUseCase
    interval time.Duration
}

func NewPeriodicWorker(uc *task.ProcessRecurringTasksUseCase, interval time.Duration) *PeriodicWorker {
    return &PeriodicWorker{
        usecase:  uc,
        interval: interval,
    }
}

func (w *PeriodicWorker) Start(ctx context.Context) {
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            log.Println("periodic worker stopped")
            return
        case <-ticker.C:
            if err := w.usecase.Execute(ctx); err != nil {
                log.Printf("periodic worker error: %v", err)
            }
        }
    }
}