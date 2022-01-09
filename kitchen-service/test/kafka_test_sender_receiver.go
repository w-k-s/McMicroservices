package test

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"golang.org/x/net/context"
)

// Sender

type KafkaSender struct {
	producer *kafka.Producer
}

func NewKafkaSender(producer *kafka.Producer) KafkaSender {
	return KafkaSender{
		producer,
	}
}

// TODO: Sending takes ages in tests, can we fix this?
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

// Receiver

type KafkaReceiver struct {
	received          receivedMessages
	consumer          *kafka.Consumer
	ctx               context.Context
	cancelFunc        context.CancelFunc
	newMessageChannel chan *kafka.Message
}

func NewKafkaReceiver(consumer *kafka.Consumer) KafkaReceiver {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return KafkaReceiver{
		received:          receivedMessages{},
		consumer:          consumer,
		ctx:               ctx,
		cancelFunc:        cancelFunc,
		newMessageChannel: make(chan *kafka.Message),
	}
}

func (kr KafkaReceiver) Close() {
	log.Printf("Closing KafkaReceiver")
	kr.cancelFunc()
	kr.received.Clear()
}

func (kr KafkaReceiver) Listen() {
	go func() {
		run := true
		for run {
			select {
			case <-kr.ctx.Done():
				run = false
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
	kr.newMessageChannel <- msg
}

func (kr KafkaReceiver) WaitUntilNextMessageInTopic(timeout time.Duration, topic string) waitBuilder {
	done := make(chan bool)
	go func() {
		log.Printf("KafkaReceiver: Waiting for message in topic %q\n", topic)
		time.Sleep(timeout)
		done <- true
	}()

	for {
		select {
		case <-done:
			log.Printf("KafkaReceiver: Waiting for message in topic %q timed-out\n", topic)
			return waitBuilder{}
		case newMessage := <-kr.newMessageChannel:
			newMessageTopic := *newMessage.TopicPartition.Topic
			if newMessageTopic == topic {
				log.Printf("KafkaReceiver: Received message in topic %q: %q\n", newMessageTopic, string(newMessage.Value))
				return waitBuilder{}
			}
		}
	}
}

func (kr KafkaReceiver) FirstMessage(topic string) string {
	return kr.received.FirstMessage(topic)
}

func (kr KafkaReceiver) String() string {
	return kr.received.String()
}

type waitBuilder struct{}

func (wb waitBuilder) Plus(duration time.Duration) {
	time.Sleep(duration)
}

// Received Messages

type receivedMessages map[string][]string

func (r receivedMessages) Len() int {
	sum := 0
	for _, v := range r {
		sum += len(v)
	}
	return sum
}

func (r receivedMessages) FirstMessage(topic string) string {
	if len(r[topic]) == 0 {
		return ""
	}
	return r[topic][0]
}

func (r receivedMessages) MessagesForTopic(topic string) []string {
	return r[topic]
}

func (r receivedMessages) String() string {
	sb := strings.Builder{}
	for k, v := range r {
		sb.WriteString(fmt.Sprintf("%s: [%s]", k, strings.Join(v, ", ")))
	}
	return sb.String()
}

func (r receivedMessages) Clear() {
	for k := range r {
		delete(r, k)
	}
}
