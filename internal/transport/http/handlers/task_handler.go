package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

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

	var scheduledFor *time.Time
	if req.ScheduledFor != nil {
		parsedTime, err := time.Parse("2006-01-02", *req.ScheduledFor)
		if err != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid scheduled_for format, expected YYYY-MM-DD"))
			return
		}
		scheduledFor = &parsedTime
	}

	var recurrence *taskusecase.RecurrenceInput
	if req.Recurrence != nil {
		parsedStartDate, err := time.Parse("2006-01-02", req.Recurrence.StartDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid start_date format, expected YYYY-MM-DD"))
			return
		}

		var endDate *time.Time
		if req.Recurrence.EndDate != nil {
			parsedEndDate, err := time.Parse("2006-01-02", *req.Recurrence.EndDate)
			if err != nil {
				writeError(w, http.StatusBadRequest, errors.New("invalid end_date format, expected YYYY-MM-DD"))
				return
			}
			endDate = &parsedEndDate
		}

		specificDates := make([]time.Time, 0, len(req.Recurrence.SpecificDates))
		for _, rawDate := range req.Recurrence.SpecificDates {
			parsedDate, err := time.Parse("2006-01-02", rawDate)
			if err != nil {
				writeError(w, http.StatusBadRequest, errors.New("invalid specific_dates format, expected YYYY-MM-DD"))
				return
			}
			specificDates = append(specificDates, parsedDate)
		}

		recurrence = &taskusecase.RecurrenceInput{
			Type:          taskdomain.RecurrenceType(req.Recurrence.Type),
			StartDate:     parsedStartDate,
			EndDate:       endDate,
			EveryNDays:    req.Recurrence.EveryNDays,
			DayOfMonth:    req.Recurrence.DayOfMonth,
			SpecificDates: specificDates,
		}
	}

	created, err := h.usecase.Create(r.Context(), taskusecase.CreateInput{
		Title:        req.Title,
		Description:  req.Description,
		Status:       req.Status,
		ScheduledFor: scheduledFor,
		Recurrence:   recurrence,
	})
	if err != nil {
		writeUsecaseError(w, err)
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

	updated, err := h.usecase.Update(r.Context(), id, taskusecase.UpdateInput{
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
