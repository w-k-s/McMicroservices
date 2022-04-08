package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
	dao "github.com/w-k-s/McMicroservices/kitchen-service/pkg/persistence"
)

type defaultStockDao struct {
	*RootDao
}

func MustOpenStockDao(pool *sql.DB) dao.StockDao {
	if pool == nil {
		log.Fatalf("database is null")
	}
	return &defaultStockDao{&RootDao{pool}}
}

func (s *defaultStockDao) BeginTx() (dao.StockTx, error) {
	return StockTx(s.pool.Begin())
}

func StockTx(tx *sql.Tx, err error) (dao.StockTx,error){
	if err != nil{
		return nil,k.NewSystemError("failed to begin transaction",err)
	}
	return defaultStockTx{tx}, nil
}

type defaultStockTx struct {
	*sql.Tx
}

func (tx defaultStockTx) Increase(ctx context.Context, stock k.Stock) error {
	var err error

	for _, item := range stock {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO 
				kitchen.stock (item_name, units) 
			VALUES 
				($1,$2) 
			ON CONFLICT 
				ON CONSTRAINT uq_stock_name 
			DO UPDATE SET 
				units = (
					SELECT 
						units+$2
					FROM 
						kitchen.stock 
					WHERE 
						LOWER(item_name) = LOWER($1)
				)`,
			item.Name(),
			item.Units(),
		)

		if err != nil {
			return k.NewSystemError(fmt.Sprintf("Failed to increase stock of %q", item.Name()), err)
		}
	}

	return nil
}

func (tx defaultStockTx) Decrease(ctx context.Context, stock k.Stock) error {
	var (
		res          sql.Result
		rowsAffected int64
		err          error
	)

	for _, item := range stock {
		res, err = tx.ExecContext(
			ctx,
			`UPDATE 
				kitchen.stock 
			SET 
				units = units - $2
			WHERE 
				item_name = $1
			AND 
				units >= $2`,
			item.Name(),
			item.Units(),
		)

		if err != nil {
			return k.NewSystemError(fmt.Sprintf("failed to update stock of %q", item.Name()), err)
		}
		if rowsAffected, err = res.RowsAffected(); err != nil {
			return k.NewSystemError("failed to get result of stock update", err)
		}
		if rowsAffected == 0 {
			return k.InvalidError{Cause: fmt.Errorf("insufficient stock of %q", item.Name())}
		}
	}
	return nil
}

func (tx defaultStockTx) Get(ctx context.Context) (k.Stock, error) {
	var (
		rows *sql.Rows
		err  error
	)

	rows, err = tx.QueryContext(
		ctx,
		`SELECT 
			s.item_name,
			s.units
		FROM 
			kitchen.stock s`,
	)
	if err != nil {
		log.Printf("Failed to load stock. Reason: %q\n", err)
		return nil, k.NewSystemError("Failed to load stock", err)
	}
	defer rows.Close()

	items := make([]k.StockItem, 0)
	for rows.Next() {
		var (
			name  string
			count uint
		)

		if err = rows.Scan(&name, &count); err != nil {
			log.Printf("Error processing stock item %q. Reason: %s", name, err)
			continue
		}

		var item k.StockItem
		if item, err = k.NewStockItem(name, count); err != nil {
			log.Printf("Error creating stock item with name: %q,  units: %d from database. Reason: %q", name, count, err)
			continue
		}

		items = append(items, item)
	}

	return items, nil
}
