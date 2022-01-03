package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
	db "github.com/w-k-s/McMicroservices/kitchen-service/pkg/persistence"
)

type OrderRequest struct {
	OrderId  uint64   `json:"id"`
	Toppings []string `json:"toppings"`
}

type OrderStatus string

const (
	OrderStatusPreparing OrderStatus = "PREPARING"
	OrderStatusReady     OrderStatus = "READY"
	OrderStatusFailed    OrderStatus = "FAILED"
)

func (req OrderRequest) PreparationTime() time.Duration {
	sum := 0
	for _, topping := range req.Toppings {
		sum += len(topping)
	}
	return time.Duration(sum) * time.Second
}

type OrderResponse struct {
	OrderId uint64      `json:"id"`
	Status  OrderStatus `json:"status"`
}

type OrderService interface {
	ProcessOrder(ctx context.Context, req OrderRequest) (OrderResponse, k.Error)
}

type orderService struct {
	stockDao db.StockDao
}

func NewOrderService(stockDao db.StockDao) (OrderService, error) {
	if stockDao == nil {
		return nil, fmt.Errorf("can not create account service. stockDao is nil")
	}

	return &orderService{
		stockDao: stockDao,
	}, nil
}

func MustOrderService(stockDao db.StockDao) OrderService {
	var (
		svc OrderService
		err error
	)
	if svc, err = NewOrderService(stockDao); err != nil {
		log.Fatalf(err.Error())
	}
	return svc
}

func (svc orderService) ProcessOrder(ctx context.Context, req OrderRequest) (OrderResponse, k.Error) {
	var (
		tx  *sql.Tx
		err k.Error
	)

	log.Printf("Processing order %d: %q", req.OrderId, req.Toppings)

	tx, err = svc.stockDao.BeginTx()
	if err != nil {
		return OrderResponse{}, err
	}

	defer db.DeferRollback(tx, "ProcessOrder")

	// Decrease the stock
	var (
		stock k.Stock = k.Stock{}
		item  k.StockItem
	)

	for _, topping := range req.Toppings {
		if item, err = k.NewStockItem(topping, 1); err != nil {
			return OrderResponse{req.OrderId, OrderStatusFailed}, err
		}
		stock = append(stock, item)
	}

	err = svc.stockDao.Decrease(ctx, tx, stock)
	if err != nil {
		log.Printf("Error processing order %d. Reason: %q", req.OrderId, err.Error())
		return OrderResponse{req.OrderId, OrderStatusFailed}, err
	}

	if err = db.Commit(tx); err != nil {
		return OrderResponse{req.OrderId, OrderStatusFailed}, err
	}

	// Wait
	log.Printf("Preparing order %d. PreparationTime: %q", req.OrderId, req.PreparationTime().String())
	time.Sleep(req.PreparationTime())

	return OrderResponse{req.OrderId, OrderStatusReady}, nil
}
