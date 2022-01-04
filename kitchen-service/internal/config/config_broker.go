package config

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

type consumerConfig struct {
	groupId         string
	autoOffsetReset string
}

func NewConsumerConfig(groupId string, autoOffsetReset string) consumerConfig {
	if autoOffsetReset == "" {
		autoOffsetReset = "earliest"
	}

	return consumerConfig{
		groupId,
		autoOffsetReset,
	}
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
