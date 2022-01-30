package config

import (
	"fmt"
	"strings"
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
	case "earliest":
		return Earliest
	case "newest":
		return Newest
	default:
		panic(fmt.Sprintf("Unknown or unsupported autoOffsetReset value: %q", autoOffsetReset))
	}
}

func NewConsumerConfig(groupId string, autoOffsetReset AutoOffsetReset) consumerConfig {
	return consumerConfig{
		groupId,
		autoOffsetReset,
	}
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
