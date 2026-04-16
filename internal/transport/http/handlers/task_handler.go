// Package handlers contains HTTP handlers for the API.
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	"example.com/taskservice/internal/domain/common"
	task "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
	"example.com/taskservice/internal/domain/recurrence"
)

// TaskHandler handles task HTTP requests.
type TaskHandler struct {
	usecase taskusecase.Usecase
}

var requestValidator = validator.New()

// NewTaskHandler creates a TaskHandler.
func NewTaskHandler(usecase taskusecase.Usecase) *TaskHandler {
	return &TaskHandler{usecase: usecase}
}

// Create handles POST /tasks.
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := requestValidator.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), taskusecase.CreateInput{
		Title:               req.Title,
		Description:         req.Description,
		Status:              req.Status,
		RecurrenceType:      req.RecurrenceType,
		RecurrencePayload:   req.RecurrencePayload,
		RecurrenceStartDate: req.RecurrenceStartDate,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newTaskDTO(created))
}

// GetByID handles GET /tasks/{id}.
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskDTO(task))
}

// Update handles PUT /tasks/{id}.
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	// Read update fields from query parameters (RequestParam style).
	q := r.URL.Query()
	title := q.Get("title")
	description := q.Get("description")
	statusStr := q.Get("status")
	rtStr := q.Get("recurrence_type")
	rpStr := q.Get("recurrence_payload")
	rsd := q.Get("recurrence_start_date")

	if title == "" || statusStr == "" || rtStr == "" || rpStr == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("missing required update parameters: title,status,recurrence_type,recurrence_payload"))
		return
	}

	status := task.Status(statusStr)
	if !status.Valid() {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid status"))
		return
	}

	rt := recurrence.RecurrenceType(rtStr)

	// We accept recurrence_payload as JSON in a query param.
	var rpValue recurrence.RecurrencePayload
	if err := json.Unmarshal([]byte(rpStr), &rpValue); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid recurrence_payload: %w", err))
		return
	}

	var startDate *common.DateType
	if rsd != "" {
		parsed, err := time.Parse(common.DateLayout, rsd)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid recurrence_start_date, expected YYYY-MM-DD"))
			return
		}
		d := common.NewDate(parsed.UTC())
		startDate = &d
	}

	updated, err := h.usecase.Update(r.Context(), id, taskusecase.UpdateInput{
		Title:               title,
		Description:         description,
		Status:              status,
		RecurrenceType:      rt,
		RecurrencePayload:   rpValue,
		RecurrenceStartDate: startDate,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskDTO(updated))
}

// Patch handles PATCH /tasks/{id}.
// Patch removed: partial updates are not supported; use PUT /tasks/{id} with RequestParams.

// Delete handles DELETE /tasks/{id}.
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List handles GET /tasks and supports filtering by date via ?date=YYYY-MM-DD.
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	requestDate := r.URL.Query().Get("date")
	tasks := make([]task.Task, 0)
	var err error

	if requestDate != "" {
		parsed, err := time.Parse(common.DateLayout, requestDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid date, expected YYYY-MM-DD"))
			return
		}

		normalized := common.NewDate(parsed.UTC())
		tasks, err = h.usecase.ListByDate(r.Context(), normalized.Time)
		if err != nil {
			writeUsecaseError(w, err)
			return
		}
	} else {
		tasks, err = h.usecase.List(r.Context())
		if err != nil {
			writeUsecaseError(w, err)
			return
		}
	}

	response := make([]taskDTO, 0, len(tasks))
	for i := range tasks {
		response = append(response, newTaskDTO(&tasks[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func getIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing task id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid task id")
	}

	if id <= 0 {
		return 0, errors.New("invalid task id")
	}

	return id, nil
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, task.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}
