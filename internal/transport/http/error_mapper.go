package http

import (
	"errors"
	"net/http"

	"your-project/internal/domain/task"
	"your-project/internal/logger"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func MapErrorToHTTP(log logger.Logger, err error) (int, APIError) {
	log.Error("request failed", "error", err)

	switch {
	case errors.Is(err, task.ErrTypeRequired):
		return http.StatusBadRequest, APIError{
			Code: "TYPE_REQUIRED", Message: "Укажите тип периодичности", Field: "type",
		}
	case errors.Is(err, task.ErrStartDateRequired):
		return http.StatusBadRequest, APIError{
			Code: "START_DATE_REQUIRED", Message: "Укажите дату начала расписания", Field: "start_date",
		}
	case errors.Is(err, task.ErrStartDateAfterEnd):
		return http.StatusBadRequest, APIError{
			Code: "DATE_ORDER_INVALID", Message: "Дата начала не может быть позже даты окончания", Field: "dates",
		}
	case errors.Is(err, task.ErrInvalidInterval):
		return http.StatusBadRequest, APIError{
			Code: "INTERVAL_INVALID", Message: "Интервал должен быть не менее 1 дня", Field: "interval",
		}
	case errors.Is(err, task.ErrInvalidDayOfMonth):
		return http.StatusBadRequest, APIError{
			Code: "DAY_OF_MONTH_INVALID", Message: "День месяца должен быть от 1 до 31", Field: "day_of_month",
		}
	case errors.Is(err, task.ErrInvalidOverflowStrategy):
		return http.StatusBadRequest, APIError{
			Code: "OVERFLOW_STRATEGY_INVALID", Message: "Стратегия переполнения должна быть 'skip', 'last_day' или 'next_month'", Field: "overflow_strategy",
		}
	case errors.Is(err, task.ErrEmptyDates):
		return http.StatusBadRequest, APIError{
			Code: "DATES_EMPTY", Message: "Список конкретных дат не может быть пустым", Field: "specific_dates",
		}
	case errors.Is(err, task.ErrDuplicateDate):
		return http.StatusBadRequest, APIError{
			Code: "DATES_DUPLICATE", Message: "Обнаружены повторяющиеся даты в списке", Field: "specific_dates",
		}
	case errors.Is(err, task.ErrInvalidDateInList):
		return http.StatusBadRequest, APIError{
			Code: "DATE_INVALID", Message: "Список содержит нулевые или некорректные даты", Field: "specific_dates",
		}
	case errors.Is(err, task.ErrInvalidParityMode):
		return http.StatusBadRequest, APIError{
			Code: "PARITY_MODE_INVALID", Message: "Режим чётности должен быть 'even' или 'odd'", Field: "parity_mode",
		}
	case errors.Is(err, task.ErrUnknownType):
		return http.StatusBadRequest, APIError{
			Code: "TYPE_UNKNOWN", Message: "Неподдерживаемый тип периодичности", Field: "type",
		}
	case errors.Is(err, task.ErrInvalidTimezone):
		return http.StatusBadRequest, APIError{
			Code: "TIMEZONE_INVALID", Message: "Указан неверный формат часового пояса (IANA)", Field: "timezone",
		}
	case errors.Is(err, task.ErrTaskNotFound):
		return http.StatusNotFound, APIError{
			Code: "TASK_NOT_FOUND", Message: "Задача не найдена",
		}

	default:
		return http.StatusInternalServerError, APIError{
			Code:    "INTERNAL_ERROR",
			Message: "Внутренняя ошибка сервера",
		}
	}
}