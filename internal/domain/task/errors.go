package task

import "errors"

var (
	ErrNotFound                  = errors.New("task not found")
	ErrTaskHasGeneratedInstances = errors.New("remove failed, task has generated instances")
)
