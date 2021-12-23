package persistence

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff"
	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
)

type RootDao struct {
	db *sql.DB
}

func (r *RootDao) BeginTx() (*sql.Tx, k.Error) {
	var (
		tx  *sql.Tx
		err error
	)
	if tx, err = r.db.Begin(); err != nil {
		return nil, k.NewError(k.ErrDatabaseState, "Failed to begin transaction", err)
	}
	return tx, nil
}

func (r *RootDao) MustBeginTx() *sql.Tx {
	var (
		tx  *sql.Tx
		err error
	)

	if tx, err = r.db.Begin(); err != nil {
		log.Fatalf("Failed to begin transaction. Reason: %s", err)
	}
	return tx
}

func PingWithBackOff(db *sql.DB) error {
	var ping backoff.Operation = func() error {
		err := db.Ping()
		if err != nil {
			log.Printf("DB is not ready...backing off...: %s", err)
			return err
		}
		return nil
	}

	exponentialBackoff := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         time.Duration(50) * time.Millisecond,
		MaxElapsedTime:      time.Duration(300) * time.Millisecond,
		Clock:               backoff.SystemClock,
	}
	exponentialBackoff.Reset()

	var err error
	if err = backoff.Retry(ping, exponentialBackoff); err != nil {
		return fmt.Errorf("failed to connect to database after multiple retries. Reason: %w", err)
	}
	return nil
}

func DeferRollback(tx *sql.Tx, reference string) {
	if tx == nil {
		return
	}
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		log.Printf("failed to rollback transaction with reference %q. Reason: %s", reference, err)
	}
}

func Commit(tx *sql.Tx) k.Error {
	if tx == nil {
		log.Fatal("Commit should not be passed a nil transaction")
	}
	if err := tx.Commit(); err != nil && err != sql.ErrTxDone {
		return k.NewError(k.ErrDatabaseConnectivity, "Failed to save changes", err)
	}
	return nil
}
