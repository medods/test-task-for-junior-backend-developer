package task

import "errors"

// ErrNotFound is returned when a task cannot be found.
var ErrNotFound = errors.New("task not found")
