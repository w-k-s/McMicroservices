package app

import (
	"bytes"
	"context"
	"net/http"

	"github.com/w-k-s/McMicroservices/kitchen-service/internal/broker"
	"github.com/w-k-s/McMicroservices/kitchen-service/log"

	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

func (a *App) GetStock() http.HandlerFunc {

	stockDao := db.MustOpenStockDao(a.pool)
	stockSvc := svc.MustStockService(stockDao)

	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := stockSvc.GetStock(r.Context())
		if err != nil {
			a.serde.MustSerailizeError(w, r, 0, err)
			return
		}
		a.serde.MustSerialize(w, r, http.StatusOK, resp)
	}
}

func (a *App) ReceiveInventory(ctx context.Context, e broker.Message) error {
	log.InfoCtx(ctx).Msg("Inventory Received...")

	var receiveInventoryRequest svc.StockRequest
	if err := a.decoder.Decode(bytes.NewReader(e.Content), &receiveInventoryRequest); err != nil {
		log.ErrCtx(ctx, err).
			Str("message", string(e.Content)).
			Msg("Failed to decode inventory message")
		return err
	}

	// To improve
	stockDao := db.MustOpenStockDao(a.pool)
	stockSvc := svc.MustStockService(stockDao)

	if err := stockSvc.ReceiveInventory(ctx, receiveInventoryRequest); err != nil {
		log.ErrCtx(ctx, err).
			Str("request", string(e.Content)).
			Msg("Failed to update inventory with stock")
		return err
	}

	log.InfoCtx(ctx).
		Str("request", string(e.Content)).
		Msg("Inventory updated with stock")

	return nil
}
