package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"example.com/taskservice/internal/domain/common"
	"example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	postgresrepo "example.com/taskservice/internal/repository/postgres"
	"example.com/taskservice/internal/usecase/task"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_DSN is required")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := infrastructurepostgres.Open(ctx, dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer pool.Close()

	var count int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM tasks").Scan(&count); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if count > 0 {
		return
	}

	repo := postgresrepo.New(pool)
	uc := task.NewService(repo)

	today := common.NewDate(time.Now().UTC())
	now := time.Now().UTC()

	_, err = uc.Create(ctx, task.CreateInput{
		Title:               "One-time task",
		Description:         "Seeded",
		Status:              taskdomain.StatusNew,
		RecurrenceType:      recurrence.NoneRecurrence,
		RecurrencePayload:   recurrence.RecurrencePayload{},
		RecurrenceStartDate: &today,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, err = uc.Create(ctx, task.CreateInput{
		Title:       "Daily every 2 days",
		Description: "Seeded",
		Status:      taskdomain.StatusNew,
		RecurrenceType: recurrence.DailyEveryN,
		RecurrencePayload: recurrence.RecurrencePayload{
			IntervalDays: 2,
		},
		RecurrenceStartDate: &today,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, err = uc.Create(ctx, task.CreateInput{
		Title:       "Monthly day 15",
		Description: "Seeded",
		Status:      taskdomain.StatusNew,
		RecurrenceType: recurrence.MonthlyDay,
		RecurrencePayload: recurrence.RecurrencePayload{
			MonthDay: 15,
		},
		RecurrenceStartDate: &today,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tomorrow := common.NewDate(now.Add(24 * time.Hour))
	_, err = uc.Create(ctx, task.CreateInput{
		Title:       "Specific dates",
		Description: "Seeded",
		Status:      taskdomain.StatusNew,
		RecurrenceType: recurrence.SpecificDates,
		RecurrencePayload: recurrence.RecurrencePayload{
			SpecificDates: []common.DateType{today, tomorrow},
		},
		RecurrenceStartDate: &today,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, err = uc.Create(ctx, task.CreateInput{
		Title:       "Even days",
		Description: "Seeded",
		Status:      taskdomain.StatusNew,
		RecurrenceType: recurrence.MonthParity,
		RecurrencePayload: recurrence.RecurrencePayload{
			Parity: recurrence.Even,
		},
		RecurrenceStartDate: &today,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
