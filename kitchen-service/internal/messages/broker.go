package messages

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
)

type (
	RequestBody   []byte
	ResponseBody  []byte
	ResponseTopic string
	TopicHandler  func(ctx context.Context, request RequestBody) (ResponseTopic, ResponseBody)
	TopicRouter   map[string]TopicHandler
)

func (rb RequestBody) Reader() io.Reader {
	return bytes.NewReader([]byte(rb))
}

func (rt ResponseTopic) StringPointer() *string {
	s := string(rt)
	return &s
}

func (tr TopicRouter) Topics() []string {
	topics := []string{}
	for key, _ := range tr {
		topics = append(topics, key)
	}
	return topics
}

func NewConsumerProducerPair(brokerConfig cfg.BrokerConfig) (*kafka.Consumer, *kafka.Producer, error) {
	log.Printf("boostrap.servers: %q, security.protocol: %q, group.id: %q, auto.offset.reset: %q\n",
		strings.Join(brokerConfig.BootstrapServers(), ","),
		brokerConfig.SecurityProtocol(),
		brokerConfig.ConsumerConfig().GroupId(),
		brokerConfig.ConsumerConfig().AutoOffsetReset(),
	)
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": strings.Join(brokerConfig.BootstrapServers(), ","),
		"security.protocol": brokerConfig.SecurityProtocol(),
		"group.id":          brokerConfig.ConsumerConfig().GroupId(),
		"auto.offset.reset": brokerConfig.ConsumerConfig().AutoOffsetReset(),
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": strings.Join(brokerConfig.BootstrapServers(), ",")})
	if err != nil {
		return c, nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return c, p, nil
}

func MustNewConsumerProducerPair(brokerConfig cfg.BrokerConfig) (*kafka.Consumer, *kafka.Producer) {
	var (
		c   *kafka.Consumer
		p   *kafka.Producer
		err error
	)
	if c, p, err = NewConsumerProducerPair(brokerConfig); err != nil {
		log.Fatalf("Failed to create consumer/producer pair. Reason: %s", err)
	}
	return c, p
}

func Start(
	consumer *kafka.Consumer,
	producer *kafka.Producer,
	router TopicRouter,
) {
	log.Printf("Consumer is subscribed to topics %q\n", router.Topics())
	consumer.SubscribeTopics(router.Topics(), nil)
	startLoggingDelivery(producer)
	startReadingMessages(consumer, producer, router)
}

func startLoggingDelivery(producer *kafka.Producer) {
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Delivered message to %q: %q\n", ev.TopicPartition, string(ev.Value))
				}
			}
		}
	}()
}

func startReadingMessages(
	consumer *kafka.Consumer,
	producer *kafka.Producer,
	router TopicRouter,
) {
	go func() {
		var (
			handler      TopicHandler
			handlerFound bool
		)

		for {
			log.Println("Waiting for messages...")
			msg, err := consumer.ReadMessage(-1)
			if err != nil {
				switch e := err.(type) {
				case kafka.Error:
					log.Printf("Kafka Error: %v\n", e)
					continue
				default:
					log.Printf("Consumer error (kind: %s): %v (%v)\n", reflect.TypeOf(err), err, msg)
					continue
				}
			}

			log.Printf("Message on %q: %s\n", msg.TopicPartition, string(msg.Value))
			handler, handlerFound = router[*msg.TopicPartition.Topic]
			if !handlerFound {
				log.Printf("No handler found for topic %q. Discarding Message: %q", msg.TopicPartition, string(msg.Value))
				continue
			}

			topic, resp := handler(context.Background(), msg.Value)
			if topic != "" && len(resp) > 0 {
				log.Printf("Sending message %q on topic %q", string([]byte(resp)), topic)
				err = producer.Produce(&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: topic.StringPointer(), Partition: kafka.PartitionAny},
					Value:          resp,
				}, nil)

				if err != nil {
					log.Printf("Unable to produce message. Reason: %q", err)
				}
			}
		}
	}()
}
