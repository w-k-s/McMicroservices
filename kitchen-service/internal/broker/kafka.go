package broker

import (
	"context"
	"fmt"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/w-k-s/McMicroservices/kitchen-service/log"
)

type KafkaBrokerConfig interface {
	BootstrapServers() []string
}

type KafkaConsumerConfig interface {
	AutoOffsetReset() string
}

type kafkaConsumer struct {
	handlers        map[string][]EventHandler
	consumer        sarama.Consumer
	autoOffsetReset int64
}

func KafkaConsumer(brokerConfig KafkaBrokerConfig, consumerConfig KafkaConsumerConfig) (Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = saramaOffset(consumerConfig.AutoOffsetReset())
	consumer, err := sarama.NewConsumer(brokerConfig.BootstrapServers(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}
	return &kafkaConsumer{
		handlers:        map[string][]EventHandler{},
		consumer:        consumer,
		autoOffsetReset: config.Consumer.Offsets.Initial,
	}, nil
}

func MustConsumer(consumer Consumer, err error) Consumer {
	if err != nil {
		log.Fatal(err.Error())
	}
	return consumer
}

func saramaOffset(autoOffsetReset string) int64 {
	switch autoOffsetReset {
	case "earliest":
		return sarama.OffsetOldest
	case "newest":
		return sarama.OffsetNewest
	default:
		panic(fmt.Sprintf("autoOffsetReset %q can not be mapped to a sarama offset", autoOffsetReset))
	}
}

func (kc kafkaConsumer) AddTopicEventHandler(topic string, handler EventHandler) {
	// Probably need to make this thread-safe

	if list, ok := kc.handlers[topic]; ok {
		kc.handlers[topic] = append(list, handler)
	} else {
		kc.handlers[topic] = []EventHandler{handler}
	}
}

func (kc kafkaConsumer) Close() error {
	defer func() {
		for k, _ := range kc.handlers {
			delete(kc.handlers, k)
		}
	}()
	return kc.consumer.Close()
}

func (kc kafkaConsumer) Start(ctx context.Context) error {
	log.InfoCtx(ctx).Msg("Starting Kafka Consumer")

	// Get list of topics
	topics, err := kc.consumer.Topics()
	if err != nil {
		return fmt.Errorf("failed to fetch list of topics. Reason: '%w'", err)
	}

	// Get list of partitions for each topic
	topicPartitionsMap := map[string][]int32{}
	for _, topic := range topics {
		partitions, err := kc.consumer.Partitions(topic)
		if err != nil {
			return fmt.Errorf("failed to fetch partitions for topic %q: %w", topic, err)
		}
		topicPartitionsMap[topic] = partitions
	}

	// Create a consumer for each partition
	messageChannel := make(chan *sarama.ConsumerMessage)
	for topic, partitions := range topicPartitionsMap {
		for _, partition := range partitions {
			partitionConsumer, _ := kc.consumer.ConsumePartition(topic, partition, kc.autoOffsetReset)
			go func(partition int32) {
				log.InfoCtx(ctx).
					Str("topic", topic).
					Int32("partition", partition).
					Int64("offset", kc.autoOffsetReset).
					Msgf("Coreating consumer for partition")

				// Iterate over messgaes until channel is terminated
				for message := range partitionConsumer.Messages() {
					messageChannel <- message
				}

				log.InfoCtx(ctx).
					Str("topic", topic).
					Int32("partition", partition).
					Int64("offset", kc.autoOffsetReset).
					Msgf("Consumer Messages channel closed")
			}(partition)
		}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.InfoCtx(ctx).
					Msgf("Stopping the goroutine that's listening to messages")
				return // returning not to leak the goroutine
			case message := <-messageChannel:
				if handlers, ok := kc.handlers[message.Topic]; ok {
					for _, handler := range handlers {
						handler(ctx, Message{
							Topic:   message.Topic,
							Key:     string(message.Key),
							Content: message.Value,
						})
					}
				}
				continue
			}
		}
	}()

	return nil
}

