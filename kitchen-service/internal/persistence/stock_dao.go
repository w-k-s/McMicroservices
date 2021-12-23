package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
)

type stockTx struct {
	*sql.Tx
}

func (s stockTx) Increase(ctx context.Context, stock k.Stock) k.Error {
	var err error
	defer DeferRollback(s.Tx, "Increase Stock")

	for _, item := range stock {
		_, err = s.ExecContext(
			ctx,
			`INSERT INTO 
				kitchen.stock (name, units) 
			VALUES 
				(?,?) 
			ON CONFLICT 
				ON CONSTRAINT uq_stock_name 
			DO UPDATE SET 
				units = (
					SELECT 
						units+?
					FROM 
						kitchen.stock 
					WHERE 
						name = ?
				);`,
			item.Name(),
			item.Units(),
			item.Units(),
			item.Name(),
		)

		if err != nil {
			return k.NewError(k.ErrDatabaseState, fmt.Sprintf("Failed to increase stock of %q", item.Name()), err)
		}
	}

	return Commit(s.Tx)
}

func (s stockTx) Decrease(ctx context.Context, stock k.Stock) k.Error {
	var err error
	defer DeferRollback(s.Tx, "Decrease Stock")

	for _, item := range stock {
		_, err = s.ExecContext(
			ctx,
			`UPDATE 
				kitchen.stock 
			SET 
				units = units - ?
			WHERE 
				name = ?;`,
			item.Units(),
			item.Name(),
		)

		if err != nil {
			return k.NewError(k.ErrInsufficientStock, fmt.Sprintf("Not enough stock of %q", item.Name()), err)
		}
	}
	return Commit(s.Tx)
}

func (s stockTx) Get(ctx context.Context) (k.Stock, k.Error) {
	var (
		rows *sql.Rows
		err  error
	)
	defer DeferRollback(s.Tx, "Decrease Stock")

	rows, err = s.QueryContext(
		ctx,
		`SELECT 
			s.name,
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
