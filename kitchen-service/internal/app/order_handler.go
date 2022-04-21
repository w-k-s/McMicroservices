package app

import (
	"bytes"
	"context"

	"github.com/w-k-s/McMicroservices/kitchen-service/internal/broker"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	"github.com/w-k-s/McMicroservices/kitchen-service/log"

	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

func (a *App) HandleOrderMessage(ctx context.Context, e broker.Message) error {
	log.InfoCtx(ctx).
		Str("request", string(e.Content)).
		Msgf("Order Message received")

	var orderRequest svc.OrderRequest
	if err := a.decoder.Decode(bytes.NewReader(e.Content), &orderRequest); err != nil {
		// TODO: This should probably go in to a failed-to-process queue
		log.ErrCtx(ctx, err).Msg("Failed to decode order request")
		return err
	}

	stockDao := db.MustOpenStockDao(a.pool)
	orderSvc := svc.MustOrderService(stockDao)

	orderResponse, err := orderSvc.ProcessOrder(ctx, orderRequest)

	var topic string
	if err != nil {
		topic = TopicOrderReady
	} else {
		topic = TopicOrderFailed
	}

	var buffer bytes.Buffer
	a.encoder.MustEncode(&buffer, orderResponse)

	a.producer.SendMessage(ctx, broker.Message{
		Topic:   topic,
		Content: buffer.Bytes(),
	}, nil)

	return nil
}
