package persistence

import (
	"context"
	"database/sql"
	"fmt"

	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
)

type stockTx struct {
	*sql.Tx
}

func (s stockTx) Increase(ctx context.Context, stock k.Stock) error {
	return fmt.Errorf("Not implemented")
}

func (s stockTx) Decrease(ctx context.Context, decrease k.Stock) error {
	return fmt.Errorf("Not implemented")
}

func (s stockTx) Get(ctx context.Context) (k.Stock, error) {
	return []k.StockItem{}, fmt.Errorf("Not implemented")
}
