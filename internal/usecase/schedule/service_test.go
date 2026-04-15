package schedule

import (
	"context"
	"testing"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
)

func TestCreateActiveScheduleMaterializesDefaultWindow(t *testing.T) {
	t.Parallel()

	repo := &stubRepository{
		createResult: &scheduledomain.TaskSchedule{
			ID:          101,
			Title:       "Call patients",
			Description: "Every other day",
			Rule:        scheduledomain.DailyIntervalRule{EveryDays: 2},
			StartDate:   mustUsecaseDate(t, "2026-04-15"),
			Active:      true,
		},
	}
	txManager := &stubTxManager{}
	materializer := &stubWindowMaterializer{}
	location := time.FixedZone("UTC+8", 8*60*60)

	service := NewService(repo, txManager, materializer, location, 30)
	service.now = func() time.Time {
		return time.Date(2026, time.April, 15, 20, 0, 0, 0, time.UTC)
	}

	created, err := service.Create(context.Background(), CreateInput{
		Title:       "  Call patients  ",
		Description: "  Every other day  ",
		StartDate:   mustUsecaseDate(t, "2026-04-15"),
		Rule:        scheduledomain.DailyIntervalRule{EveryDays: 2},
	})
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}

	if !txManager.called {
		t.Fatalf("expected create to run inside transaction")
	}

	if created.ID != 101 {
		t.Fatalf("unexpected created id: got %d want 101", created.ID)
	}

	if len(materializer.materializeCalls) != 1 {
		t.Fatalf("unexpected materialize call count: got %d want 1", len(materializer.materializeCalls))
	}

	call := materializer.materializeCalls[0]
	if call.scheduleID != 101 {
		t.Fatalf("unexpected materialized schedule id: got %d want 101", call.scheduleID)
	}

	assertUsecaseDate(t, call.from, "2026-04-16")
	assertUsecaseDate(t, call.to, "2026-05-16")
}

func TestCreateInactiveScheduleSkipsMaterialization(t *testing.T) {
	t.Parallel()

	active := false
	repo := &stubRepository{
		createResult: &scheduledomain.TaskSchedule{
			ID:          202,
			Title:       "Inventory",
			Description: "Inactive schedule",
			Rule:        scheduledomain.MonthParityRule{Parity: scheduledomain.MonthParityEven},
			StartDate:   mustUsecaseDate(t, "2026-04-15"),
			Active:      false,
		},
	}
	service := NewService(repo, &stubTxManager{}, &stubWindowMaterializer{}, time.UTC, 30)

	if _, err := service.Create(context.Background(), CreateInput{
		Title:       "Inventory",
		Description: "Inactive schedule",
		StartDate:   mustUsecaseDate(t, "2026-04-15"),
		Active:      &active,
		Rule:        scheduledomain.MonthParityRule{Parity: scheduledomain.MonthParityEven},
	}); err != nil {
		t.Fatalf("create inactive schedule: %v", err)
	}

	materializer := service.materialize.(*stubWindowMaterializer)
	if len(materializer.materializeCalls) != 0 {
		t.Fatalf("expected no materialization for inactive schedule")
	}
}

func TestUpdateScheduleRebuildsDefaultWindow(t *testing.T) {
	t.Parallel()

	repo := &stubRepository{
		updateResult: &scheduledomain.TaskSchedule{
			ID:          303,
			Title:       "Monthly report",
			Description: "Updated rule",
			Rule:        scheduledomain.MonthlyDayRule{DayOfMonth: 15},
			StartDate:   mustUsecaseDate(t, "2026-04-01"),
			Active:      true,
		},
	}
	materializer := &stubWindowMaterializer{}
	service := NewService(repo, &stubTxManager{}, materializer, time.UTC, 30)
	service.now = func() time.Time {
		return time.Date(2026, time.April, 15, 9, 0, 0, 0, time.UTC)
	}

	updated, err := service.Update(context.Background(), 303, UpdateInput{
		Title:       "Monthly report",
		Description: "Updated rule",
		StartDate:   mustUsecaseDate(t, "2026-04-01"),
		Rule:        scheduledomain.MonthlyDayRule{DayOfMonth: 15},
	})
	if err != nil {
		t.Fatalf("update schedule: %v", err)
	}

	if updated.ID != 303 {
		t.Fatalf("unexpected updated id: got %d want 303", updated.ID)
	}

	if len(materializer.rebuildCalls) != 1 {
		t.Fatalf("unexpected rebuild call count: got %d want 1", len(materializer.rebuildCalls))
	}

	call := materializer.rebuildCalls[0]
	if call.scheduleID != 303 {
		t.Fatalf("unexpected rebuild schedule id: got %d want 303", call.scheduleID)
	}

	assertUsecaseDate(t, call.from, "2026-04-15")
	assertUsecaseDate(t, call.to, "2026-05-15")
}

type stubRepository struct {
	createResult *scheduledomain.TaskSchedule
	updateResult *scheduledomain.TaskSchedule
}

func (s *stubRepository) Create(ctx context.Context, schedule *scheduledomain.TaskSchedule) (*scheduledomain.TaskSchedule, error) {
	if s.createResult != nil {
		return s.createResult, nil
	}

	return schedule, nil
}

func (s *stubRepository) GetByID(ctx context.Context, id int64) (*scheduledomain.TaskSchedule, error) {
	return nil, nil
}

func (s *stubRepository) List(ctx context.Context) ([]scheduledomain.TaskSchedule, error) {
	return nil, nil
}

func (s *stubRepository) Update(ctx context.Context, schedule *scheduledomain.TaskSchedule) (*scheduledomain.TaskSchedule, error) {
	if s.updateResult != nil {
		return s.updateResult, nil
	}

	return schedule, nil
}

func (s *stubRepository) Deactivate(ctx context.Context, id int64, updatedAt time.Time) error {
	return nil
}

func (s *stubRepository) ListActiveBetween(ctx context.Context, from time.Time, to time.Time) ([]scheduledomain.TaskSchedule, error) {
	return nil, nil
}

type stubTxManager struct {
	called bool
}

func (s *stubTxManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	s.called = true
	return fn(ctx)
}

type materializeCall struct {
	scheduleID int64
	from       time.Time
	to         time.Time
}

type stubWindowMaterializer struct {
	materializeCalls []materializeCall
	rebuildCalls     []materializeCall
}

func (s *stubWindowMaterializer) MaterializeScheduleWindow(ctx context.Context, scheduleID int64, from time.Time, to time.Time) error {
	s.materializeCalls = append(s.materializeCalls, materializeCall{scheduleID: scheduleID, from: from, to: to})
	return nil
}

func (s *stubWindowMaterializer) RebuildScheduleWindow(ctx context.Context, scheduleID int64, from time.Time, to time.Time) error {
	s.rebuildCalls = append(s.rebuildCalls, materializeCall{scheduleID: scheduleID, from: from, to: to})
	return nil
}

func mustUsecaseDate(t *testing.T, raw string) time.Time {
	t.Helper()

	value, err := scheduledomain.ParseDate(raw)
	if err != nil {
		t.Fatalf("parse date %q: %v", raw, err)
	}

	return value
}

func assertUsecaseDate(t *testing.T, got time.Time, expected string) {
	t.Helper()

	if formatted := scheduledomain.FormatDate(got); formatted != expected {
		t.Fatalf("unexpected date: got %s want %s", formatted, expected)
	}
}
