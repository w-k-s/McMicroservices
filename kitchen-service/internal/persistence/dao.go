package persistence

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	"github.com/w-k-s/McMicroservices/kitchen-service/log"
)

type RootDao struct {
	pool *sql.DB
}

func OpenPool(dbConfig cfg.DBConfig) (*sql.DB, error) {
	var (
		db  *sql.DB
		err error
	)

	if db, err = sql.Open(dbConfig.DriverName(), dbConfig.ConnectionString()); err != nil {
		return nil, fmt.Errorf("failed to open connection. Reason: %w", err)
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(3) // Required, otherwise pinging will result in EOF
	db.SetMaxOpenConns(3)

	if err = PingWithBackOff(db); err != nil {
		return nil, fmt.Errorf("failed to ping database. Reason: %w", err)
	}
	return db, nil
}

func Must(db *sql.DB, err error) *sql.DB {
	if err != nil {
		log.Fatalf("Failed to open database connection pool. Reason: %s", err)
	}
	return db
}

func PingWithBackOff(db *sql.DB) error {
	var ping backoff.Operation = func() error {
		err := db.Ping()
		if err != nil {
			log.WithFields(map[string]interface{}{
				"reason": err,
			}).Print("DB is not ready...backing off...")
			return err
		}
		return nil
	}

	exponentialBackoff := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         time.Duration(100) * time.Millisecond,
		MaxElapsedTime:      time.Duration(5) * time.Second,
		Clock:               backoff.SystemClock,
	}
	exponentialBackoff.Reset()

	var err error
	if err = backoff.Retry(ping, exponentialBackoff); err != nil {
		return fmt.Errorf("failed to connect to database after multiple retries. Reason: %w", err)
	}
	return nil
}
