package services

import (
	"context"
	"fmt"
	"log"

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
	dao db.Dao
}

func NewStockService(dao db.Dao) (StockService, error) {
	if dao == nil {
		return nil, fmt.Errorf("can not create account service. dao is nil")
	}

	return &stockService{
		dao: dao,
	}, nil
}

func MustStockService(dao db.Dao) StockService {
	var (
		svc StockService
		err error
	)
	if svc, err = NewStockService(dao); err != nil {
		log.Fatalf(err.Error())
	}
	return svc
}

func (svc stockService) GetStock(ctx context.Context) (StockResponse, error) {
	var (
		tx    db.StockTx
		stock k.Stock
		err   error
	)

	tx, err = svc.dao.NewStockTx(ctx)
	if err != nil {
		return StockResponse{}, err
	}

	stock, err = tx.Get(ctx)
	if err != nil {
		return StockResponse{}, err
	}

	items := []StockItemResponse{}
	for _, item := range stock {
		items = append(items, StockItemResponse{item.Name(), item.Units()})
	}

	return StockResponse{items}, nil
}
