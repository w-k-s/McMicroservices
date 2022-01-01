package messages

import (
	"fmt"
	"log"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
)

type (
	RequestBody   []byte
	ResponseBody  []byte
	ResponseTopic string
	TopicHandler  func(request RequestBody) (ResponseTopic, ResponseBody)
	TopicRouter   map[string]TopicHandler
)

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
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": strings.Join(brokerConfig.BootstrapServers(), ","),
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
					log.Printf("Delivered message to %v\n", ev.TopicPartition)
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
			msg, err := consumer.ReadMessage(-1)
			if err != nil {
				log.Printf("Consumer error: %v (%v)\n", err, msg)
			}

			log.Printf("Message on %q: %s\n", msg.TopicPartition, string(msg.Value))
			handler, handlerFound = router[*msg.TopicPartition.Topic]
			if !handlerFound {
				log.Printf("No handler found for topic %q. Discarding Message: %q", msg.TopicPartition, string(msg.Value))
				continue
			}

			topic, resp := handler(msg.Value)
			err = producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: topic.StringPointer(), Partition: kafka.PartitionAny},
				Value:          resp,
			}, nil)

			if err != nil {
				log.Printf("Unable to produce message. Reason: %q", err)
			}
		}
	}()
}
