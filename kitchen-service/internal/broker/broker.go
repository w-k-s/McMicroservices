package broker

import (
	"context"
)

type EventHandler func(ctx context.Context, e Message) error

type Message struct {
	Topic   string
	Key     string
	Content []byte
}

type SendMessageOptions interface{}

type Consumer interface {
	AddTopicEventHandler(topic string, handler EventHandler)
	Start(ctx context.Context) error
	Close() error
}

type MockConsumer interface {
	AddTopicEventHandler(topic string, handler EventHandler)
	Start(ctx context.Context) error
	Close() error

	YieldMessage(message Message)
}

type Producer interface {
	SendMessage(
		ctx context.Context,
		message Message,
		opts SendMessageOptions,
	) error
	Close() error
}

type MessageContentVerifier func(val []byte) error
type MockProducer interface {
	SendMessage(
		ctx context.Context,
		message Message,
		opts SendMessageOptions,
	) error
	Close() error

	VerifyMessageSent(verifier MessageContentVerifier)
}
