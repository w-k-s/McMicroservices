package persistence

import (
	"context"

	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
)

type Dao interface {
	NewStockTx(ctx context.Context) (StockTx, error)
}

type StockTx interface {
	Commit() error
	Rollback() error

	Increase(ctx context.Context, stock k.Stock) error
	Decrease(ctx context.Context, decrease k.Stock) error
	Get(ctx context.Context) (k.Stock, error)
}
