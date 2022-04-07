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

func (req OrderRequest) PreparationTime() time.Duration {
	sum := 0
	for _, topping := range req.Toppings {
		sum += len(topping)
	}
	return time.Duration(sum) * time.Second
}

// TODO: Should be separate events (probs with an event wrapper). Will do later.
type OrderResponse struct {
	OrderId       uint64        `json:"id"`
	Status        k.OrderStatus `json:"status"`
	FailureReason string        `json:"reason,omitempty"`
}

type OrderService interface {
	ProcessOrder(ctx context.Context, req OrderRequest) (OrderResponse, error)
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

// I'm not happy with the return type.
// The OrderResponse should be sent to a different topic depending upon whether error is nil or not. Can we improve this?
// Can we return different event types and switch between topic based on the type of the event?
func (svc orderService) ProcessOrder(ctx context.Context, req OrderRequest) (OrderResponse, error) {
	var (
		tx  *sql.Tx
		err error
	)

	log.Printf("Processing order %d: %q", req.OrderId, req.Toppings)

	tx, err = svc.stockDao.BeginTx()
	if err != nil {
		return OrderResponse{req.OrderId, k.OrderStatusFailed, err.Error()}, err
	}

	defer db.DeferRollback(tx, "ProcessOrder")

	// Decrease the stock
	var (
		stock k.Stock = k.Stock{}
		item  k.StockItem
	)

	for _, topping := range req.Toppings {
		if item, err = k.NewStockItem(topping, 1); err != nil {
			return OrderResponse{req.OrderId, k.OrderStatusFailed, err.Error()}, err
		}
		stock = append(stock, item)
	}

	err = svc.stockDao.Decrease(ctx, tx, stock)
	if err != nil {
		log.Printf("Error processing order %d. Reason: %q", req.OrderId, err.Error())
		return OrderResponse{req.OrderId, k.OrderStatusFailed, err.Error()}, err
	}

	if err = db.Commit(tx); err != nil {
		log.Printf("Error committing stock for order %d. Reason: %q", req.OrderId, err.Error())
		return OrderResponse{req.OrderId, k.OrderStatusFailed, err.Error()}, err
	}

	// Wait
	log.Printf("Preparing order %d. PreparationTime: %q", req.OrderId, req.PreparationTime().String())
	time.Sleep(req.PreparationTime())

	return OrderResponse{req.OrderId, k.OrderStatusReady, ""}, nil
}
