package messages

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
)

const (
	Durable    = true
	Exclusive  = true
	Immediate  = true
	Internal   = true
	Mandatory  = true
	NoLocal    = true
	NoWait     = true
	AutoAck    = true
	AutoDelete = true

	OrderExchange             = "orders"
	InventoryDeliveryExchange = "inventory_delivery"
)

func NewAmqpConnection(brokerConfig cfg.BrokerConfig) (*amqp.Connection, *amqp.Channel, *amqp.Channel, error) {
	var (
		conn     *amqp.Connection
		consumer *amqp.Channel
		producer *amqp.Channel
		err      error
	)
	if conn, err = amqp.Dial(brokerConfig.ServerAddress()); err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to connect to RabbitMQ. Reason: %w", err)
	}

	if consumer, err = conn.Channel(); err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to open consumer channel. Reason: %w", err)
	}

	if producer, err = conn.Channel(); err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to open consumer channel. Reason: %w", err)
	}

	return conn, consumer, producer, nil
}

func Must(conn *amqp.Connection, c *amqp.Channel, p *amqp.Channel, err error) (*amqp.Connection, *amqp.Channel, *amqp.Channel) {
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ. Reason: %s", err)
	}
	return conn, c, p
}

func MustDeclareExchanges(consumer *amqp.Channel, producer *amqp.Channel) {
	exchanges := []struct {
		Name string
		Type string
	}{
		{Name: OrderExchange, Type: "topic"},
		{Name: InventoryDeliveryExchange, Type: "fanout"},
	}

	for _, exchange := range exchanges {
		if err := producer.ExchangeDeclare(
			exchange.Name, // name
			exchange.Type, // type
			Durable,       // durable
			!AutoDelete,   // auto-deleted
			!Internal,     // internal
			!NoWait,       // no-wait
			nil,           // arguments
		); err != nil {
			log.Fatalf("Failed to declare exchange %q. Reason: %q", exchange.Name, err)
		}
	}
}
