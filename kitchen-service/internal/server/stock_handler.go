package server

import (
	"net/http"

	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
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

	s.MustEncodeJson(w, resp, http.StatusCreated)
}
