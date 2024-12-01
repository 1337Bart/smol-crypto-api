package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var fs embed.FS

func RunMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create the postgres driver: %w", err)
	}

	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return fmt.Errorf("could not create the iofs driver: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs", d,
		"postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("could not create the migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("could not run up migrations: %w", err)
	}

	return nil
}
