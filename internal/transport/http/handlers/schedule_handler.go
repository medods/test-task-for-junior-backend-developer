package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	scheduleusecase "example.com/taskservice/internal/usecase/schedule"
)

type TaskScheduleHandler struct {
	usecase scheduleusecase.Usecase
}

func NewTaskScheduleHandler(usecase scheduleusecase.Usecase) *TaskScheduleHandler {
	return &TaskScheduleHandler{usecase: usecase}
}

func (h *TaskScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	input, err := decodeScheduleMutationRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), input)
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newTaskScheduleDTO(created))
}

func (h *TaskScheduleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	schedule, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskScheduleDTO(schedule))
}

func (h *TaskScheduleHandler) List(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.usecase.List(r.Context())
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	response := make([]taskScheduleDTO, 0, len(schedules))
	for i := range schedules {
		response = append(response, newTaskScheduleDTO(&schedules[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TaskScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input, err := decodeScheduleMutationRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, scheduleusecase.UpdateInput(input))
	if err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskScheduleDTO(updated))
}

func (h *TaskScheduleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Deactivate(r.Context(), id); err != nil {
		writeScheduleUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func decodeScheduleMutationRequest(r *http.Request) (scheduleusecase.CreateInput, error) {
	var req taskScheduleMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		return scheduleusecase.CreateInput{}, err
	}

	startDate, err := scheduledomain.ParseDate(req.StartDate)
	if err != nil {
		return scheduleusecase.CreateInput{}, fmt.Errorf("invalid start_date")
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsed, err := scheduledomain.ParseDate(*req.EndDate)
		if err != nil {
			return scheduleusecase.CreateInput{}, fmt.Errorf("invalid end_date")
		}

		endDate = &parsed
	}

	rule, err := recurrenceDTOToDomain(req.Recurrence)
	if err != nil {
		return scheduleusecase.CreateInput{}, err
	}

	return scheduleusecase.CreateInput{
		Title:       req.Title,
		Description: req.Description,
		StartDate:   startDate,
		EndDate:     endDate,
		Active:      req.Active,
		Rule:        rule,
	}, nil
}

func recurrenceDTOToDomain(input recurrenceDTO) (scheduledomain.RecurrenceRule, error) {
	switch scheduledomain.RuleType(input.Type) {
	case scheduledomain.RuleTypeDailyInterval:
		if input.EveryDays == nil {
			return nil, errors.New("recurrence.every_days is required")
		}
		return scheduledomain.DailyIntervalRule{EveryDays: *input.EveryDays}, nil
	case scheduledomain.RuleTypeMonthlyDay:
		if input.DayOfMonth == nil {
			return nil, errors.New("recurrence.day_of_month is required")
		}
		return scheduledomain.MonthlyDayRule{DayOfMonth: *input.DayOfMonth}, nil
	case scheduledomain.RuleTypeSpecificDates:
		dates := make([]time.Time, 0, len(input.Dates))
		for _, raw := range input.Dates {
			parsed, err := scheduledomain.ParseDate(raw)
			if err != nil {
				return nil, fmt.Errorf("invalid recurrence.dates")
			}
			dates = append(dates, parsed)
		}
		return scheduledomain.SpecificDatesRule{Dates: dates}, nil
	case scheduledomain.RuleTypeMonthParity:
		return scheduledomain.MonthParityRule{Parity: scheduledomain.MonthParity(input.Parity)}, nil
	case scheduledomain.RuleTypeWeekly:
		return nil, errors.New("weekly recurrence is reserved and not supported yet")
	default:
		return nil, errors.New("invalid recurrence.type")
	}
}

func newTaskScheduleDTO(model *scheduledomain.TaskSchedule) taskScheduleDTO {
	var endDate *string
	if model.EndDate != nil {
		formatted := scheduledomain.FormatDate(*model.EndDate)
		endDate = &formatted
	}

	return taskScheduleDTO{
		ID:          model.ID,
		Title:       model.Title,
		Description: model.Description,
		StartDate:   scheduledomain.FormatDate(model.StartDate),
		EndDate:     endDate,
		Active:      model.Active,
		Recurrence:  recurrenceDomainToDTO(model.Rule),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func recurrenceDomainToDTO(rule scheduledomain.RecurrenceRule) recurrenceDTO {
	switch typed := rule.(type) {
	case scheduledomain.DailyIntervalRule:
		everyDays := typed.EveryDays
		return recurrenceDTO{Type: string(typed.Type()), EveryDays: &everyDays}
	case *scheduledomain.DailyIntervalRule:
		everyDays := typed.EveryDays
		return recurrenceDTO{Type: string(typed.Type()), EveryDays: &everyDays}
	case scheduledomain.MonthlyDayRule:
		dayOfMonth := typed.DayOfMonth
		return recurrenceDTO{Type: string(typed.Type()), DayOfMonth: &dayOfMonth}
	case *scheduledomain.MonthlyDayRule:
		dayOfMonth := typed.DayOfMonth
		return recurrenceDTO{Type: string(typed.Type()), DayOfMonth: &dayOfMonth}
	case scheduledomain.SpecificDatesRule:
		return recurrenceDTO{Type: string(typed.Type()), Dates: formatSpecificDates(typed.Dates)}
	case *scheduledomain.SpecificDatesRule:
		return recurrenceDTO{Type: string(typed.Type()), Dates: formatSpecificDates(typed.Dates)}
	case scheduledomain.MonthParityRule:
		return recurrenceDTO{Type: string(typed.Type()), Parity: string(typed.Parity)}
	case *scheduledomain.MonthParityRule:
		return recurrenceDTO{Type: string(typed.Type()), Parity: string(typed.Parity)}
	default:
		return recurrenceDTO{}
	}
}

func formatSpecificDates(dates []time.Time) []string {
	result := make([]string, 0, len(dates))
	for _, value := range dates {
		result = append(result, scheduledomain.FormatDate(value))
	}

	return result
}
