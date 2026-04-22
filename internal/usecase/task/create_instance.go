package task

import (
    "context"
    "time"

    taskdomain "example.com/taskservice/internal/domain/task"
)

type CreateInstanceInput struct {
    TemplateID   int64
    ScheduledFor time.Time
}

type CreateInstanceUseCase struct {
    repo Repository
}

func NewCreateInstanceUseCase(repo Repository) *CreateInstanceUseCase {
    return &CreateInstanceUseCase{repo: repo}
}

func (uc *CreateInstanceUseCase) Execute(ctx context.Context, input CreateInstanceInput) (*taskdomain.Task, error) {
    // Получаем шаблон
    template, err := uc.repo.GetByID(ctx, input.TemplateID)
    if err != nil {
        return nil, err
    }

    // Создаём копию
    instance := *template
    instance.ID = 0
    instance.Status = taskdomain.StatusNew
    instance.ParentID = &template.ID
    instance.Periodicity = nil
    instance.ScheduledFor = &input.ScheduledFor
    instance.LastRunAt = nil
    instance.CreatedAt = time.Now().UTC()
    instance.UpdatedAt = instance.CreatedAt

    return uc.repo.Create(ctx, &instance)
}