package handlers

import (
	"encoding/json"
	"errors"
	"example.com/taskservice/internal/domain"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	taskusecase "example.com/taskservice/internal/usecase/task"
)

type TaskHandler struct {
	Service taskusecase.ServiceAPI
}

var ErrNotFound = errors.New("task not found")

func NewTaskHandler(service taskusecase.ServiceAPI) *TaskHandler {
	return &TaskHandler{Service: service}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req TaskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var recurrence *taskusecase.RecurrenceInput
	if req.Recurrence != nil {
		recurrence = &taskusecase.RecurrenceInput{
			Type:   domain.RecurrenceType(req.Recurrence.Type),
			Config: req.Recurrence.Config,
		}
	}

	created, err := h.Service.Create(r.Context(), taskusecase.CreateInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		TaskDate:    req.TaskDate,
		Recurrence:  recurrence,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	if created == nil {
		writeError(w, http.StatusInternalServerError, errors.New("created task is nil"))
		return
	}

	writeJSON(w, http.StatusCreated, newTaskDTO(created))
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskDTO(task))
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req TaskMutationDTO
	if err = decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.Service.Update(r.Context(), id, taskusecase.UpdateInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskDTO(updated))
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err = h.Service.Delete(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.Service.List(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]TaskDTO, 0, len(tasks))
	for i := range tasks {
		response = append(response, newTaskDTO(tasks[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TaskHandler) DeactivateTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err = h.Service.DeactivateTemplate(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "disabled"})
}

func (h *TaskHandler) ActivateTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err = h.Service.ActivateTemplate(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "enabled"})
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
		return fmt.Errorf("could not decode body: %w", err)
	}

	return nil
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
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
