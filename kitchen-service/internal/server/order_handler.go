package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

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
) OrderHandler {
	ctx, cancelFunc := context.WithCancel(context.Background())
	orderHandler := &orderHandler{
		orderService: orderService,
		consumer:     consumer,
		producer:     producer,
		cancelFunc:   cancelFunc,
	}

	orderHandler.listenForStockDeliveryEvents(ctx)

	return orderHandler
}

func (oh orderHandler) Close() error {
	return multierr.Combine(
		oh.consumer.Close(),
		oh.producer.Close(),
	)
}

func (oh orderHandler) listenForStockDeliveryEvents(ctx context.Context) {
	var (
		partitionList []int32
		err           error
	)
	if partitionList, err = oh.consumer.Partitions(TopicCreateOrder); err != nil {
		log.Printf("Failed to get partition list for stockHandler. Reason: %q", err)
		return
	}

	// Create a cosumer for each partition.
	// Each consumer will listen for messages asynchronously
	// All of the consumers will send their messages to a single messageChannel
	initialOffset := sarama.OffsetOldest
	messageChannel := make(chan *sarama.ConsumerMessage)
	for _, partition := range partitionList {
		pc, _ := oh.consumer.ConsumePartition(TopicCreateOrder, partition, initialOffset)
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
				oh.publishResponse(oh.HandleOrderMessage(ctx, message.Value))
				continue
			}
		}
	}()
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
		return TopicOrderFailed, oh.MustMarshal(json.Marshal(orderResponse))
	}
	return TopicOrderReady, oh.MustMarshal(json.Marshal(orderResponse))
}

func (oh orderHandler) publishResponse(topic string, body []byte) {
	var (
		partition int32
		offset    int64
		err       error
	)
	message := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: -1,
		Value:     sarama.StringEncoder(body),
	}
	if partition, offset, err = oh.producer.SendMessage(message); err != nil {
		log.Printf("Failed to publish message %q to topic %q (partition: %d). Reason: %q", string(body), topic, -1, err)
		return
	}
	log.Printf("Message %q published to topic %q with partition %d and offset %d", string(body), topic, partition, offset)
}