// --
type KafkaProducerConfig interface {
	BootstrapServers() []string
}

type kafkaProducer struct {
	producer sarama.AsyncProducer
}

func KafkaProducer(config KafkaProducerConfig) (Producer, error) {
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Partitioner = sarama.NewRandomPartitioner
	producerConfig.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewAsyncProducer(config.BootstrapServers(), producerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}
	return &kafkaProducer{
		producer: producer,
	}, nil
}

func MustProducer(producer Producer, err error) Producer {
	if err != nil {
		log.Fatal(err.Error())
	}
	return producer
}

func (kp kafkaProducer) Close() error {
	return kp.producer.Close()
}

func (kp kafkaProducer) SendMessage(
	ctx context.Context,
	message Message,
	opts SendMessageOptions,
) error {
	producerMessage := &sarama.ProducerMessage{
		Topic: message.Topic,
		Key:   sarama.StringEncoder(message.Key),
		Value: sarama.StringEncoder(message.Content),
	}
	kp.producer.Input() <- producerMessage

	log.InfoCtx(ctx).
		Msg("Message enqueued")

	return nil
}

// Mock Kafka

type mockKafkaConsumer struct {
	kafkaConsumer
	topicConsumer map[string]*mocks.PartitionConsumer
}

func MockKafkaConsumer(t *testing.T, topicPartitionMap map[string]int32, config KafkaConsumerConfig) (MockConsumer, error) {

	testConsumer := mocks.NewConsumer(t, nil)

	topicPartitionsMap := map[string][]int32{}
	for topic, partition := range topicPartitionMap {
		topicPartitionsMap[topic] = []int32{partition}
	}
	testConsumer.SetTopicMetadata(topicPartitionsMap)

	topicConsumers := map[string]*mocks.PartitionConsumer{}
	for topic, partition := range topicPartitionMap {
		pc := testConsumer.ExpectConsumePartition(topic, partition, saramaOffset(config.AutoOffsetReset()))
		topicConsumers[topic] = pc
	}

	return &mockKafkaConsumer{
		kafkaConsumer{
			handlers:        map[string][]EventHandler{},
			consumer:        testConsumer,
			autoOffsetReset: saramaOffset(config.AutoOffsetReset()),
		},
		topicConsumers,
	}, nil
}

func (m mockKafkaConsumer) YieldMessage(msg Message) {
	if pc, ok := m.topicConsumer[msg.Topic]; ok {
		pc.YieldMessage(&sarama.ConsumerMessage{
			Topic: msg.Topic,
			Key:   []byte(msg.Key),
			Value: msg.Content,
		})
	}
	panic(fmt.Sprintf("There is no mock partitionConsumr created for Topic: %s", msg.Topic))
}

type mockKafkaProducer struct {
	mockProducer *mocks.AsyncProducer
}

func MockKafkaProducer(t *testing.T, topicPartitionMap map[string]int32, config KafkaConsumerConfig) (MockProducer, error) {
	testProducer := mocks.NewAsyncProducer(t, nil)
	return &mockKafkaProducer{
		testProducer,
	}, nil
}

func (m mockKafkaProducer) VerifyMessageSent(verifier MessageContentVerifier) {
	m.mockProducer.ExpectInputWithCheckerFunctionAndSucceed(mocks.ValueChecker(verifier))
}

func (m mockKafkaProducer) SendMessage(
	ctx context.Context,
	message Message,
	opts SendMessageOptions,
) error {
	m.mockProducer.Input() <- &sarama.ProducerMessage{
		Topic: message.Topic,
		Key:   sarama.StringEncoder(message.Key),
		Value: sarama.StringEncoder(message.Content),
	}
	return nil
}

func (m mockKafkaProducer) Close() error {
	return m.mockProducer.Close()
}
