package persistence

import (
	"context"
	"database/sql"
	"log"

	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
)

type StockDao interface {
	BeginTx() (StockTx, error)
}

type Tx interface{
	Commit() error
	Rollback() error
}

type StockTx interface {
	Commit() error
	Rollback() error

	Increase(ctx context.Context, stock k.Stock) error
	Decrease(ctx context.Context, decrease k.Stock) error
	Get(ctx context.Context) (k.Stock, error)
}

func DeferRollback(tx Tx, reference string) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		log.Printf("failed to rollback transaction with reference %q. Reason: %s", reference, err)
	}
}

func Commit(tx Tx) error {
	if err := tx.Commit(); err != nil && err != sql.ErrTxDone {
		return k.NewSystemError("failed to save changes", err)
	}
	return nil
}
