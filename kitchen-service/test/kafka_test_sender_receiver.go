package test

import (
	"encoding/json"
	"log"
	"reflect"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"golang.org/x/net/context"
)

type KafkaSender struct {
	producer *kafka.Producer
}

func NewKafkaSender(producer *kafka.Producer) KafkaSender {
	return KafkaSender{
		producer,
	}
}

func (ks KafkaSender) MustSendAsJSON(topic string, message interface{}) {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	if err = ks.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          bytes,
	}, nil); err != nil {
		log.Fatalln(err)
	}

	log.Printf("Sent message on topic %q: %q\n", topic, string(bytes))
}

type KafkaReceiver struct {
	received   map[string][]string
	consumer   *kafka.Consumer
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewKafkaReceiver(consumer *kafka.Consumer) KafkaReceiver {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return KafkaReceiver{
		received:   map[string][]string{},
		consumer:   consumer,
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}
}

func (kr KafkaReceiver) Close() {
	kr.cancelFunc()
}

func (kr KafkaReceiver) Listen(ctx context.Context) {
	go func() {
		run := true
		for run {
			select {
			case <-ctx.Done():
				run = false
				log.Printf("KafkaReceiver cancelled")
			default:
				kr.handleMessage()
			}
		}
	}()
}

func (kr KafkaReceiver) handleMessage() {
	msg, err := kr.consumer.ReadMessage(-1)
	if err != nil {
		log.Printf("Consumer error (kind: %s): %v (%v)\n", reflect.TypeOf(err), err, msg)
	}
	if list, ok := kr.received[*msg.TopicPartition.Topic]; ok {
		kr.received[*msg.TopicPartition.Topic] = append(list, string(msg.Value))
	} else {
		kr.received[*msg.TopicPartition.Topic] = []string{string(msg.Value)}
	}
}

func (kr KafkaReceiver) FirstMessage(topic string) string {
	if len(kr.received[topic]) == 0 {
		return ""
	}
	return kr.received[topic][0]
}

func (kr KafkaReceiver) MessagesForTopic(topic string) []string {
	return kr.received[topic]
}
