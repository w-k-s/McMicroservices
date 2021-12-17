package services

import (
	"context"
	"fmt"
	"log"

	db "github.com/w-k-s/McMicroservices/kitchen-service/pkg/persistence"
)

type StockResponse struct {
	Stock string `json:"stock"`
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
	return StockResponse{"Hello World"}, nil
}
