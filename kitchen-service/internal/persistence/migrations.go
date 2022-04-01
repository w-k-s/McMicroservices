package persistence

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	"github.com/w-k-s/McMicroservices/kitchen-service/log"
)

func RunMigrations(db *sql.DB, dbConfig cfg.DBConfig) error {
	driverName := dbConfig.DriverName()
	migrationsDirectory := dbConfig.MigrationDirectory()

	if len(migrationsDirectory) == 0 {
		return fmt.Errorf("invalid migrations directory: '%s'. Must be an absolute path", migrationsDirectory)
	}

	var (
		driver     database.Driver
		migrations *migrate.Migrate
		err        error
	)

	if driver, err = postgres.WithInstance(db, &postgres.Config{
		DatabaseName: dbConfig.Name(),
		SchemaName:   dbConfig.Schema(),
	}); err != nil {
		return fmt.Errorf("failed to create instance of psql driver. Reason: %w", err)
	}

	if migrations, err = migrate.NewWithDatabaseInstance(migrationsDirectory, driverName, driver); err != nil {
		return fmt.Errorf("failed to load migrations from %s. Reason: %w", migrationsDirectory, err)
	}

	if err = migrations.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations from %s. Reason: %w", migrationsDirectory, err)
	}

	return nil
}

func MustRunMigrations(pool *sql.DB, dbConfig cfg.DBConfig) {
	if err := RunMigrations(pool, dbConfig); err != nil {
		log.Fatalf("Failed to run migrations. Reason: %s", err)
	}
	log.Print("Migrations applied")
}
