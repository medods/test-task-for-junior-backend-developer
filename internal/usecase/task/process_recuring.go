package task

import (
    "context"
    "log"
    "time"

    taskdomain "example.com/taskservice/internal/domain/task"
    "example.com/taskservice/internal/domain/task/periodicity"
)

// ProcessRecurringTasksUseCase отвечает за генерацию экземпляров периодических задач.
type ProcessRecurringTasksUseCase struct {
    repo         Repository
    createInstance *CreateInstanceUseCase
}

func NewProcessRecurringTasksUseCase(repo Repository, createUC *CreateInstanceUseCase) *ProcessRecurringTasksUseCase {
    return &ProcessRecurringTasksUseCase{
        repo:         repo,
        createInstance: createUC,
    }
}

// Execute обрабатывает все периодические задачи, создавая необходимые экземпляры.
func (uc *ProcessRecurringTasksUseCase) Execute(ctx context.Context) error {
    now := time.Now().UTC()

    tasks, err := uc.repo.ListPeriodic(ctx)
    if err != nil {
        return err
    }

    for _, t := range tasks {
        if err := uc.processTask(ctx, t, now); err != nil {
            log.Printf("failed to process task %d: %v", t.ID, err)
            // продолжаем обработку остальных
        }
    }
    return nil
}

func (uc *ProcessRecurringTasksUseCase) processTask(ctx context.Context, t *taskdomain.Task, now time.Time) error {
    // Инициализация LastRunAt
    if t.LastRunAt == nil {
        t.LastRunAt = &t.CreatedAt
        if _, err := uc.repo.Update(ctx, t); err != nil {
            return err
        }
    }

    // Получаем стратегию
	params := &periodicity.Params{
		Type:    periodicity.PeriodicityTypeRRule,
		RRule:   t.Periodicity.RRule,
		ExDates: t.Periodicity.ExDates,
		RDates:  t.Periodicity.RDates,
		Dtstart: t.CreatedAt, // <-- важно!
	}
    strat, err := periodicity.Factory(params)
    if err != nil {
        return err
    }

    // Вычисляем следующую дату
    next, err := strat.NextAfter(*t.LastRunAt)
    if err != nil {
        return err
    }
    if next.IsZero() {
        return nil // больше нет повторений
    }
    if next.After(now) {
        return nil // ещё не время
    }

	log.Printf("[DEBUG] Task %d: LastRunAt=%v, now=%v, next=%v, next.After(now)=%v", 
    t.ID, t.LastRunAt, now, next, next.After(now))
	
    // Создаём экземпляр через отдельный usecase
    _, err = uc.createInstance.Execute(ctx, CreateInstanceInput{
        TemplateID:  t.ID,
        ScheduledFor: next,
    })
    if err != nil {
        // Обработка дубликатов внутри usecase
        if !IsDuplicateError(err) {
            return err
        }
        // если дубликат, просто логируем и продолжаем
        log.Printf("instance already exists for template %d at %s", t.ID, next.Format(time.RFC3339))
    }

	if err := uc.repo.UpdateLastRunAt(ctx, t.ID, &next); err != nil {
        return err
    }
    // Обновляем LastRunAt у шаблона
    t.LastRunAt = &next
    
    return err
}