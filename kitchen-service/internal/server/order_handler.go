package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"schneider.vip/problem"

	msg "github.com/w-k-s/McMicroservices/kitchen-service/internal/messages"
	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
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
		responseJson  []byte
		err           error
		kitchenError  k.Error
	)
	if err = decoder.Decode(orderRequest); err != nil {
		return OrderFailed, problem.New(
			problem.Type(fmt.Sprintf("/api/v1/problems/%d", k.ErrUnmarshalling)),
			problem.Status(k.ErrUnmarshalling.Status()),
			problem.Title(k.ErrUnmarshalling.Name()),
			problem.Detail(err.Error()),
		).JSON()
	}

	if orderResponse, kitchenError = oh.orderService.ProcessOrder(ctx, orderRequest); kitchenError != nil {
		return OrderFailed, problem.New(
			problem.Type(fmt.Sprintf("/api/v1/problems/%d", kitchenError.Code())),
			problem.Status(kitchenError.Code().Status()),
			problem.Title(kitchenError.Code().Name()),
			problem.Detail(kitchenError.Error()),
		).JSON()
	}

	if responseJson, err = json.Marshal(orderResponse); err != nil {
		log.Printf("UNEXPECTED: Failed to marshal orderResponse %v", orderResponse)
	}

	return OrderReady, msg.ResponseBody(responseJson)
}
