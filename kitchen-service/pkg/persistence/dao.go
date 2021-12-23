package persistence

import (
	"context"
	"database/sql"

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
