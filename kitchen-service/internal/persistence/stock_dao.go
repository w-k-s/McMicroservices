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

func (s defaultStockDao) Increase(ctx context.Context, tx *sql.Tx, stock k.Stock) k.Error {
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
			return k.NewError(k.ErrDatabaseState, fmt.Sprintf("Failed to increase stock of %q, Reason: %q", item.Name(), err.Error()), err)
		}
	}

	return nil
}

func (s defaultStockDao) Decrease(ctx context.Context, tx *sql.Tx, stock k.Stock) k.Error {
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
			return k.NewError(k.ErrDatabaseState, fmt.Sprintf("Failed to update stock of %q", item.Name()), err)
		}
		if rowsAffected, err = res.RowsAffected(); err != nil {
			return k.NewError(k.ErrDatabaseState, "Failed to get result of stock update", err)
		}
		if rowsAffected == 0 {
			return k.NewError(k.ErrInsufficientStock, fmt.Sprintf("Insufficient stock of %q", item.Name()), nil)
		}
	}
	return nil
}

func (s defaultStockDao) Get(ctx context.Context, tx *sql.Tx) (k.Stock, k.Error) {
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
		return nil, k.NewError(k.ErrDatabaseState, "Failed to load stock", err)
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
