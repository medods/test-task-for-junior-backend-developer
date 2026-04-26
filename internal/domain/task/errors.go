package task

import "errors"

var ErrNotFound = errors.New("task not found")
var ErrRecurrenceNotFound = errors.New("recurrence not found")
