package task

import (
	"context"
	"errors"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func TestServiceCreateOneTimeDefaultsAndTrim(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)
	now := mustDateTime(2026, time.April, 28, 10, 15, time.UTC)
	service.now = func() time.Time { return now }

	created, err := service.Create(context.Background(), CreateInput{
		Title:       "  Ward round  ",
		Description: "  Check patients  ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if !repo.createCalled {
		t.Fatal("expected Create() to call repo.Create")
	}

	if repo.createRecurringTemplateCalled {
		t.Fatal("did not expect recurring template creation for one-time task")
	}

	if repo.lastCreatedTask == nil {
		t.Fatal("expected captured created task")
	}

	if repo.lastCreatedTask.Title != "Ward round" {
		t.Fatalf("title = %q, want %q", repo.lastCreatedTask.Title, "Ward round")
	}

	if repo.lastCreatedTask.Description != "Check patients" {
		t.Fatalf("description = %q, want %q", repo.lastCreatedTask.Description, "Check patients")
	}

	if repo.lastCreatedTask.Status != taskdomain.StatusNew {
		t.Fatalf("status = %q, want %q", repo.lastCreatedTask.Status, taskdomain.StatusNew)
	}

	if repo.lastCreatedTask.Mode != taskdomain.ModeOneTime {
		t.Fatalf("mode = %q, want %q", repo.lastCreatedTask.Mode, taskdomain.ModeOneTime)
	}

	if !repo.lastCreatedTask.CreatedAt.Equal(now) || !repo.lastCreatedTask.UpdatedAt.Equal(now) {
		t.Fatalf("timestamps = %v/%v, want %v", repo.lastCreatedTask.CreatedAt, repo.lastCreatedTask.UpdatedAt, now)
	}

	if created == nil || created.Mode != taskdomain.ModeOneTime {
		t.Fatalf("created task = %+v, want one_time task", created)
	}
}

func TestServiceCreateRecurringTemplateGeneratesInstances(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)
	now := mustDateTime(2026, time.April, 28, 9, 0, time.UTC)
	service.now = func() time.Time { return now }

	startDate := mustDateTime(2026, time.April, 28, 18, 30, time.FixedZone("UTC+3", 3*60*60))

	_, err := service.Create(context.Background(), CreateInput{
		Title:       "Call patients",
		Description: "Daily follow-up calls",
		Status:      taskdomain.StatusNew,
		Mode:        taskdomain.ModeRecurringTemplate,
		Recurrence: &taskdomain.Recurrence{
			Type:       taskdomain.RecurrenceDaily,
			EveryNDays: 2,
			StartDate:  &startDate,
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if !repo.createRecurringTemplateCalled {
		t.Fatal("expected Create() to call repo.CreateRecurringTemplate")
	}

	if repo.lastCreateRecurringTemplate == nil {
		t.Fatal("expected captured recurring template")
	}

	if repo.lastCreateRecurringTemplate.Mode != taskdomain.ModeRecurringTemplate {
		t.Fatalf("mode = %q, want %q", repo.lastCreateRecurringTemplate.Mode, taskdomain.ModeRecurringTemplate)
	}

	if repo.lastCreateRecurringTemplate.Recurrence == nil || repo.lastCreateRecurringTemplate.Recurrence.StartDate == nil {
		t.Fatal("expected normalized recurrence on template")
	}

	if got := repo.lastCreateRecurringTemplate.Recurrence.StartDate; !got.Equal(mustDate(t, "2026-04-28")) {
		t.Fatalf("start_date = %v, want %v", *got, mustDate(t, "2026-04-28"))
	}

	wantDates := []time.Time{
		mustDate(t, "2026-04-28"),
		mustDate(t, "2026-04-30"),
		mustDate(t, "2026-05-02"),
		mustDate(t, "2026-05-04"),
		mustDate(t, "2026-05-06"),
		mustDate(t, "2026-05-08"),
		mustDate(t, "2026-05-10"),
		mustDate(t, "2026-05-12"),
		mustDate(t, "2026-05-14"),
		mustDate(t, "2026-05-16"),
		mustDate(t, "2026-05-18"),
		mustDate(t, "2026-05-20"),
		mustDate(t, "2026-05-22"),
		mustDate(t, "2026-05-24"),
		mustDate(t, "2026-05-26"),
	}

	assertDateStringsEqual(t, repo.lastCreateRecurringTemplateDates, wantDates)
}

func TestServiceCreateRejectsDirectRecurringInstance(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)

	_, err := service.Create(context.Background(), CreateInput{
		Title:  "Generated task",
		Status: taskdomain.StatusNew,
		Mode:   taskdomain.ModeRecurringInstance,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Create() error = %v, want ErrInvalidInput", err)
	}

	if repo.createCalled || repo.createRecurringTemplateCalled {
		t.Fatal("did not expect repository writes for invalid recurring instance create")
	}
}

func TestServiceUpdateRecurringTemplateRegeneratesFromToday(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)
	now := mustDateTime(2026, time.May, 2, 15, 30, time.UTC)
	service.now = func() time.Time { return now }

	existingStartDate := mustDate(t, "2026-04-28")
	repo.getByIDResult = &taskdomain.Task{
		ID:          42,
		Title:       "Call patients",
		Description: "Old description",
		Status:      taskdomain.StatusNew,
		Mode:        taskdomain.ModeRecurringTemplate,
		Recurrence: &taskdomain.Recurrence{
			Type:       taskdomain.RecurrenceDaily,
			EveryNDays: 1,
			StartDate:  &existingStartDate,
		},
		CreatedAt: mustDateTime(2026, time.April, 20, 9, 0, time.UTC),
		UpdatedAt: mustDateTime(2026, time.April, 20, 9, 0, time.UTC),
	}

	newStartDate := mustDateTime(2026, time.April, 30, 21, 15, time.FixedZone("UTC+3", 3*60*60))

	updated, err := service.Update(context.Background(), 42, UpdateInput{
		Title:       "  Updated title  ",
		Description: "  Updated description  ",
		Status:      taskdomain.StatusInProgress,
		Recurrence: &taskdomain.Recurrence{
			Type:       taskdomain.RecurrenceDaily,
			EveryNDays: 3,
			StartDate:  &newStartDate,
		},
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if !repo.updateRecurringTemplateCalled {
		t.Fatal("expected Update() to call repo.UpdateRecurringTemplate")
	}

	if repo.lastUpdateRecurringTemplate == nil {
		t.Fatal("expected captured updated template")
	}

	if repo.lastUpdateRecurringTemplate.Title != "Updated title" {
		t.Fatalf("title = %q, want %q", repo.lastUpdateRecurringTemplate.Title, "Updated title")
	}

	if repo.lastUpdateRecurringTemplate.Description != "Updated description" {
		t.Fatalf("description = %q, want %q", repo.lastUpdateRecurringTemplate.Description, "Updated description")
	}

	if repo.lastRegenerateFrom == nil || !repo.lastRegenerateFrom.Equal(mustDate(t, "2026-05-02")) {
		t.Fatalf("regenerateFrom = %v, want %v", repo.lastRegenerateFrom, mustDate(t, "2026-05-02"))
	}

	if repo.lastUpdateRecurringTemplate.Recurrence == nil || repo.lastUpdateRecurringTemplate.Recurrence.StartDate == nil {
		t.Fatal("expected recurrence on updated template")
	}

	if got := repo.lastUpdateRecurringTemplate.Recurrence.StartDate; !got.Equal(mustDate(t, "2026-04-30")) {
		t.Fatalf("start_date = %v, want %v", *got, mustDate(t, "2026-04-30"))
	}

	wantDates := []time.Time{
		mustDate(t, "2026-05-03"),
		mustDate(t, "2026-05-06"),
		mustDate(t, "2026-05-09"),
		mustDate(t, "2026-05-12"),
		mustDate(t, "2026-05-15"),
		mustDate(t, "2026-05-18"),
		mustDate(t, "2026-05-21"),
		mustDate(t, "2026-05-24"),
		mustDate(t, "2026-05-27"),
		mustDate(t, "2026-05-30"),
	}

	assertDateStringsEqual(t, repo.lastUpdateRecurringTemplateDates, wantDates)

	if updated == nil || updated.Mode != taskdomain.ModeRecurringTemplate {
		t.Fatalf("updated task = %+v, want recurring_template", updated)
	}
}

func TestServiceUpdateRecurringInstanceRejectsRecurrence(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)
	scheduledFor := mustDate(t, "2026-05-10")
	sourceTaskID := int64(7)

	repo.getByIDResult = &taskdomain.Task{
		ID:           100,
		Title:        "Generated task",
		Description:  "Instance",
		Status:       taskdomain.StatusNew,
		Mode:         taskdomain.ModeRecurringInstance,
		ScheduledFor: &scheduledFor,
		SourceTaskID: &sourceTaskID,
	}

	startDate := mustDate(t, "2026-05-10")
	_, err := service.Update(context.Background(), 100, UpdateInput{
		Title:       "Generated task",
		Description: "Instance",
		Status:      taskdomain.StatusDone,
		Recurrence: &taskdomain.Recurrence{
			Type:       taskdomain.RecurrenceDaily,
			EveryNDays: 1,
			StartDate:  &startDate,
		},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Update() error = %v, want ErrInvalidInput", err)
	}

	if repo.updateCalled {
		t.Fatal("did not expect repo.Update for invalid recurring instance update")
	}
}

func TestServiceUpdateRejectsModeChange(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)

	repo.getByIDResult = &taskdomain.Task{
		ID:          5,
		Title:       "One-time task",
		Description: "Task",
		Status:      taskdomain.StatusNew,
		Mode:        taskdomain.ModeOneTime,
	}

	newMode := taskdomain.ModeRecurringTemplate
	startDate := mustDate(t, "2026-05-01")

	_, err := service.Update(context.Background(), 5, UpdateInput{
		Title:       "One-time task",
		Description: "Task",
		Status:      taskdomain.StatusNew,
		Mode:        &newMode,
		Recurrence: &taskdomain.Recurrence{
			Type:       taskdomain.RecurrenceDaily,
			EveryNDays: 1,
			StartDate:  &startDate,
		},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Update() error = %v, want ErrInvalidInput", err)
	}

	if repo.updateCalled || repo.updateRecurringTemplateCalled {
		t.Fatal("did not expect repository update when mode change is forbidden")
	}
}

func TestServiceDeleteRecurringTemplateUsesSpecialDelete(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)

	repo.getByIDResult = &taskdomain.Task{
		ID:    77,
		Mode:  taskdomain.ModeRecurringTemplate,
		Title: "Inventory check",
	}

	if err := service.Delete(context.Background(), 77); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if !repo.deleteRecurringTemplateCalled {
		t.Fatal("expected Delete() to call repo.DeleteRecurringTemplate")
	}

	if repo.deleteCalled {
		t.Fatal("did not expect generic repo.Delete for recurring template")
	}
}

func TestServiceListSyncsRecurringTemplatesBeforeReturningTasks(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)
	now := mustDateTime(2026, time.April, 28, 8, 0, time.UTC)
	service.now = func() time.Time { return now }

	startDate := mustDate(t, "2026-04-28")
	repo.listRecurringTemplatesResult = []taskdomain.Task{
		{
			ID:          12,
			Title:       "Call patients",
			Description: "Daily follow-up calls",
			Status:      taskdomain.StatusNew,
			Mode:        taskdomain.ModeRecurringTemplate,
			Recurrence: &taskdomain.Recurrence{
				Type:       taskdomain.RecurrenceDaily,
				EveryNDays: 2,
				StartDate:  &startDate,
			},
		},
	}
	repo.listResult = []taskdomain.Task{
		{
			ID:     1,
			Title:  "One-time task",
			Status: taskdomain.StatusNew,
			Mode:   taskdomain.ModeOneTime,
		},
	}

	tasks, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if !repo.listRecurringTemplatesCalled {
		t.Fatal("expected List() to call repo.ListRecurringTemplates")
	}

	if !repo.createRecurringInstancesCalled {
		t.Fatal("expected List() to synchronize recurring instances")
	}

	if !repo.listCalled {
		t.Fatal("expected List() to call repo.List")
	}

	if repo.lastCreateRecurringInstancesTemplate == nil || repo.lastCreateRecurringInstancesTemplate.ID != 12 {
		t.Fatalf("synchronized template = %+v, want template id 12", repo.lastCreateRecurringInstancesTemplate)
	}

	wantDates := []time.Time{
		mustDate(t, "2026-04-28"),
		mustDate(t, "2026-04-30"),
		mustDate(t, "2026-05-02"),
		mustDate(t, "2026-05-04"),
		mustDate(t, "2026-05-06"),
		mustDate(t, "2026-05-08"),
		mustDate(t, "2026-05-10"),
		mustDate(t, "2026-05-12"),
		mustDate(t, "2026-05-14"),
		mustDate(t, "2026-05-16"),
		mustDate(t, "2026-05-18"),
		mustDate(t, "2026-05-20"),
		mustDate(t, "2026-05-22"),
		mustDate(t, "2026-05-24"),
		mustDate(t, "2026-05-26"),
	}

	assertDateStringsEqual(t, repo.lastCreateRecurringInstancesDates, wantDates)

	if len(tasks) != 1 || tasks[0].ID != 1 {
		t.Fatalf("List() returned %+v, want one task with id 1", tasks)
	}
}
