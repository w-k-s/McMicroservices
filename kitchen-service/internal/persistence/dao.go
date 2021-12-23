package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff"
	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
	dao "github.com/w-k-s/McMicroservices/kitchen-service/pkg/persistence"
)

type defaultDao struct {
	db *sql.DB
}

func MustOpenDao(driverName, dataSourceName string) *defaultDao {
	var (
		db  *sql.DB
		err error
	)
	if db, err = sql.Open(driverName, dataSourceName); err != nil {
		log.Fatalf("Failed to connect to data source: %q with driver driver: %q. Reason: %s", dataSourceName, driverName, err)
	}
	return &defaultDao{db}
}

func (d defaultDao) NewStockTx(ctx context.Context) (dao.StockTx, error) {
	var (
		tx  *sql.Tx
		err error
	)
	if tx, err = d.db.BeginTx(ctx, nil); err != nil {
		return stockTx{}, err
	}
	return stockTx{tx}, nil
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
