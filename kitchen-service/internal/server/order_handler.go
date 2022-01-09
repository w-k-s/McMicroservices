package server

import (
	"context"
	"encoding/json"
	"log"

	msg "github.com/w-k-s/McMicroservices/kitchen-service/internal/messages"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

const (
	CreateOrder string            = "order_created"
	OrderReady  msg.ResponseTopic = "order_ready"
	OrderFailed msg.ResponseTopic = "order_failed"
)

type OrderHandler interface {
	HandleOrderMessage(ctx context.Context, request msg.RequestBody) (msg.ResponseTopic, msg.ResponseBody)
}

type orderHandler struct {
	Handler
	orderService svc.OrderService
}

func NewOrderHandler(
	orderService svc.OrderService,
) OrderHandler {

	return &orderHandler{
		orderService: orderService,
	}
}

func (oh orderHandler) HandleOrderMessage(ctx context.Context, request msg.RequestBody) (msg.ResponseTopic, msg.ResponseBody) {

	decoder := json.NewDecoder(request.Reader())
	decoder.UseNumber()

	var (
		orderRequest  svc.OrderRequest
		orderResponse svc.OrderResponse
		err           error
	)
	if err = decoder.Decode(&orderRequest); err != nil {
		// TODO: This should probably go in to a failed-to-process queue
		log.Printf("Failed to decode order request. Reason: %s", err)
		return "",[]byte{}
	}

	if orderResponse, err = oh.orderService.ProcessOrder(ctx, orderRequest); err != nil{
		return OrderFailed, oh.MustMarshal(json.Marshal(orderResponse))
	}
	return OrderReady, oh.MustMarshal(json.Marshal(orderResponse))
}
