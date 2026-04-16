package main

import (
	"errors"
	"fmt"
	"os"

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
