package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/shared/dateutil"
)

type RecurrenceRepository struct {
	pool *pgxpool.Pool
}

func NewRecurrenceRepository(pool *pgxpool.Pool) *RecurrenceRepository {
	return &RecurrenceRepository{pool: pool}
}

func (r *RecurrenceRepository) Create(ctx context.Context, rule *recurrencedomain.Rule) (*recurrencedomain.Rule, error) {
	const query = `
		INSERT INTO recurrence_rules (
			title,
			description,
			active,
			schedule_type,
			starts_on,
			ends_on,
			interval_days,
			day_of_month,
			day_parity,
			specific_dates,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, title, description, active, schedule_type, starts_on, ends_on, interval_days, day_of_month, day_parity, specific_dates, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		rule.Title,
		rule.Description,
		rule.Active,
		rule.ScheduleType,
		rule.StartsOn,
		rule.EndsOn,
		rule.EveryNDays,
		rule.DayOfMonth,
		rule.DayParity,
		normalizeDateSlice(rule.SpecificDates),
		time.Now().UTC(),
		time.Now().UTC(),
	)

	created, err := scanRule(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *RecurrenceRepository) GetByID(ctx context.Context, id int64) (*recurrencedomain.Rule, error) {
	const query = `
		SELECT id, title, description, active, schedule_type, starts_on, ends_on, interval_days, day_of_month, day_parity, specific_dates, created_at, updated_at
		FROM recurrence_rules
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	rule, err := scanRule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, recurrencedomain.ErrNotFound
		}

		return nil, err
	}

	return rule, nil
}

func (r *RecurrenceRepository) Update(ctx context.Context, rule *recurrencedomain.Rule) (*recurrencedomain.Rule, error) {
	const query = `
		UPDATE recurrence_rules
		SET title = $1,
			description = $2,
			active = $3,
			schedule_type = $4,
			starts_on = $5,
			ends_on = $6,
			interval_days = $7,
			day_of_month = $8,
			day_parity = $9,
			specific_dates = $10,
			updated_at = $11
		WHERE id = $12
		RETURNING id, title, description, active, schedule_type, starts_on, ends_on, interval_days, day_of_month, day_parity, specific_dates, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		rule.Title,
		rule.Description,
		rule.Active,
		rule.ScheduleType,
		rule.StartsOn,
		rule.EndsOn,
		rule.EveryNDays,
		rule.DayOfMonth,
		rule.DayParity,
		normalizeDateSlice(rule.SpecificDates),
		time.Now().UTC(),
		rule.ID,
	)

	updated, err := scanRule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, recurrencedomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *RecurrenceRepository) Delete(ctx context.Context, id int64, fromDate time.Time) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const deleteFutureTasksQuery = `
		DELETE FROM tasks
		WHERE recurrence_rule_id = $1
		  AND scheduled_for >= $2
		  AND status = $3
		  AND is_modified = FALSE
	`
	if _, err := tx.Exec(ctx, deleteFutureTasksQuery, id, dateutil.Normalize(fromDate), taskdomain.StatusNew); err != nil {
		return err
	}

	const deleteRuleQuery = `DELETE FROM recurrence_rules WHERE id = $1`
	result, err := tx.Exec(ctx, deleteRuleQuery, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return recurrencedomain.ErrNotFound
	}

	return tx.Commit(ctx)
}

func (r *RecurrenceRepository) List(ctx context.Context) ([]recurrencedomain.Rule, error) {
	const query = `
		SELECT id, title, description, active, schedule_type, starts_on, ends_on, interval_days, day_of_month, day_parity, specific_dates, created_at, updated_at
		FROM recurrence_rules
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return collectRules(rows)
}

func (r *RecurrenceRepository) ListActive(ctx context.Context) ([]recurrencedomain.Rule, error) {
	const query = `
		SELECT id, title, description, active, schedule_type, starts_on, ends_on, interval_days, day_of_month, day_parity, specific_dates, created_at, updated_at
		FROM recurrence_rules
		WHERE active = TRUE
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return collectRules(rows)
}

func (r *RecurrenceRepository) SyncGeneratedTasks(ctx context.Context, rule *recurrencedomain.Rule, fromDate time.Time, occurrences []time.Time) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const deleteQuery = `
		DELETE FROM tasks
		WHERE recurrence_rule_id = $1
		  AND scheduled_for >= $2
		  AND status = $3
		  AND is_modified = FALSE
	`
	if _, err := tx.Exec(ctx, deleteQuery, rule.ID, dateutil.Normalize(fromDate), taskdomain.StatusNew); err != nil {
		return err
	}

	if rule.Active {
		const insertQuery = `
			INSERT INTO tasks (
				title,
				description,
				status,
				scheduled_for,
				recurrence_rule_id,
				is_modified,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, $4, $5, FALSE, $6, $7)
			ON CONFLICT (recurrence_rule_id, scheduled_for) WHERE recurrence_rule_id IS NOT NULL
			DO UPDATE SET
				title = EXCLUDED.title,
				description = EXCLUDED.description,
				updated_at = EXCLUDED.updated_at
			WHERE tasks.status = $3 AND tasks.is_modified = FALSE
		`

		now := time.Now().UTC()
		for _, occurrence := range occurrences {
			if _, err := tx.Exec(
				ctx,
				insertQuery,
				rule.Title,
				rule.Description,
				taskdomain.StatusNew,
				dateutil.Normalize(occurrence),
				rule.ID,
				now,
				now,
			); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

type ruleScanner interface {
	Scan(dest ...any) error
}

func scanRule(scanner ruleScanner) (*recurrencedomain.Rule, error) {
	var (
		rule          recurrencedomain.Rule
		scheduleType  string
		endsOn        *time.Time
		dayParity     *string
		specificDates []time.Time
	)

	if err := scanner.Scan(
		&rule.ID,
		&rule.Title,
		&rule.Description,
		&rule.Active,
		&scheduleType,
		&rule.StartsOn,
		&endsOn,
		&rule.EveryNDays,
		&rule.DayOfMonth,
		&dayParity,
		&specificDates,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	); err != nil {
		return nil, err
	}

	rule.ScheduleType = recurrencedomain.ScheduleType(scheduleType)
	if endsOn != nil {
		normalized := dateutil.Normalize(*endsOn)
		rule.EndsOn = &normalized
	}
	if dayParity != nil {
		parity := recurrencedomain.DayParity(*dayParity)
		rule.DayParity = &parity
	}
	rule.StartsOn = dateutil.Normalize(rule.StartsOn)
	rule.SpecificDates = normalizeDateSlice(specificDates)

	return &rule, nil
}

func collectRules(rows pgx.Rows) ([]recurrencedomain.Rule, error) {
	rules := make([]recurrencedomain.Rule, 0)
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}

		rules = append(rules, *rule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func normalizeDateSlice(values []time.Time) []time.Time {
	if len(values) == 0 {
		return []time.Time{}
	}

	normalized := make([]time.Time, 0, len(values))
	for _, value := range values {
		normalized = append(normalized, dateutil.Normalize(value))
	}

	return normalized
}
