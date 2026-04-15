package materializer

import (
	"context"
	"fmt"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	schedules  ScheduleRepository
	tasks      TaskRepository
	location   *time.Location
	windowDays int
	now        func() time.Time
}

func NewService(schedules ScheduleRepository, tasks TaskRepository, location *time.Location, windowDays int) *Service {
	return &Service{
		schedules:  schedules,
		tasks:      tasks,
		location:   location,
		windowDays: windowDays,
		now:        time.Now,
	}
}

func (s *Service) DefaultWindow() (time.Time, time.Time) {
	from := scheduledomain.DateFromLocation(s.now(), s.location)
	to := from.AddDate(0, 0, s.windowDays)
	return from, to
}

func (s *Service) MaterializeDefaultWindow(ctx context.Context) error {
	from, to := s.DefaultWindow()
	return s.MaterializeWindow(ctx, from, to)
}

func (s *Service) MaterializeWindow(ctx context.Context, from time.Time, to time.Time) error {
	windowStart, windowEnd, err := normalizeWindow(from, to)
	if err != nil {
		return err
	}

	schedules, err := s.schedules.ListActiveBetween(ctx, windowStart, windowEnd)
	if err != nil {
		return err
	}

	for i := range schedules {
		if err := s.materializeSchedule(ctx, &schedules[i], windowStart, windowEnd); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) MaterializeScheduleWindow(ctx context.Context, scheduleID int64, from time.Time, to time.Time) error {
	windowStart, windowEnd, err := normalizeWindow(from, to)
	if err != nil {
		return err
	}

	schedule, err := s.schedules.GetByID(ctx, scheduleID)
	if err != nil {
		return err
	}

	if !schedule.Active {
		return nil
	}

	return s.materializeSchedule(ctx, schedule, windowStart, windowEnd)
}

func (s *Service) RebuildScheduleWindow(ctx context.Context, scheduleID int64, from time.Time, to time.Time) error {
	windowStart, windowEnd, err := normalizeWindow(from, to)
	if err != nil {
		return err
	}

	if err := s.tasks.DeleteFutureNewGenerated(ctx, scheduleID, windowStart, windowEnd); err != nil {
		return err
	}

	return s.MaterializeScheduleWindow(ctx, scheduleID, windowStart, windowEnd)
}

func (s *Service) materializeSchedule(ctx context.Context, schedule *scheduledomain.TaskSchedule, from time.Time, to time.Time) error {
	occurrences := schedule.OccurrenceDates(from, to)
	if len(occurrences) == 0 {
		return nil
	}

	now := s.now().UTC()
	for _, occurrence := range occurrences {
		scheduleID := schedule.ID
		scheduledFor := occurrence
		task := &taskdomain.Task{
			Title:        schedule.Title,
			Description:  schedule.Description,
			Status:       taskdomain.StatusNew,
			CreatedAt:    now,
			UpdatedAt:    now,
			ScheduledFor: &scheduledFor,
			ScheduleID:   &scheduleID,
		}

		if err := s.tasks.CreateGenerated(ctx, task); err != nil {
			return err
		}
	}

	return nil
}

func normalizeWindow(from time.Time, to time.Time) (time.Time, time.Time, error) {
	windowStart := scheduledomain.NormalizeDate(from)
	windowEnd := scheduledomain.NormalizeDate(to)
	if windowEnd.Before(windowStart) {
		return time.Time{}, time.Time{}, fmt.Errorf("date_to must be greater than or equal to date_from")
	}

	return windowStart, windowEnd, nil
}
