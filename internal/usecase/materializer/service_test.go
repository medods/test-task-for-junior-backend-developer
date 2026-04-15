package materializer

import (
	"context"
	"testing"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	taskdomain "example.com/taskservice/internal/domain/task"
)

func TestRebuildScheduleWindowDeletesFutureNewAndRegeneratesTasks(t *testing.T) {
	t.Parallel()

	scheduleID := int64(42)
	schedule := scheduledomain.TaskSchedule{
		ID:          scheduleID,
		Title:       "Call patients",
		Description: "Every other day",
		Rule:        scheduledomain.DailyIntervalRule{EveryDays: 2},
		StartDate:   mustScheduleDate(t, "2026-04-15"),
		Active:      true,
	}

	scheduleRepo := &stubScheduleRepository{schedule: schedule}
	taskRepo := &stubTaskRepository{}
	service := NewService(scheduleRepo, taskRepo, time.UTC, 30)
	service.now = func() time.Time {
		return time.Date(2026, time.April, 15, 9, 0, 0, 0, time.UTC)
	}

	from := mustScheduleDate(t, "2026-04-15")
	to := mustScheduleDate(t, "2026-04-20")
	if err := service.RebuildScheduleWindow(context.Background(), scheduleID, from, to); err != nil {
		t.Fatalf("rebuild schedule window: %v", err)
	}

	if !taskRepo.deleted {
		t.Fatalf("expected generated tasks cleanup before rebuild")
	}

	if len(taskRepo.created) != 3 {
		t.Fatalf("unexpected created tasks count: got %d want 3", len(taskRepo.created))
	}

	assertTaskDate(t, taskRepo.created[0], "2026-04-15")
	assertTaskDate(t, taskRepo.created[1], "2026-04-17")
	assertTaskDate(t, taskRepo.created[2], "2026-04-19")
}

type stubScheduleRepository struct {
	schedule scheduledomain.TaskSchedule
}

func (s *stubScheduleRepository) GetByID(ctx context.Context, id int64) (*scheduledomain.TaskSchedule, error) {
	return &s.schedule, nil
}

func (s *stubScheduleRepository) ListActiveBetween(ctx context.Context, from time.Time, to time.Time) ([]scheduledomain.TaskSchedule, error) {
	return []scheduledomain.TaskSchedule{s.schedule}, nil
}

type stubTaskRepository struct {
	created []*taskdomain.Task
	deleted bool
}

func (s *stubTaskRepository) CreateGenerated(ctx context.Context, task *taskdomain.Task) error {
	cloned := *task
	s.created = append(s.created, &cloned)
	return nil
}

func (s *stubTaskRepository) DeleteFutureNewGenerated(ctx context.Context, scheduleID int64, from time.Time, to time.Time) error {
	s.deleted = true
	return nil
}

func mustScheduleDate(t *testing.T, raw string) time.Time {
	t.Helper()

	value, err := scheduledomain.ParseDate(raw)
	if err != nil {
		t.Fatalf("parse date %q: %v", raw, err)
	}

	return value
}

func assertTaskDate(t *testing.T, task *taskdomain.Task, expected string) {
	t.Helper()

	if task.ScheduledFor == nil {
		t.Fatalf("task %+v has no scheduled_for", task)
	}

	if got := scheduledomain.FormatDate(*task.ScheduledFor); got != expected {
		t.Fatalf("unexpected scheduled_for: got %s want %s", got, expected)
	}
}
