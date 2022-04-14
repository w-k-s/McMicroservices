package server

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/w-k-s/McMicroservices/kitchen-service/log"

	"github.com/Shopify/sarama"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
	"go.uber.org/multierr"
)

const (
	TopicCreateOrder string = "order_created"
	TopicOrderReady  string = "order_ready"
	TopicOrderFailed string = "order_failed"
)

type OrderHandler interface {
	HandleOrderMessage(ctx context.Context, request []byte) (string, []byte)
	Close() error
}

type orderHandler struct {
	Handler
	orderService svc.OrderService
	consumer     sarama.Consumer
	producer     sarama.SyncProducer
	cancelFunc   context.CancelFunc
}

func NewOrderHandler(
	orderService svc.OrderService,
	consumer sarama.Consumer,
	producer sarama.SyncProducer,
	logger log.Logger,
) OrderHandler {
	ctx, cancelFunc := context.WithCancel(logger.WithContext(context.Background()))

	orderHandler := &orderHandler{
		orderService: orderService,
		consumer:     consumer,
		producer:     producer,
		cancelFunc:   cancelFunc,
	}

	orderHandler.listenForNewOrderEvents(ctx)

	return orderHandler
}

func (oh orderHandler) Close() error {
	return multierr.Combine(
		oh.consumer.Close(),
		oh.producer.Close(),
	)
}

func (oh orderHandler) listenForNewOrderEvents(ctx context.Context) {
	log.InfoCtx(ctx).Msg("Listening for New Orders")

	var (
		partitionList []int32
		err           error
	)
	if partitionList, err = oh.consumer.Partitions(TopicCreateOrder); err != nil {
		log.ErrCtx(ctx, err).Msg("Failed to get partition list for stockHandler")
		return
	}

	// Create a cosumer for each partition.
	// Each consumer will listen for messages asynchronously
	// All of the consumers will send their messages to a single messageChannel
	initialOffset := sarama.OffsetOldest
	messageChannel := make(chan *sarama.ConsumerMessage)
	for _, partition := range partitionList {
		pc, _ := oh.consumer.ConsumePartition(TopicCreateOrder, partition, initialOffset)

		log.InfoCtx(ctx).
			Str("topic", TopicCreateOrder).
			Int32("partition", partition).
			Int64("offset", initialOffset).
			Msgf("Creating a consumer")

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
				topic, reply := oh.HandleOrderMessage(ctx, message.Value)
				oh.publishResponse(ctx, topic, reply)
				continue
			}
		}
	}()
}

func (oh orderHandler) HandleOrderMessage(ctx context.Context, request []byte) (string, []byte) {
	log.InfoCtx(ctx).
		Str("message", string(request)).
		Msgf("Order Message received")
	decoder := json.NewDecoder(bytes.NewReader(request))
	decoder.UseNumber()

	var (
		orderRequest  svc.OrderRequest
		orderResponse svc.OrderResponse
		err           error
	)
	if err = decoder.Decode(&orderRequest); err != nil {
		// TODO: This should probably go in to a failed-to-process queue
		log.ErrCtx(ctx, err).Msg("Failed to decode order request")
		return "", []byte{}
	}

	if orderResponse, err = oh.orderService.ProcessOrder(ctx, orderRequest); err != nil {
		return TopicOrderFailed, oh.MustMarshal(json.Marshal(orderResponse))
	}
	return TopicOrderReady, oh.MustMarshal(json.Marshal(orderResponse))
}

func (oh orderHandler) publishResponse(ctx context.Context, topic string, body []byte) {
	var (
		partition int32
		offset    int64
		err       error
	)
	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(body),
	}
	if partition, offset, err = oh.producer.SendMessage(message); err != nil {
		log.ErrCtx(ctx, err).
			Str("message", string(body)).
			Str("topic", topic).
			Msgf("Failed to publish message")
		return
	}
	log.InfoCtx(ctx).
		Str("message", string(body)).
		Str("topic", topic).
		Int32("partition", partition).
		Int64("offset", offset).
		Msg("Message published")
}
