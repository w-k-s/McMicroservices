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

type StockService interface {
	GetStock(ctx context.Context) (StockResponse, error)
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

func (svc stockService) GetStock(ctx context.Context) (StockResponse, error) {
	var (
		stock k.Stock
		tx    *sql.Tx
		err   error
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
