package handlers

import (
	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

func (dto taskCreateDTO) toCreateInput() (taskusecase.CreateInput, error) {
	mode := dto.Mode
	if mode == "" {
		mode = taskdomain.ModeOneTime
	}

	recurrence, err := recurrenceDTOToDomain(dto.Recurrence)
	if err != nil {
		return taskusecase.CreateInput{}, err
	}

	return taskusecase.CreateInput{
		Title:       dto.Title,
		Description: dto.Description,
		Status:      dto.Status,
		Mode:        mode,
		Recurrence:  recurrence,
	}, nil
}

func (dto taskUpdateDTO) toUpdateInput() (taskusecase.UpdateInput, error) {
	recurrence, err := recurrenceDTOToDomain(dto.Recurrence)
	if err != nil {
		return taskusecase.UpdateInput{}, err
	}

	return taskusecase.UpdateInput{
		Title:       dto.Title,
		Description: dto.Description,
		Status:      dto.Status,
		Mode:        dto.Mode,
		Recurrence:  recurrence,
	}, nil
}

func recurrenceDTOToDomain(dto *recurrenceDTO) (*taskdomain.Recurrence, error) {
	if dto == nil {
		return nil, nil
	}

	startDate, err := parseDatePtr(dto.StartDate)
	if err != nil {
		return nil, err
	}

	dates, err := parseDateSlice(dto.Dates)
	if err != nil {
		return nil, err
	}

	return &taskdomain.Recurrence{
		Type:       dto.Type,
		EveryNDays: dto.EveryNDays,
		DayOfMonth: dto.DayOfMonth,
		StartDate:  startDate,
		Dates:      dates,
	}, nil
}

func newRecurrenceDTO(recurrence *taskdomain.Recurrence) *recurrenceDTO {
	if recurrence == nil {
		return nil
	}

	return &recurrenceDTO{
		Type:       recurrence.Type,
		EveryNDays: recurrence.EveryNDays,
		DayOfMonth: recurrence.DayOfMonth,
		StartDate:  formatDatePtr(recurrence.StartDate),
		Dates:      formatDateSlice(recurrence.Dates),
	}
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:           task.ID,
		Title:        task.Title,
		Description:  task.Description,
		Status:       task.Status,
		Mode:         task.Mode,
		ScheduledFor: formatDatePtr(task.ScheduledFor),
		SourceTaskID: task.SourceTaskID,
		Recurrence:   newRecurrenceDTO(task.Recurrence),
		CreatedAt:    task.CreatedAt.UTC(),
		UpdatedAt:    task.UpdatedAt.UTC(),
	}
}
