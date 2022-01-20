package test

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

// Sender

type TestMessageSender struct {
	producer *amqp.Channel
}

func NewTestMessageSender(producer *amqp.Channel) TestMessageSender {
	return TestMessageSender{
		producer,
	}
}

func (ms TestMessageSender) MustSendAsJSON(exchange, key string, message interface{}) {
	var (
		bytes []byte
		err   error
	)

	if bytes, err = json.Marshal(message); err != nil {
		log.Fatalf("Failed to marshal message %q. Reason: %q", string(bytes), err)
	}

	if err = ms.producer.Publish(exchange, key, true, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        bytes,
	}); err != nil {
		log.Fatalf("Failed to send message %q to exchange %q. Reason: %q", string(bytes), exchange, err)
	}

	log.Printf("Sent message on exchange %q: %q\n", exchange, string(bytes))
}

// Receiver

type TestMessageReceiver struct {
	received          receivedMessages
	consumer          *amqp.Channel
	ctx               context.Context
	cancelFunc        context.CancelFunc
	newMessageChannel chan string
}

func NewTestMessageReceiver(consumer *amqp.Channel) TestMessageReceiver {
	ctx, cancelFunc := context.WithCancel(context.Background())
	newMessageChannel := make(chan string)
	return TestMessageReceiver{
		received:          receivedMessages{},
		consumer:          consumer,
		ctx:               ctx,
		cancelFunc:        cancelFunc,
		newMessageChannel: newMessageChannel,
	}
}

func (mr TestMessageReceiver) Close() {
	log.Printf("Closing TestMessageReceiver")
	mr.cancelFunc()
	mr.received.Clear()
}

func (mr TestMessageReceiver) Listen(exchange, key string) {
	var (
		messageChannel <-chan amqp.Delivery
		q              amqp.Queue
		err            error
	)

	if q, err = mr.consumer.QueueDeclare("", false, true, false, false, nil); err != nil {
		log.Fatalf("Failed to create a queue to listen to key %q in exchange %q", key, exchange)
	}

	if err = mr.consumer.QueueBind(q.Name, key, exchange, false, nil); err != nil {
		log.Fatalf("Failed to bind queue %q to listen to key %q in exchange %q", q.Name, key, exchange)
	}

	if messageChannel, err = mr.consumer.Consume(q.Name, "", true, false, false, false, nil); err != nil {
		log.Fatalf("Failed to create consumer channel. Reason: %q", err)
	}

	go func() {
		run := true
		for run {
			select {
			case <-mr.ctx.Done():
				run = false
			case message := <-messageChannel:
				mr.handleMessage(exchange, key, message.Body)
			}
		}
	}()
}

func (mr TestMessageReceiver) handleMessage(exchange, key string, body []byte) {
	path := fmt.Sprintf("%s/%s", exchange, key)
	if list, ok := mr.received[path]; ok {
		mr.received[path] = append(list, string(body))
	} else {
		mr.received[path] = []string{string(body)}
	}
	mr.newMessageChannel <- path
}

func (mr TestMessageReceiver) WaitUntilNextMessageInTopic(timeout time.Duration, exchange, key string) waitBuilder {
	done := make(chan bool)
	path := fmt.Sprintf("%s/%s", exchange, key)
	go func() {
		log.Printf("TestMessageReceiver: Waiting for message from exchange %q with key %q\n", exchange, key)
		time.Sleep(timeout)
		done <- true
	}()

	for {
		select {
		case <-done:
			log.Printf("TestMessageReceiver: Waiting for message from exchange %q with key %q timed-out\n", exchange, key)
			return waitBuilder{}
		case newMessagePath := <-mr.newMessageChannel:
			if newMessagePath == path {
				log.Printf("TestMessageReceiver: Received message in %q\n", path)
				return waitBuilder{}
			}
		}
	}
}

func (mr TestMessageReceiver) FirstMessage(exchange, key string) string {
	return mr.received.FirstMessage(fmt.Sprintf("%s/%s", exchange, key))
}

func (mr TestMessageReceiver) String() string {
	return mr.received.String()
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
