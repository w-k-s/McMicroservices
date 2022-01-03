package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	msg "github.com/w-k-s/McMicroservices/kitchen-service/internal/messages"
	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
	"schneider.vip/problem"
)

const (
	InventoryDelivery string            = "inventory_delivery"
	InventoryRejected msg.ResponseTopic = "inventory_rejected"
)

type stockHandler struct {
	Handler
	stockSvc svc.StockService
}

func NewStockHandler(stockSvc svc.StockService) stockHandler {
	return stockHandler{
		Handler{},
		stockSvc,
	}
}

func (s stockHandler) GetStock(w http.ResponseWriter, req *http.Request) {

	var (
		resp svc.StockResponse
		err  error
	)

	if resp, err = s.stockSvc.GetStock(req.Context()); err != nil {
		s.MustEncodeProblem(w, req, err)
		return
	}

	s.MustEncodeJson(w, resp, http.StatusOK)
}

func (s stockHandler) ReceiveInventory(ctx context.Context, request msg.RequestBody) (msg.ResponseTopic, msg.ResponseBody) {

	decoder := json.NewDecoder(request.Reader())
	decoder.UseNumber()

	var (
		receiveInventoryRequest svc.StockRequest
		err                     error
		kitchenError            k.Error
	)
	if err = decoder.Decode(receiveInventoryRequest); err != nil {
		return InventoryRejected, problem.New(
			problem.Type(fmt.Sprintf("/api/v1/problems/%d", k.ErrUnmarshalling)),
			problem.Status(k.ErrUnmarshalling.Status()),
			problem.Title(k.ErrUnmarshalling.Name()),
			problem.Detail(err.Error()),
		).JSON()
	}

	if kitchenError = s.stockSvc.ReceiveInventory(ctx, receiveInventoryRequest); kitchenError != nil {
		return InventoryRejected, problem.New(
			problem.Type(fmt.Sprintf("/api/v1/problems/%d", kitchenError.Code())),
			problem.Status(kitchenError.Code().Status()),
			problem.Title(kitchenError.Code().Name()),
			problem.Detail(kitchenError.Error()),
		).JSON()
	}

	return "", msg.ResponseBody([]byte{})
}
