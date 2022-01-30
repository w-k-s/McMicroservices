package messages

import (
	"fmt"
	"log"

	"github.com/Shopify/sarama"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
)

type ConsumerFactory func(cfg.BrokerConfig) (sarama.Consumer, error)
type ProducerFactory func(cfg.BrokerConfig) (sarama.SyncProducer, error)

func NewConsumer(brokerConfig cfg.BrokerConfig) (sarama.Consumer, error) {
	consumerConfig := sarama.NewConfig()
	consumerConfig.Consumer.Offsets.Initial = saramaOffset(brokerConfig.ConsumerConfig().AutoOffsetReset())
	return sarama.NewConsumer(brokerConfig.BootstrapServers(), consumerConfig)
}

func NewProducer(brokerConfig cfg.BrokerConfig) (sarama.SyncProducer, error) {
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Partitioner = sarama.NewRandomPartitioner
	producerConfig.Producer.RequiredAcks = sarama.WaitForAll
	return sarama.NewSyncProducer(brokerConfig.BootstrapServers(), producerConfig)
}

func MustConsumer(c sarama.Consumer, err error) sarama.Consumer {
	if err != nil {
		log.Fatalf("Failed to create consumer. Reason: %s", err)
	}
	return c
}

func MustProducer(p sarama.SyncProducer, err error) sarama.SyncProducer {
	if err != nil {
		log.Fatalf("Failed to create producer. Reason: %s", err)
	}
	return p
}

func saramaOffset(autoOffsetReset cfg.AutoOffsetReset) int64 {
	switch autoOffsetReset {
	case cfg.Earliest:
		return sarama.OffsetOldest
	case cfg.Newest:
		return sarama.OffsetNewest
	default:
		panic(fmt.Sprintf("autoOffsetReset %q can not be mapped to a sarama offset", autoOffsetReset))
	}
}
