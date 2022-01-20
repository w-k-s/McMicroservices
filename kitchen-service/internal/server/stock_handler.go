package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/streadway/amqp"
	msg "github.com/w-k-s/McMicroservices/kitchen-service/internal/messages"
	k "github.com/w-k-s/McMicroservices/kitchen-service/pkg/kitchen"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

type stockHandler struct {
	Handler
	consumerChannel *amqp.Channel
	cancelFunc      context.CancelFunc
	stockSvc        svc.StockService
}

func NewStockHandler(stockSvc svc.StockService, consumerChannel *amqp.Channel) stockHandler {
	ctx, cancelFunc := context.WithCancel(context.Background())
	handler := stockHandler{
		Handler{},
		consumerChannel,
		cancelFunc,
		stockSvc,
	}

	go handler.listenForInventoryEvents(ctx)

	return handler
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

func (s stockHandler) Close() {
	s.cancelFunc()
}

func (s stockHandler) listenForInventoryEvents(ctx context.Context) {
	var (
		queueName      string = "kitchen_service.inventory_queue"
		messageChannel <-chan amqp.Delivery
		err            error
	)
	if _, err = s.consumerChannel.QueueDeclare(queueName, msg.Durable, !msg.AutoDelete, !msg.Exclusive, !msg.NoWait, nil); err != nil {
		log.Fatalf("Failed to declare queue %q. Reason: %s", queueName, err)
	}

	if err = s.consumerChannel.QueueBind(
		queueName,                     // queue name
		"",                            // routing key
		msg.InventoryDeliveryExchange, // exchange
		!msg.NoWait,
		nil); err != nil {
		log.Fatalf("Failed to bind queue %q to exchange %q. Reason: %s", queueName, msg.OrderExchange, err)
	}

	if messageChannel, err = s.consumerChannel.Consume(
		queueName,      // queue
		"",             // consumer
		msg.AutoAck,    // auto-ack
		!msg.Exclusive, // exclusive
		!msg.NoLocal,   // no-local
		!msg.NoWait,    // no-wait
		nil,            // args
	); err != nil {
		log.Fatalf("Queue %q failed to consume from exchange %q. Reason: %s", queueName, msg.OrderExchange, err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return // returning not to leak the goroutine
			case message := <-messageChannel:
				s.receiveInventory(ctx, message.Body)
				continue
			}
		}
	}()
}

func (s stockHandler) receiveInventory(ctx context.Context, request []byte) {
	log.Println("Inventory Received...")
	decoder := json.NewDecoder(bytes.NewReader(request))
	decoder.UseNumber()

	var (
		receiveInventoryRequest svc.StockRequest
		err                     error
		kitchenError            k.Error
	)
	if err = decoder.Decode(&receiveInventoryRequest); err != nil {
		log.Printf("Failed to decode inventory message %q. Reason: %q ", string(request), err)
		return
	}

	if kitchenError = s.stockSvc.ReceiveInventory(ctx, receiveInventoryRequest); kitchenError != nil {
		log.Printf("Failed to update inventory with stock %q. Reason: %q ", receiveInventoryRequest.Stock, err)
		return
	}

	log.Printf("Inventory updated with stock %q", receiveInventoryRequest.Stock)
}
