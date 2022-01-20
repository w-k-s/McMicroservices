package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
	msg "github.com/w-k-s/McMicroservices/kitchen-service/internal/messages"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

type OrderHandler interface {
	HandleOrderMessage(ctx context.Context, request []byte) (string, []byte)
}

type orderHandler struct {
	Handler
	orderService svc.OrderService
	consumer     *amqp.Channel
	producer     *amqp.Channel
	cancelFunc   context.CancelFunc
}

func NewOrderHandler(
	orderService svc.OrderService,
	consumer *amqp.Channel,
	producer *amqp.Channel,
) OrderHandler {

	ctx, cancelFunc := context.WithCancel(context.Background())
	orderHandler := &orderHandler{
		orderService: orderService,
		consumer:     consumer,
		producer:     producer,
		cancelFunc:   cancelFunc,
	}

	go orderHandler.listenForOrderEvents(ctx)

	return orderHandler
}

func (s orderHandler) listenForOrderEvents(ctx context.Context) {
	var (
		queueName      string = "kitchen_service.order_queue"
		messageChannel <-chan amqp.Delivery
		err            error
	)
	if _, err = s.consumer.QueueDeclare(queueName, msg.Durable, !msg.AutoDelete, !msg.Exclusive, !msg.NoWait, nil); err != nil {
		log.Fatalf("Failed to declare queue %q. Reason: %s", queueName, err)
	}

	if err = s.consumer.QueueBind(
		queueName,         // queue name
		"order.new",       // routing key
		msg.OrderExchange, // exchange
		!msg.NoWait,
		nil,
	); err != nil {
		log.Fatalf("Failed to bind queue %q to exchange %q. Reason: %s", queueName, msg.OrderExchange, err)
	}

	if messageChannel, err = s.consumer.Consume(
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
				request := message.Body
				if len(request) == 0 {
					continue
				}
				key, resp := s.HandleOrderMessage(ctx, request)
				if len(key) == 0 {
					break
				}
				if err = s.producer.Publish(
					msg.OrderExchange, // exchange
					key,               // routing key
					false,             // mandatory
					false,             // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(resp),
					}); err != nil {
					log.Printf("Failed to deliver message %q (routing key:  %q). Reason: %q", string(resp), key, err)
				}

			}
		}
	}()
}

func (oh orderHandler) Close() {
	oh.cancelFunc()
}

func (oh orderHandler) HandleOrderMessage(ctx context.Context, request []byte) (string, []byte) {

	decoder := json.NewDecoder(bytes.NewReader(request))
	decoder.UseNumber()

	var (
		orderRequest  svc.OrderRequest
		orderResponse svc.OrderResponse
		err           error
	)
	if err = decoder.Decode(&orderRequest); err != nil {
		// TODO: This should probably go in to a failed-to-process queue
		log.Printf("Failed to decode order request. Reason: %s", err)
		return "", []byte{}
	}

	if orderResponse, err = oh.orderService.ProcessOrder(ctx, orderRequest); err != nil {
		return "order.failed", oh.MustMarshal(json.Marshal(orderResponse))
	}
	return "order.ready", oh.MustMarshal(json.Marshal(orderResponse))
}
