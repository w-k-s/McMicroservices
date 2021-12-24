package persistence

import (
	"context"
	"database/sql"
	"log"

	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
)

type Dao interface {
	BeginTx() (*sql.Tx, error)
	MustBeginTx() *sql.Tx
}

type StockDao interface {
	BeginTx() (*sql.Tx, k.Error)
	MustBeginTx() *sql.Tx

	Close() error

	Increase(ctx context.Context, tx *sql.Tx, stock k.Stock) k.Error
	Decrease(ctx context.Context, tx *sql.Tx, decrease k.Stock) k.Error
	Get(ctx context.Context, tx *sql.Tx) (k.Stock, k.Error)
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
