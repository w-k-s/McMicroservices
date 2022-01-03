package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"

	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
	db "github.com/w-k-s/McMicroservices/kitchen-service/pkg/persistence"
)

type StockItemResponse struct {
	Name  string `json:"name"`
	Units uint   `json:"units"`
}

type StockResponse struct {
	Stock []StockItemResponse `json:"stock"`
}

type StockItemRequest struct {
	Name  string `json:"name"`
	Units uint   `json:"units"`
}

type StockRequest struct {
	Stock []StockItemRequest `json:"stock"`
}

type StockService interface {
	GetStock(ctx context.Context) (StockResponse, k.Error)
	ReceiveInventory(ctx context.Context, req StockRequest) k.Error
}

type stockService struct {
	stockDao db.StockDao
}

func NewStockService(stockDao db.StockDao) (StockService, error) {
	if stockDao == nil {
		return nil, fmt.Errorf("can not create account service. stockDao is nil")
	}

	return &stockService{
		stockDao: stockDao,
	}, nil
}

func MustStockService(stockDao db.StockDao) StockService {
	var (
		svc StockService
		err error
	)
	if svc, err = NewStockService(stockDao); err != nil {
		log.Fatalf(err.Error())
	}
	return svc
}

func (svc stockService) GetStock(ctx context.Context) (StockResponse, k.Error) {
	var (
		stock k.Stock
		tx    *sql.Tx
		err   k.Error
	)

	tx, err = svc.stockDao.BeginTx()
	if err != nil {
		return StockResponse{}, err
	}

	defer db.DeferRollback(tx, "GetStock")

	stock, err = svc.stockDao.Get(ctx, tx)
	if err != nil {
		return StockResponse{}, err
	}

	if err = db.Commit(tx); err != nil {
		return StockResponse{}, err
	}

	sort.Sort(stock)
	items := []StockItemResponse{}
	for _, item := range stock {
		items = append(items, StockItemResponse{item.Name(), item.Units()})
	}

	return StockResponse{items}, nil
}

func (svc stockService) ReceiveInventory(ctx context.Context, req StockRequest) k.Error {
	var (
		tx  *sql.Tx
		err k.Error
	)

	tx, err = svc.stockDao.BeginTx()
	if err != nil {
		return err
	}

	defer db.DeferRollback(tx, "ReceiveInventory")

	received := k.Stock{}
	for _, requestItem := range req.Stock {
		var stockItem k.StockItem
		if stockItem, err = k.NewStockItem(requestItem.Name, requestItem.Units); err != nil {
			return err
		}
		received = append(received, stockItem)
	}

	err = svc.stockDao.Increase(ctx, tx, received)
	if err != nil {
		return err
	}

	if err = db.Commit(tx); err != nil {
		return err
	}

	return nil
}
