package handlers

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    "github.com/gorilla/mux"

    taskdomain "example.com/taskservice/internal/domain/task"
    taskusecase "example.com/taskservice/internal/usecase/task"
    "example.com/taskservice/internal/domain/recurrence"
)

type mockUsecase struct{}

func (m *mockUsecase) Create(ctx context.Context, input taskusecase.CreateInput) (*taskdomain.Task, error) {
    now := time.Now().UTC()
    return &taskdomain.Task{
        ID:                  1,
        Title:               input.Title,
        Description:         input.Description,
        Status:              input.Status,
        CreatedAt:           now,
        UpdatedAt:           now,
        RecurrenceType:      input.RecurrenceType,
        RecurrencePayload:   input.RecurrencePayload,
        RecurrenceStartDate: input.RecurrenceStartDate,
    }, nil
}

func (m *mockUsecase) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
    now := time.Now().UTC()
    return &taskdomain.Task{ID: id, Title: "t1", CreatedAt: now, UpdatedAt: now, Status: taskdomain.StatusNew, RecurrenceType: recurrence.NoneRecurrence, RecurrencePayload: recurrence.RecurrencePayload{}}, nil
}

func (m *mockUsecase) Update(ctx context.Context, id int64, input taskusecase.UpdateInput) (*taskdomain.Task, error) { return m.GetByID(ctx, id) }
func (m *mockUsecase) Patch(ctx context.Context, id int64, input taskusecase.PatchInput) (*taskdomain.Task, error) { return m.GetByID(ctx, id) }
func (m *mockUsecase) Delete(ctx context.Context, id int64) error { return nil }
func (m *mockUsecase) List(ctx context.Context) ([]taskdomain.Task, error) {
    now := time.Now().UTC()
    return []taskdomain.Task{{ID: 1, Title: "t1", CreatedAt: now, UpdatedAt: now, Status: taskdomain.StatusNew, RecurrenceType: recurrence.NoneRecurrence, RecurrencePayload: recurrence.RecurrencePayload{}}}, nil
}
func (m *mockUsecase) ListByDate(ctx context.Context, date time.Time) ([]taskdomain.Task, error) { return m.List(ctx) }

func TestCreateHandler(t *testing.T) {
    h := NewTaskHandler(&mockUsecase{})

    body := map[string]any{"title": "hello", "description": "d", "status": "new", "recurrence_type": "none", "recurrence_payload": map[string]any{}}
    b, _ := json.Marshal(body)

    req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()

    h.Create(rr, req)

    if rr.Code != http.StatusCreated {
        t.Fatalf("expected status 201 got %d body=%s", rr.Code, rr.Body.String())
    }

    var resp map[string]any
    if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
        t.Fatalf("invalid json: %v", err)
    }
    if resp["title"] != "hello" {
        t.Fatalf("unexpected title: %v", resp["title"]) }
}

func TestGetListHandlers(t *testing.T) {
    h := NewTaskHandler(&mockUsecase{})

    req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
    rr := httptest.NewRecorder()
    router := mux.NewRouter()
    router.HandleFunc("/tasks/{id}", h.GetByID).Methods("GET")
    router.ServeHTTP(rr, req)
    if rr.Code != http.StatusOK { t.Fatalf("GetByID status %d body=%s", rr.Code, rr.Body.String()) }

    req2 := httptest.NewRequest(http.MethodGet, "/tasks", nil)
    rr2 := httptest.NewRecorder()
    h.List(rr2, req2)
    if rr2.Code != http.StatusOK { t.Fatalf("List status %d", rr2.Code) }
}

func TestPatchHandlerNullStartDate(t *testing.T) {
    h := NewTaskHandler(&mockUsecase{})

    payload := map[string]any{"title": "x", "recurrence_start_date": nil}
    b, _ := json.Marshal(payload)
    req := httptest.NewRequest(http.MethodPatch, "/tasks/1", bytes.NewReader(b))
    rr := httptest.NewRecorder()
    router := mux.NewRouter()
    router.HandleFunc("/tasks/{id}", h.Patch).Methods("PATCH")
    router.ServeHTTP(rr, req)
    req.Header.Set("Content-Type", "application/json")
    if rr.Code != http.StatusOK { t.Fatalf("Patch status %d body=%s", rr.Code, rr.Body.String()) }
}
