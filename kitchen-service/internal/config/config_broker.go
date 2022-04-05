package config

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type BrokerConfig struct {
	boostrapServers  []string
	securityProtocol string
	consumerConfig   consumerConfig
	producerConfig   producerConfig
}

func (bc BrokerConfig) BootstrapServers() []string {
	return bc.boostrapServers
}

func (bc BrokerConfig) SecurityProtocol() string {
	if bc.securityProtocol == "" {
		return "plaintext"
	}
	return bc.securityProtocol
}

func (bc BrokerConfig) ConsumerConfig() consumerConfig {
	return bc.consumerConfig
}

func (bc BrokerConfig) ProducerConfig() producerConfig {
	return bc.producerConfig
}

func NewBrokerConfig(
	boostrapServers []string,
	securityProtocol string,
	consumerConfig consumerConfig,
) BrokerConfig {
	return BrokerConfig{
		boostrapServers:  boostrapServers,
		securityProtocol: securityProtocol,
		consumerConfig:   consumerConfig,
		producerConfig:   NewProducerConfig(),
	}
}

type AutoOffsetReset string

const (
	Earliest AutoOffsetReset = "earliest"
	Newest   AutoOffsetReset = "newest"
)

type consumerConfig struct {
	groupId         string
	autoOffsetReset AutoOffsetReset
}

func MustAutoOffsetReset(autoOffsetReset string) AutoOffsetReset {
	switch strings.ToLower(autoOffsetReset) {
	case string(Earliest):
		return Earliest
	case string(Newest):
		return Newest
	default:
		panic(fmt.Sprintf("Unknown or unsupported autoOffsetReset value: %q", autoOffsetReset))
	}
}

func NewConsumerConfig(groupId string, autoOffsetReset string) (consumerConfig, error) {
	errors := validate.Validate(
		&validators.StringLengthInRange{Name: "Kafka Consumer Auto Offset", Field: autoOffsetReset, Min: 1, Max: 0, Message: "Kafka Consumer Auto offset is required"},
		&validators.StringInclusion{Name: "Kafka Consumer Auto Offset", Field: autoOffsetReset, List: []string{"earliest", "newest"}, Message: fmt.Sprintf("Kafka Consumer Auto offset must either be 'earliest' or 'newest'. Got %q", autoOffsetReset)},
	)

	if errors.HasAny() {
		return consumerConfig{}, errors
	}

	return consumerConfig{
		groupId,
		MustAutoOffsetReset(autoOffsetReset),
	}, nil
}

func (cc consumerConfig) GroupId() string {
	return cc.groupId
}

func (cc consumerConfig) AutoOffsetReset() AutoOffsetReset {
	return cc.autoOffsetReset
}

type producerConfig struct {
}

func NewProducerConfig() producerConfig {
	return producerConfig{}
}
