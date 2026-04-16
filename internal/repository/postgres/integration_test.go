package postgres_test

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"

	"example.com/taskservice/internal/domain/common"
	"example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	postgresrepo "example.com/taskservice/internal/repository/postgres"
)

func TestMigrations_Staircase(t *testing.T) {
	baseDSN := os.Getenv("TEST_DATABASE_DSN")
	if baseDSN == "" {
		t.Skip("TEST_DATABASE_DSN is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, dbDSN, cleanup := createTestDatabase(t, ctx, baseDSN)
	defer cleanup()

	m := newMigrator(t, dbDSN)
	defer func() { _, _ = m.Close() }()

	if err := staircase(m); err != nil {
		t.Fatalf("staircase failed: %v", err)
	}

	conn, err := pgx.Connect(ctx, dbDSN)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx)

	var regclass *string
	if err := conn.QueryRow(ctx, "SELECT to_regclass('public.tasks')").Scan(&regclass); err != nil {
		t.Fatalf("check tasks existence: %v", err)
	}
	if regclass != nil {
		t.Fatalf("expected tasks table to be dropped after staircase, got %q", *regclass)
	}
}

func TestRepository_ListByDate(t *testing.T) {
	baseDSN := os.Getenv("TEST_DATABASE_DSN")
	if baseDSN == "" {
		t.Skip("TEST_DATABASE_DSN is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, dbDSN, cleanup := createTestDatabase(t, ctx, baseDSN)
	defer cleanup()

	m := newMigrator(t, dbDSN)
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			t.Fatalf("migrate up: %v", err)
		}
	}
	_, _ = m.Close()

	pool, err := infrastructurepostgres.Open(ctx, dbDSN)
	if err != nil {
		t.Fatalf("pool open: %v", err)
	}
	defer pool.Close()

	repo := postgresrepo.New(pool)

	baseDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	start := common.NewDate(baseDate)

	now := time.Now().UTC()
	create := func(title string, rt recurrence.RecurrenceType, payload recurrence.RecurrencePayload) {
		t.Helper()
		model := &taskdomain.Task{
			Title:               title,
			Description:         "",
			Status:              taskdomain.StatusNew,
			CreatedAt:           now,
			UpdatedAt:           now,
			RecurrenceType:      rt,
			RecurrencePayload:   payload,
			RecurrenceStartDate: &start,
		}
		if _, err := repo.Create(ctx, model); err != nil {
			t.Fatalf("create %s: %v", title, err)
		}
	}

	create("one-time", recurrence.NoneRecurrence, recurrence.RecurrencePayload{})
	create("daily2", recurrence.DailyEveryN, recurrence.RecurrencePayload{IntervalDays: 2})
	create("monthly15", recurrence.MonthlyDay, recurrence.RecurrencePayload{MonthDay: 15})
	create("parity-even", recurrence.MonthParity, recurrence.RecurrencePayload{Parity: recurrence.Even})
	create("specific", recurrence.SpecificDates, recurrence.RecurrencePayload{SpecificDates: []common.DateType{common.NewDate(baseDate.Add(48 * time.Hour))}})

	assertTitles := func(date time.Time, want map[string]bool) {
		t.Helper()
		tasks, err := repo.ListByDate(ctx, date)
		if err != nil {
			t.Fatalf("ListByDate: %v", err)
		}
		got := map[string]bool{}
		for i := range tasks {
			got[tasks[i].Title] = true
		}
		for title := range want {
			if !got[title] {
				t.Fatalf("%s: expected %q in result, got=%v", date.Format(common.DateLayout), title, got)
			}
		}
		if len(got) != len(want) {
			t.Fatalf("%s: unexpected task count, want=%v got=%v", date.Format(common.DateLayout), want, got)
		}
	}

	assertTitles(baseDate, map[string]bool{"one-time": true, "daily2": true})
	assertTitles(baseDate.Add(24*time.Hour), map[string]bool{"parity-even": true})
	assertTitles(baseDate.Add(48*time.Hour), map[string]bool{"daily2": true, "specific": true})
}

func staircase(m *migrate.Migrate) error {
	for {
		err := m.Steps(1)
		if err == nil {
			continue
		}
		if errors.Is(err, migrate.ErrNoChange) {
			break
		}
		return err
	}

	for {
		err := m.Steps(-1)
		if err == nil {
			continue
		}
		if errors.Is(err, migrate.ErrNoChange) {
			break
		}
		return err
	}

	return nil
}

func newMigrator(t *testing.T, dsn string) *migrate.Migrate {
	t.Helper()

	migrationsDir := filepath.Join(repoRoot(t), "migrations")
	u := url.URL{Path: migrationsDir}
	sourceURL := "file://" + u.EscapedPath()

	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		t.Fatalf("migrate.New: %v", err)
	}
	return m
}

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found")
		}
		dir = parent
	}
}

func createTestDatabase(t *testing.T, ctx context.Context, baseDSN string) (dbName string, dbDSN string, cleanup func()) {
	t.Helper()

	u, err := url.Parse(baseDSN)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}

	adminU := *u
	adminU.Path = "/postgres"
	adminDSN := adminU.String()

	conn, err := pgx.Connect(ctx, adminDSN)
	if err != nil {
		t.Fatalf("admin connect: %v", err)
	}
	defer conn.Close(ctx)

	dbName = "taskservice_test_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	if _, err := conn.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
		t.Fatalf("create database: %v", err)
	}

	dbU := *u
	dbU.Path = "/" + dbName
	dbDSN = dbU.String()

	cleanup = func() {
		ctx := context.Background()

		conn, err := pgx.Connect(ctx, adminDSN)
		if err != nil {
			t.Logf("cleanup connect: %v", err)
			return
		}
		defer conn.Close(ctx)

		if _, err := conn.Exec(ctx, "DROP DATABASE "+dbName+" WITH (FORCE)"); err == nil {
			return
		}

		_, _ = conn.Exec(ctx, "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1", dbName)
		if _, err := conn.Exec(ctx, "DROP DATABASE "+dbName); err != nil {
			t.Logf("cleanup drop database: %v", err)
		}
	}

	return dbName, dbDSN, cleanup
}
