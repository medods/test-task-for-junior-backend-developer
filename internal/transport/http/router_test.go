package transporthttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/taskservice/internal/domain/recurrence"
	domaintask "example.com/taskservice/internal/domain/task"
	docs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
	taskusecase "example.com/taskservice/internal/usecase/task"
	"golang.org/x/time/rate"
)

type mockUsecase struct{}

func (m *mockUsecase) Create(ctx context.Context, input taskusecase.CreateInput) (*domaintask.Task, error) {
	now := time.Now().UTC()
	return &domaintask.Task{ID: 1, Title: input.Title, CreatedAt: now, UpdatedAt: now, Status: domaintask.StatusNew, RecurrenceType: input.RecurrenceType, RecurrencePayload: input.RecurrencePayload, RecurrenceStartDate: input.RecurrenceStartDate}, nil
}
func (m *mockUsecase) GetByID(ctx context.Context, id int64) (*domaintask.Task, error) {
	now := time.Now().UTC()
	return &domaintask.Task{ID: id, Title: "t", CreatedAt: now, UpdatedAt: now, Status: domaintask.StatusNew, RecurrenceType: recurrence.NoneRecurrence, RecurrencePayload: recurrence.RecurrencePayload{}}, nil
}
func (m *mockUsecase) Update(ctx context.Context, id int64, input taskusecase.UpdateInput) (*domaintask.Task, error) {
	return m.GetByID(ctx, id)
}
func (m *mockUsecase) Delete(ctx context.Context, id int64) error { return nil }
func (m *mockUsecase) List(ctx context.Context) ([]domaintask.Task, error) {
	now := time.Now().UTC()
	return []domaintask.Task{{ID: 1, Title: "t", CreatedAt: now, UpdatedAt: now, Status: domaintask.StatusNew, RecurrenceType: recurrence.NoneRecurrence, RecurrencePayload: recurrence.RecurrencePayload{}}}, nil
}
func (m *mockUsecase) ListByDate(ctx context.Context, date time.Time) ([]domaintask.Task, error) {
	return m.List(ctx)
}

func TestRateLimitMiddleware_Exceeded(t *testing.T) {
	l := rate.NewLimiter(0, 0)
	mw := rateLimitMiddleware(l)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := mw(next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 got %d", rr.Code)
	}
}

func TestNewRouter_HandleRoutes(t *testing.T) {
	mock := &mockUsecase{}
	taskHandler := httphandlers.NewTaskHandler(mock)
	docsHandler := docs.NewHandler()
	router := NewRouter(taskHandler, docsHandler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger/openapi.json", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("swagger spec expected 200 got %d", rr.Code)
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	router.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("api tasks expected 200 got %d", rr2.Code)
	}
}
