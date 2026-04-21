package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	"example.com/taskservice/internal/shared/dateutil"
	recurrenceusecase "example.com/taskservice/internal/usecase/recurrence"
)

type RecurrenceHandler struct {
	usecase recurrenceusecase.Usecase
}

func NewRecurrenceHandler(usecase recurrenceusecase.Usecase) *RecurrenceHandler {
	return &RecurrenceHandler{usecase: usecase}
}

func (h *RecurrenceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req recurrenceRuleMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input, err := recurrenceInputFromDTO(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), input)
	if err != nil {
		writeRecurrenceUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newRecurrenceRuleDTO(created))
}

func (h *RecurrenceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	rule, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeRecurrenceUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newRecurrenceRuleDTO(rule))
}

func (h *RecurrenceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req recurrenceRuleMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input, err := recurrenceInputFromDTO(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, input)
	if err != nil {
		writeRecurrenceUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newRecurrenceRuleDTO(updated))
}

func (h *RecurrenceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeRecurrenceUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RecurrenceHandler) List(w http.ResponseWriter, r *http.Request) {
	rules, err := h.usecase.List(r.Context())
	if err != nil {
		writeRecurrenceUsecaseError(w, err)
		return
	}

	response := make([]recurrenceRuleDTO, 0, len(rules))
	for i := range rules {
		response = append(response, newRecurrenceRuleDTO(&rules[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func writeRecurrenceUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, recurrencedomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, recurrenceusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func recurrenceInputFromDTO(req recurrenceRuleMutationDTO) (recurrenceusecase.CreateInput, error) {
	startsOn, err := parseOptionalDateField("schedule.starts_on", req.Schedule.StartsOn)
	if err != nil {
		return recurrenceusecase.CreateInput{}, err
	}

	endsOn, err := parseOptionalDateField("schedule.ends_on", req.Schedule.EndsOn)
	if err != nil {
		return recurrenceusecase.CreateInput{}, err
	}

	dates := make([]time.Time, 0, len(req.Schedule.Dates))
	for _, rawDate := range req.Schedule.Dates {
		parsed, err := parseDateField("schedule.dates", rawDate)
		if err != nil {
			return recurrenceusecase.CreateInput{}, err
		}
		dates = append(dates, parsed)
	}

	return recurrenceusecase.CreateInput{
		Title:         req.Title,
		Description:   req.Description,
		Active:        req.Active,
		ScheduleType:  req.Schedule.Type,
		StartsOn:      startsOn,
		EndsOn:        endsOn,
		EveryNDays:    req.Schedule.EveryNDays,
		DayOfMonth:    req.Schedule.DayOfMonth,
		DayParity:     req.Schedule.DayParity,
		SpecificDates: dates,
	}, nil
}

func parseDateField(fieldName, rawValue string) (time.Time, error) {
	parsed, err := dateutil.Parse(rawValue)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s must be in YYYY-MM-DD format", fieldName)
	}

	return parsed, nil
}

func parseOptionalDateField(fieldName, rawValue string) (*time.Time, error) {
	if rawValue == "" {
		return nil, nil
	}

	parsed, err := parseDateField(fieldName, rawValue)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
