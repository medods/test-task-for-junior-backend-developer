package task

import (
	"context"
	"log/slog"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Scheduler struct {
	repo Repository
	now  func() time.Time
}

func NewScheduler(repo Repository) *Scheduler {
	return &Scheduler{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	for {
		delay := s.durationUntilMidnight()
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}

		if err := s.Tick(ctx); err != nil {
			slog.Error("scheduler tick failed", "error", err)
		}
	}
}

func (s *Scheduler) Tick(ctx context.Context) error {
	today := s.now()

	tasks, err := s.repo.ListDueRecurrences(ctx, today)
	if err != nil {
		return err
	}

	for _, t := range tasks {
		if err := s.resetTask(ctx, t, today); err != nil {
			slog.Error("reset recurring task", "task_id", t.ID, "error", err)
		}
	}

	return nil
}

func (s *Scheduler) resetTask(ctx context.Context, t taskdomain.Task, today time.Time) error {
	rec, _, err := s.repo.GetRecurrence(ctx, *t.RecurrenceID)
	if err != nil {
		return err
	}

	next := rec.NextOccurrence(today)
	if next != nil {
		if err := s.repo.UpdateNextOccurrence(ctx, *t.RecurrenceID, *next); err != nil {
			return err
		}
	}

	t.Status = taskdomain.StatusNew
	t.UpdatedAt = today
	_, err = s.repo.Update(ctx, &t)
	return err
}

func (s *Scheduler) durationUntilMidnight() time.Duration {
	now := s.now()
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return next.Sub(now)
}
