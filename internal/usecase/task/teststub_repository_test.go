package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type stubRepository struct {
	createCalled                   bool
	getByIDCalled                  bool
	updateCalled                   bool
	deleteCalled                   bool
	listCalled                     bool
	createRecurringTemplateCalled  bool
	updateRecurringTemplateCalled  bool
	createRecurringInstancesCalled bool
	deleteRecurringTemplateCalled  bool
	listRecurringTemplatesCalled   bool

	lastCreatedTask                      *taskdomain.Task
	lastUpdatedTask                      *taskdomain.Task
	lastDeletedID                        int64
	lastCreateRecurringTemplate          *taskdomain.Task
	lastCreateRecurringTemplateDates     []time.Time
	lastUpdateRecurringTemplate          *taskdomain.Task
	lastRegenerateFrom                   *time.Time
	lastUpdateRecurringTemplateDates     []time.Time
	lastCreateRecurringInstancesTemplate *taskdomain.Task
	lastCreateRecurringInstancesDates    []time.Time
	lastDeleteRecurringTemplateID        int64

	createResult                   *taskdomain.Task
	getByIDResult                  *taskdomain.Task
	updateResult                   *taskdomain.Task
	listResult                     []taskdomain.Task
	createRecurringTemplateResult  *taskdomain.Task
	updateRecurringTemplateResult  *taskdomain.Task
	createRecurringInstancesResult []taskdomain.Task
	listRecurringTemplatesResult   []taskdomain.Task

	createErr                   error
	getByIDErr                  error
	updateErr                   error
	deleteErr                   error
	listErr                     error
	createRecurringTemplateErr  error
	updateRecurringTemplateErr  error
	createRecurringInstancesErr error
	deleteRecurringTemplateErr  error
	listRecurringTemplatesErr   error
}

func (r *stubRepository) Create(_ context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	r.createCalled = true
	r.lastCreatedTask = cloneTask(task)

	if r.createErr != nil {
		return nil, r.createErr
	}

	if r.createResult != nil {
		return cloneTask(r.createResult), nil
	}

	return cloneTask(task), nil
}

func (r *stubRepository) GetByID(_ context.Context, id int64) (*taskdomain.Task, error) {
	r.getByIDCalled = true

	if r.getByIDErr != nil {
		return nil, r.getByIDErr
	}

	return cloneTask(r.getByIDResult), nil
}

func (r *stubRepository) Update(_ context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	r.updateCalled = true
	r.lastUpdatedTask = cloneTask(task)

	if r.updateErr != nil {
		return nil, r.updateErr
	}

	if r.updateResult != nil {
		return cloneTask(r.updateResult), nil
	}

	return cloneTask(task), nil
}

func (r *stubRepository) Delete(_ context.Context, id int64) error {
	r.deleteCalled = true
	r.lastDeletedID = id

	return r.deleteErr
}

func (r *stubRepository) List(_ context.Context) ([]taskdomain.Task, error) {
	r.listCalled = true

	if r.listErr != nil {
		return nil, r.listErr
	}

	return cloneTaskSlice(r.listResult), nil
}

func (r *stubRepository) CreateRecurringTemplate(
	_ context.Context,
	template *taskdomain.Task,
	scheduledDates []time.Time,
) (*taskdomain.Task, error) {
	r.createRecurringTemplateCalled = true
	r.lastCreateRecurringTemplate = cloneTask(template)
	r.lastCreateRecurringTemplateDates = cloneTimeSlice(scheduledDates)

	if r.createRecurringTemplateErr != nil {
		return nil, r.createRecurringTemplateErr
	}

	if r.createRecurringTemplateResult != nil {
		return cloneTask(r.createRecurringTemplateResult), nil
	}

	return cloneTask(template), nil
}

func (r *stubRepository) UpdateRecurringTemplate(
	_ context.Context,
	template *taskdomain.Task,
	regenerateFrom *time.Time,
	scheduledDates []time.Time,
) (*taskdomain.Task, error) {
	r.updateRecurringTemplateCalled = true
	r.lastUpdateRecurringTemplate = cloneTask(template)
	r.lastRegenerateFrom = cloneTimePtr(regenerateFrom)
	r.lastUpdateRecurringTemplateDates = cloneTimeSlice(scheduledDates)

	if r.updateRecurringTemplateErr != nil {
		return nil, r.updateRecurringTemplateErr
	}

	if r.updateRecurringTemplateResult != nil {
		return cloneTask(r.updateRecurringTemplateResult), nil
	}

	return cloneTask(template), nil
}

func (r *stubRepository) CreateRecurringInstances(
	_ context.Context,
	template *taskdomain.Task,
	scheduledDates []time.Time,
) ([]taskdomain.Task, error) {
	r.createRecurringInstancesCalled = true
	r.lastCreateRecurringInstancesTemplate = cloneTask(template)
	r.lastCreateRecurringInstancesDates = cloneTimeSlice(scheduledDates)

	if r.createRecurringInstancesErr != nil {
		return nil, r.createRecurringInstancesErr
	}

	return cloneTaskSlice(r.createRecurringInstancesResult), nil
}

func (r *stubRepository) DeleteRecurringTemplate(_ context.Context, templateID int64) error {
	r.deleteRecurringTemplateCalled = true
	r.lastDeleteRecurringTemplateID = templateID

	return r.deleteRecurringTemplateErr
}

func (r *stubRepository) ListRecurringTemplates(_ context.Context) ([]taskdomain.Task, error) {
	r.listRecurringTemplatesCalled = true

	if r.listRecurringTemplatesErr != nil {
		return nil, r.listRecurringTemplatesErr
	}

	return cloneTaskSlice(r.listRecurringTemplatesResult), nil
}

func cloneTask(task *taskdomain.Task) *taskdomain.Task {
	if task == nil {
		return nil
	}

	cloned := *task
	cloned.ScheduledFor = cloneTimePtr(task.ScheduledFor)
	cloned.SourceTaskID = cloneInt64Ptr(task.SourceTaskID)
	cloned.Recurrence = cloneRecurrence(task.Recurrence)

	return &cloned
}

func cloneTaskSlice(tasks []taskdomain.Task) []taskdomain.Task {
	if tasks == nil {
		return nil
	}

	result := make([]taskdomain.Task, 0, len(tasks))
	for i := range tasks {
		result = append(result, *cloneTask(&tasks[i]))
	}

	return result
}

func cloneRecurrence(recurrence *taskdomain.Recurrence) *taskdomain.Recurrence {
	if recurrence == nil {
		return nil
	}

	cloned := *recurrence
	cloned.StartDate = cloneTimePtr(recurrence.StartDate)
	cloned.Dates = cloneTimeSlice(recurrence.Dates)

	return &cloned
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	cloned := *value

	return &cloned
}

func cloneTimeSlice(values []time.Time) []time.Time {
	if values == nil {
		return nil
	}

	result := make([]time.Time, 0, len(values))
	result = append(result, values...)

	return result
}

func cloneInt64Ptr(value *int64) *int64 {
	if value == nil {
		return nil
	}

	cloned := *value

	return &cloned
}
