package config

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type BrokerConfig interface {
	BootstrapServers() []string
	SecurityProtocol() string
	ConsumerConfig() consumerConfig
	ProducerConfig() producerConfig
}

type defaultBrokerConfig struct {
	boostrapServers  []string
	securityProtocol string
	consumerConfig   consumerConfig
	producerConfig   producerConfig
}

func (bc defaultBrokerConfig) BootstrapServers() []string {
	return bc.boostrapServers
}

func (bc defaultBrokerConfig) SecurityProtocol() string {
	if bc.securityProtocol == "" {
		return "plaintext"
	}
	return bc.securityProtocol
}

func (bc defaultBrokerConfig) ConsumerConfig() consumerConfig {
	return bc.consumerConfig
}

func (bc defaultBrokerConfig) ProducerConfig() producerConfig {
	return bc.producerConfig
}

func NewBrokerConfig(
	boostrapServers []string,
	securityProtocol string,
	consumerConfig consumerConfig,
) (BrokerConfig, error) {
	errors := validate.Validate(
		&boostrapServersValidator{Name: "Bootstrap servers", Field: boostrapServers},
	)

	if errors.HasAny() {
		return nil, errors
	}

	return defaultBrokerConfig{
		boostrapServers:  boostrapServers,
		securityProtocol: securityProtocol,
		consumerConfig:   consumerConfig,
		producerConfig:   NewProducerConfig(),
	}, nil
}

const (
	Earliest string = "earliest"
	Newest   string = "newest"
)

type consumerConfig struct {
	groupId         string
	autoOffsetReset string
}

func MustAutoOffsetReset(autoOffsetReset string) string {
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

func (cc consumerConfig) AutoOffsetReset() string {
	return cc.autoOffsetReset
}

type producerConfig struct {
}

func NewProducerConfig() producerConfig {
	return producerConfig{}
}

type boostrapServersValidator struct {
	Name  string
	Field []string
}

func (v *boostrapServersValidator) IsValid(errors *validate.Errors) {
	if len(v.Field) == 0 {
		errors.Add(v.Name, "servers list can not be empty")
	}
	for _, server := range v.Field {
		if len(server) == 0 {
			errors.Add(v.Name, "server list can not contain empty strings")
		}
	}
}
