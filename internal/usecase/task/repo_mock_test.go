package task

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"example.com/taskservice/internal/domain/common"
	"example.com/taskservice/internal/domain/recurrence"
	domaintask "example.com/taskservice/internal/domain/task"
)

type memRepo struct {
	mu    sync.Mutex
	next  int64
	items map[int64]domaintask.Task
}

func newMemRepo() *memRepo {
	return &memRepo{next: 1, items: make(map[int64]domaintask.Task)}
}

func (r *memRepo) Create(ctx context.Context, t *domaintask.Task) (*domaintask.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.next
	r.next++
	t.ID = id
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	t.UpdatedAt = t.CreatedAt
	r.items[id] = *t
	v := r.items[id]
	return &v, nil
}

func (r *memRepo) GetByID(ctx context.Context, id int64) (*domaintask.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if v, ok := r.items[id]; ok {
		vv := v
		return &vv, nil
	}
	return nil, domaintask.ErrNotFound
}

func (r *memRepo) Update(ctx context.Context, t *domaintask.Task) (*domaintask.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[t.ID]; !ok {
		return nil, domaintask.ErrNotFound
	}
	t.UpdatedAt = time.Now().UTC()
	r.items[t.ID] = *t
	v := r.items[t.ID]
	return &v, nil
}

func (r *memRepo) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return domaintask.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *memRepo) List(ctx context.Context) ([]domaintask.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domaintask.Task, 0, len(r.items))
	for _, v := range r.items {
		out = append(out, v)
	}
	return out, nil
}

func (r *memRepo) ListByDate(ctx context.Context, date time.Time) ([]domaintask.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domaintask.Task, 0)
	for _, v := range r.items {
		if v.ActiveOn(date) {
			out = append(out, v)
		}
	}
	return out, nil
}

func TestService_CRUD_and_ListByDate(t *testing.T) {
	repo := newMemRepo()
	s := NewService(repo)
	ctx := context.Background()

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	start := common.NewDate(base)

	created, err := s.Create(ctx, CreateInput{
		Title:               "one-time",
		Description:         "d",
		Status:              domaintask.StatusNew,
		RecurrenceType:      recurrence.NoneRecurrence,
		RecurrencePayload:   recurrence.RecurrencePayload{},
		RecurrenceStartDate: &start,
	})
	if err != nil {
		t.Fatalf("create err: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected id set")
	}

	got, err := s.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get err: %v", err)
	}
	if got.Title != "one-time" {
		t.Fatalf("unexpected title: %v", got.Title)
	}

	upd := UpdateInput{
		Title:               "one-time-upd",
		Description:         "d2",
		Status:              domaintask.StatusNew,
		RecurrenceType:      recurrence.NoneRecurrence,
		RecurrencePayload:   recurrence.RecurrencePayload{},
		RecurrenceStartDate: &start,
	}
	updated, err := s.Update(ctx, created.ID, upd)
	if err != nil {
		t.Fatalf("update err: %v", err)
	}
	if updated.Title != "one-time-upd" {
		t.Fatalf("update not applied")
	}

	// Use Update (full replacement) instead of Patch
	upd2 := UpdateInput{
		Title:               "patched",
		Description:         created.Description,
		Status:              created.Status,
		RecurrenceType:      created.RecurrenceType,
		RecurrencePayload:   created.RecurrencePayload,
		RecurrenceStartDate: created.RecurrenceStartDate,
	}
	patched, err := s.Update(ctx, created.ID, upd2)
	if err != nil {
		t.Fatalf("update err: %v", err)
	}
	if patched.Title != "patched" {
		t.Fatalf("update not applied")
	}

	if err := s.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete err: %v", err)
	}
	if _, err := s.GetByID(ctx, created.ID); !errors.Is(err, domaintask.ErrNotFound) {
		t.Fatalf("expected not found after delete")
	}

	t1, _ := s.Create(ctx, CreateInput{Title: "daily2", Status: domaintask.StatusNew, RecurrenceType: recurrence.DailyEveryN, RecurrencePayload: recurrence.RecurrencePayload{IntervalDays: 2}, RecurrenceStartDate: &start})
	t2, _ := s.Create(ctx, CreateInput{Title: "specific", Status: domaintask.StatusNew, RecurrenceType: recurrence.SpecificDates, RecurrencePayload: recurrence.RecurrencePayload{SpecificDates: []common.DateType{common.NewDate(base.Add(48 * time.Hour))}}, RecurrenceStartDate: &start})
	_ = t1
	_ = t2

	list0, err := s.ListByDate(ctx, base)
	if err != nil {
		t.Fatalf("listbydate err: %v", err)
	}
	if len(list0) == 0 {
		t.Fatalf("expected tasks on base date")
	}
}
