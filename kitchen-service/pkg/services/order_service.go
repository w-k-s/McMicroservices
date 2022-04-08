package services

import (
	"context"
	"time"

	"github.com/w-k-s/McMicroservices/kitchen-service/log"

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

func MustOrderService(stockDao db.StockDao) OrderService {
	if stockDao == nil {
		log.Fatal("can not create account service. stockDao is nil")
	}

	return &orderService{
		stockDao: stockDao,
	}
}

// I'm not happy with the return type.
// The OrderResponse should be sent to a different topic depending upon whether error is nil or not. Can we improve this?
// Can we return different event types and switch between topic based on the type of the event?
func (svc orderService) ProcessOrder(ctx context.Context, req OrderRequest) (OrderResponse, error) {
	var (
		tx  db.StockTx
		err error
	)

	log.InfoCtx(ctx).
		UInt64("orderId", req.OrderId).
		Struct("toppings", req.Toppings).
		Msg("Processing order")

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

	err = tx.Decrease(ctx, stock)
	if err != nil {
		log.ErrCtx(ctx, err).
			UInt64("orderId", req.OrderId).
			Msg("Error processing order")
		return OrderResponse{req.OrderId, k.OrderStatusFailed, err.Error()}, err
	}

	if err = db.Commit(tx); err != nil {
		log.ErrCtx(ctx, err).
			UInt64("orderId", req.OrderId).
			Msg("Error committing stock")
		return OrderResponse{req.OrderId, k.OrderStatusFailed, err.Error()}, err
	}

	// Wait
	log.InfoCtx(ctx).
		UInt64("orderId", req.OrderId).
		Duration("PreparationTime", req.PreparationTime()).
		Msg("Preparing order")
	time.Sleep(req.PreparationTime())

	return OrderResponse{req.OrderId, k.OrderStatusReady, ""}, nil
}
