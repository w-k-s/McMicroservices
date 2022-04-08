package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/w-k-s/McMicroservices/kitchen-service/log"

	"github.com/Shopify/sarama"

	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

const (
	TopicInventoryDelivery string = "inventory_delivery"
)

type stockHandler struct {
	Handler
	stockSvc   svc.StockService
	consumer   sarama.Consumer
	cancelFunc context.CancelFunc
}

func NewStockHandler(stockSvc svc.StockService, consumer sarama.Consumer, logger log.Logger) stockHandler {
	ctx, cancelFunc := context.WithCancel(logger.WithContext(context.Background()))
	handler := stockHandler{
		Handler{},
		stockSvc,
		consumer,
		cancelFunc,
	}

	handler.listenForStockDeliveryEvents(ctx)

	return handler
}

func (s stockHandler) Close() error {
	var err error
	if err = s.consumer.Close(); err != nil {
		log.Printf("Failed to close order consumer. Reason: %q", err)
	}
	s.cancelFunc()
	return err
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

func (s stockHandler) listenForStockDeliveryEvents(ctx context.Context) {
	var (
		partitionList []int32
		err           error
	)
	if partitionList, err = s.consumer.Partitions(TopicInventoryDelivery); err != nil {
		log.ErrCtx(ctx, err).Msg("Failed to get partition list for stockHandler")
		return
	}

	// Create a cosumer for each partition.
	// Each consumer will listen for messages asynchronously
	// All of the consumers will send their messages to a single messageChannel
	initialOffset := sarama.OffsetOldest
	messageChannel := make(chan *sarama.ConsumerMessage)
	for _, partition := range partitionList {
		pc, _ := s.consumer.ConsumePartition(TopicInventoryDelivery, partition, initialOffset)
		go func(pc sarama.PartitionConsumer) {
			for message := range pc.Messages() {
				messageChannel <- message
			}
		}(pc)
	}

	// Hanldle the messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				return // returning not to leak the goroutine
			case message := <-messageChannel:
				s.receiveInventory(ctx, message.Value)
				continue
			}
		}
	}()
}

func (s stockHandler) receiveInventory(ctx context.Context, request []byte) {
	log.InfoCtx(ctx).Msg("Inventory Received...")
	decoder := json.NewDecoder(bytes.NewReader(request))
	decoder.UseNumber()

	var (
		receiveInventoryRequest svc.StockRequest
		err                     error
	)
	if err = decoder.Decode(&receiveInventoryRequest); err != nil {
		log.ErrCtx(ctx, err).
			Str("message", string(request)).
			Msg("Failed to decode inventory message")
		return
	}

	if err = s.stockSvc.ReceiveInventory(ctx, receiveInventoryRequest); err != nil {
		log.ErrCtx(ctx, err).
			Str("request", string(request)).
			Msg("Failed to update inventory with stock")
		return
	}

	log.InfoCtx(ctx).
		Str("request", string(request)).
		Msg("Inventory updated with stock")
}
