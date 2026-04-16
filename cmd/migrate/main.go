package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_DSN is required")
		os.Exit(2)
	}

	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "migrations"
	}

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	m, err := migrate.New("file://"+dir, dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer func() {
		_, _ = m.Close()
	}()

	switch cmd {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "staircase":
		err = staircase(m)
	case "version":
		v, dirty, verr := m.Version()
		if verr != nil {
			if errors.Is(verr, migrate.ErrNilVersion) {
				fmt.Println("0 false")
				return
			}
			fmt.Fprintln(os.Stderr, verr)
			os.Exit(1)
		}
		fmt.Printf("%d %t\n", v, dirty)
		return
	default:
		fmt.Fprintln(os.Stderr, "unknown command")
		os.Exit(2)
	}

	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func staircase(m *migrate.Migrate) error {
	// Step up one-by-one. If the file-source driver fails due to
	// filesystem/path issues (common when working with paths containing
	// spaces/special chars), fall back to a full Up() to ensure migrations
	// are applied.
	for {
		err := m.Steps(1)
		if err == nil {
			continue
		}
		if errors.Is(err, migrate.ErrNoChange) {
			break
		}
		if isFileMissingErr(err) {
			if upErr := m.Up(); upErr != nil && !errors.Is(upErr, migrate.ErrNoChange) {
				return fmt.Errorf("staircase fallback Up failed: %w", upErr)
			}
			break
		}
		return err
	}

	// Step down one-by-one. If Steps(-1) fails similarly, fall back to Down()
	// to revert applied migrations.
	for {
		err := m.Steps(-1)
		if err == nil {
			continue
		}
		if errors.Is(err, migrate.ErrNoChange) {
			break
		}
		if isFileMissingErr(err) {
			if downErr := m.Down(); downErr != nil && !errors.Is(downErr, migrate.ErrNoChange) {
				return fmt.Errorf("staircase fallback Down failed: %w", downErr)
			}
			break
		}
		return err
	}

	return nil
}

func isFileMissingErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, os.ErrNotExist) {
		return true
	}
	// migrate/file driver sometimes returns an error string containing
	// "file does not exist" or similar messages; check common substrings.
	s := err.Error()
	if strings.Contains(s, "file does not exist") || strings.Contains(s, "no such file") || strings.Contains(s, "failed to read") {
		return true
	}
	return false
}
