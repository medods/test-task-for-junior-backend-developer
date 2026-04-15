package postgres

import (
	"context"
	"errors"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskScheduleRepository struct {
	pool *pgxpool.Pool
}

func NewTaskScheduleRepository(pool *pgxpool.Pool) *TaskScheduleRepository {
	return &TaskScheduleRepository{pool: pool}
}

func (r *TaskScheduleRepository) Create(ctx context.Context, schedule *scheduledomain.TaskSchedule) (*scheduledomain.TaskSchedule, error) {
	ruleType, payload, err := scheduledomain.EncodeRule(schedule.Rule)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO task_schedules (title, description, rule_type, rule_payload, start_date, end_date, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7, $8, $9)
		RETURNING id, title, description, rule_type, rule_payload, start_date, end_date, active, created_at, updated_at
	`

	now := time.Now().UTC()
	if !schedule.CreatedAt.IsZero() {
		now = schedule.CreatedAt
	}
	row := infrastructurepostgres.DBFromContext(ctx, r.pool).QueryRow(
		ctx,
		query,
		schedule.Title,
		schedule.Description,
		ruleType,
		string(payload),
		schedule.StartDate,
		schedule.EndDate,
		schedule.Active,
		now,
		now,
	)

	return scanSchedule(row)
}

func (r *TaskScheduleRepository) GetByID(ctx context.Context, id int64) (*scheduledomain.TaskSchedule, error) {
	const query = `
		SELECT id, title, description, rule_type, rule_payload, start_date, end_date, active, created_at, updated_at
		FROM task_schedules
		WHERE id = $1
	`

	row := infrastructurepostgres.DBFromContext(ctx, r.pool).QueryRow(ctx, query, id)
	schedule, err := scanSchedule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, scheduledomain.ErrNotFound
		}

		return nil, err
	}

	return schedule, nil
}

func (r *TaskScheduleRepository) List(ctx context.Context) ([]scheduledomain.TaskSchedule, error) {
	const query = `
		SELECT id, title, description, rule_type, rule_payload, start_date, end_date, active, created_at, updated_at
		FROM task_schedules
		ORDER BY id DESC
	`

	rows, err := infrastructurepostgres.DBFromContext(ctx, r.pool).Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]scheduledomain.TaskSchedule, 0)
	for rows.Next() {
		schedule, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}

		result = append(result, *schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *TaskScheduleRepository) Update(ctx context.Context, schedule *scheduledomain.TaskSchedule) (*scheduledomain.TaskSchedule, error) {
	ruleType, payload, err := scheduledomain.EncodeRule(schedule.Rule)
	if err != nil {
		return nil, err
	}

	const query = `
		UPDATE task_schedules
		SET title = $1,
			description = $2,
			rule_type = $3,
			rule_payload = $4::jsonb,
			start_date = $5,
			end_date = $6,
			active = $7,
			updated_at = $8
		WHERE id = $9
		RETURNING id, title, description, rule_type, rule_payload, start_date, end_date, active, created_at, updated_at
	`

	row := infrastructurepostgres.DBFromContext(ctx, r.pool).QueryRow(
		ctx,
		query,
		schedule.Title,
		schedule.Description,
		ruleType,
		string(payload),
		schedule.StartDate,
		schedule.EndDate,
		schedule.Active,
		time.Now().UTC(),
		schedule.ID,
	)

	updated, err := scanSchedule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, scheduledomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *TaskScheduleRepository) Deactivate(ctx context.Context, id int64, updatedAt time.Time) error {
	const query = `
		UPDATE task_schedules
		SET active = FALSE,
			updated_at = $2
		WHERE id = $1
	`

	result, err := infrastructurepostgres.DBFromContext(ctx, r.pool).Exec(ctx, query, id, updatedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return scheduledomain.ErrNotFound
	}

	return nil
}

func (r *TaskScheduleRepository) ListActiveBetween(ctx context.Context, from time.Time, to time.Time) ([]scheduledomain.TaskSchedule, error) {
	const query = `
		SELECT id, title, description, rule_type, rule_payload, start_date, end_date, active, created_at, updated_at
		FROM task_schedules
		WHERE active = TRUE
		  AND start_date <= $2
		  AND (end_date IS NULL OR end_date >= $1)
		ORDER BY id
	`

	rows, err := infrastructurepostgres.DBFromContext(ctx, r.pool).Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]scheduledomain.TaskSchedule, 0)
	for rows.Next() {
		schedule, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}

		result = append(result, *schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

type scheduleScanner interface {
	Scan(dest ...any) error
}

func scanSchedule(scanner scheduleScanner) (*scheduledomain.TaskSchedule, error) {
	var (
		model       scheduledomain.TaskSchedule
		ruleType    string
		rulePayload []byte
		endDate     pgtype.Date
	)

	if err := scanner.Scan(
		&model.ID,
		&model.Title,
		&model.Description,
		&ruleType,
		&rulePayload,
		&model.StartDate,
		&endDate,
		&model.Active,
		&model.CreatedAt,
		&model.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if len(rulePayload) == 0 {
		rulePayload = []byte("{}")
	}

	model.StartDate = scheduledomain.NormalizeDate(model.StartDate)
	if endDate.Valid {
		value := scheduledomain.NormalizeDate(endDate.Time)
		model.EndDate = &value
	}

	rule, err := scheduledomain.DecodeRule(scheduledomain.RuleType(ruleType), rulePayload)
	if err != nil {
		return nil, err
	}

	model.Rule = rule
	return &model, nil
}
