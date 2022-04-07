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
	BeginTx() (*sql.Tx, error)
	MustBeginTx() *sql.Tx

	Increase(ctx context.Context, tx *sql.Tx, stock k.Stock) error
	Decrease(ctx context.Context, tx *sql.Tx, decrease k.Stock) error
	Get(ctx context.Context, tx *sql.Tx) (k.Stock, error)
}

func DeferRollback(tx *sql.Tx, reference string) {
	if tx == nil {
		return
	}
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		log.Printf("failed to rollback transaction with reference %q. Reason: %s", reference, err)
	}
}

func Commit(tx *sql.Tx) error {
	if tx == nil {
		log.Fatal("Commit should not be passed a nil transaction")
	}
	if err := tx.Commit(); err != nil && err != sql.ErrTxDone {
		return k.NewSystemError("failed to save changes", err)
	}
	return nil
}
