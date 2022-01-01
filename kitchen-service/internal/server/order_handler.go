package server

import (
	msg "github.com/w-k-s/McMicroservices/kitchen-service/internal/messages"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

const (
	CreateOrder string = "order_created"
	OrderReady  string = "order_ready"
	OrderFailed string = "order_failed"
)

type OrderHandler interface {
	HandleOrderMessage(request msg.RequestBody) (msg.ResponseTopic, msg.ResponseBody)
}

type orderHandler struct {
	orderService svc.StockService
}

func NewOrderHandler(
	orderService svc.StockService,
) OrderHandler {

	return &orderHandler{
		orderService: orderService,
	}
}

func (oh orderHandler) HandleOrderMessage(request msg.RequestBody) (msg.ResponseTopic, msg.ResponseBody) {
	return msg.ResponseTopic("order_ready"), msg.ResponseBody([]byte("Hello World"))
}
