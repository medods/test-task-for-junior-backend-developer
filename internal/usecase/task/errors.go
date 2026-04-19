package task

import "errors"

var ErrInvalidInput = errors.New("invalid task input")

var ErrInvalidStartEndDates = errors.New("обязательны start_date и end_date")
var ErrEndBeforeStartDate = errors.New("end_date не может быть раньше start_date")

var ErrInvalidRecurrence = errors.New("invalid recurrence input")
