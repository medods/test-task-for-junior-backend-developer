package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type TaskHandler struct {
	usecase taskusecase.Usecase
}

func NewTaskHandler(usecase taskusecase.Usecase) *TaskHandler {
	return &TaskHandler{usecase: usecase}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	// ВАЛИДАЦИЯ ПЕРИОДИЧНОСТИ НА УРОВНЕ HANDLER
	if err := validateRecurrenceFields(req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), taskusecase.CreateInput{
		Title:            req.Title,
		Description:      req.Description,
		Status:           req.Status,
		RecurrenceType:   req.RecurrenceType,
		RecurrenceEveryN: req.RecurrenceEveryN,
		MonthlyDay:       req.MonthlyDay,
		SpecificDates:    req.SpecificDates,
		EvenOdd:          req.EvenOdd,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newTaskDTO(created))
}

func validateRecurrenceFields(req taskMutationDTO) error {
	// Если тип не указан, пропускаем (обычная задача)
	if req.RecurrenceType == "" {
		return nil
	}

	switch req.RecurrenceType {
	case taskdomain.RecurrenceDaily:
		if req.RecurrenceEveryN == nil || *req.RecurrenceEveryN < 1 {
			return errors.New("daily recurrence requires recurrence_every_n >= 1")
		}

	case taskdomain.RecurrenceMonthlyDay:
		if req.MonthlyDay == nil || *req.MonthlyDay < 1 || *req.MonthlyDay > 30 {
			return errors.New("monthly_day must be between 1 and 30")
		}

	case taskdomain.RecurrenceSpecific:
		if len(req.SpecificDates) == 0 {
			return errors.New("specific dates list cannot be empty")
		}

	case taskdomain.RecurrenceEvenOdd:
		if req.EvenOdd == nil || (*req.EvenOdd != "even" && *req.EvenOdd != "odd") {
			return errors.New("even_odd must be 'even' or 'odd'")
		}

	default:
		return fmt.Errorf("invalid recurrence_type: %s", req.RecurrenceType)
	}

	return nil
}

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

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	// Валидация периодичности, если поля переданы
	if req.RecurrenceType != "" {
		if err := validateRecurrenceFields(req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
	}

	// Конвертируем string в *string для полей, которые могут быть nil
	var titlePtr *string
	if req.Title != "" {
		titlePtr = &req.Title
	}

	var descriptionPtr *string
	if req.Description != "" {
		descriptionPtr = &req.Description
	}

	var statusPtr *taskdomain.Status
	if req.Status != "" {
		statusPtr = &req.Status
	}

	updated, err := h.usecase.Update(r.Context(), id, taskusecase.UpdateInput{
		Title:            titlePtr,
		Description:      descriptionPtr,
		Status:           statusPtr,
		RecurrenceType:   &req.RecurrenceType,  // RecurrenceType тоже может быть указателем
		RecurrenceEveryN: req.RecurrenceEveryN, // уже *int,没问题
		MonthlyDay:       req.MonthlyDay,       // уже *int
		SpecificDates:    req.SpecificDates,    // []time.Time
		EvenOdd:          req.EvenOdd,          // уже *string
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

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.usecase.List(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
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
	case errors.Is(err, taskdomain.ErrNotFound):
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
