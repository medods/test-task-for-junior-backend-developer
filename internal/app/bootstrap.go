package app

import (
	"context"

	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	postgresrepo "example.com/taskservice/internal/repository/postgres"
	materializerusecase "example.com/taskservice/internal/usecase/materializer"
	scheduleusecase "example.com/taskservice/internal/usecase/schedule"
	taskusecase "example.com/taskservice/internal/usecase/task"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	Pool            *pgxpool.Pool
	TaskService     taskusecase.Usecase
	ScheduleService scheduleusecase.Usecase
	Materializer    *materializerusecase.Service
}

func NewContainer(ctx context.Context, cfg Config) (*Container, error) {
	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	taskRepo := postgresrepo.New(pool)
	scheduleRepo := postgresrepo.NewTaskScheduleRepository(pool)
	txManager := infrastructurepostgres.NewTxManager(pool)
	materializer := materializerusecase.NewService(scheduleRepo, taskRepo, cfg.Location, cfg.MaterializationWindowDays)
	taskService := taskusecase.NewService(taskRepo)
	scheduleService := scheduleusecase.NewService(scheduleRepo, txManager, materializer, cfg.Location, cfg.MaterializationWindowDays)

	return &Container{
		Pool:            pool,
		TaskService:     taskService,
		ScheduleService: scheduleService,
		Materializer:    materializer,
	}, nil
}
