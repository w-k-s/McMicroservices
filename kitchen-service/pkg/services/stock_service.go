package services

import (
	"context"
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
	GetStock(ctx context.Context) (StockResponse, error)
	ReceiveInventory(ctx context.Context, req StockRequest) error
}

type stockService struct {
	stockDao db.StockDao
}

func MustStockService(stockDao db.StockDao) StockService {
	if stockDao == nil {
		log.Fatal("can not create account service. stockDao is nil")
	}
	return &stockService{
		stockDao: stockDao,
	}
}

func (svc stockService) GetStock(ctx context.Context) (StockResponse, error) {
	var (
		stock k.Stock
		tx    db.StockTx
		err   error
	)

	tx, err = svc.stockDao.BeginTx()
	if err != nil {
		return StockResponse{}, err
	}

	defer db.DeferRollback(tx, "GetStock")

	stock, err = tx.Get(ctx)
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

func (svc stockService) ReceiveInventory(ctx context.Context, req StockRequest) error {
	var (
		tx  db.StockTx
		err error
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

	err = tx.Increase(ctx, received)
	if err != nil {
		return err
	}

	if err = db.Commit(tx); err != nil {
		return err
	}

	return nil
}
