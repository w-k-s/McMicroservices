package config

type BrokerConfig struct {
	serverAddress string
}

func (bc BrokerConfig) ServerAddress() string {
	return bc.serverAddress
}

func NewBrokerConfig(
	serverAddress string,
) BrokerConfig {
	return BrokerConfig{
		serverAddress: serverAddress,
	}
}
