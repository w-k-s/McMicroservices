package config

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lalamove/nui/nstrings"
	"github.com/w-k-s/McMicroservices/kitchen-service/log"
	"github.com/w-k-s/konfig"
	"github.com/w-k-s/konfig/loader/klenv"
	"github.com/w-k-s/konfig/loader/klfile"
	"github.com/w-k-s/konfig/loader/klhttp"
	"github.com/w-k-s/konfig/parser"
	"github.com/w-k-s/konfig/parser/kpjson"
	"github.com/w-k-s/konfig/parser/kpyaml"
)

type Config struct {
	server ServerConfig
	broker BrokerConfig
	db     DBConfig
}

func (c Config) Server() ServerConfig {
	return c.server
}

func (c Config) Database() DBConfig {
	return c.db
}

func (c Config) Broker() BrokerConfig {
	return c.broker
}

func NewConfig(serverConfig ServerConfig, brokerConfig BrokerConfig, dbConfig DBConfig) (*Config, error) {
	config := &Config{
		server: serverConfig,
		broker: brokerConfig,
		db:     dbConfig,
	}

	return config, nil
}

func LoadConfig(configFilePath string) (*Config, error) {
	return LoadConfigWithClient(configFilePath, http.DefaultClient)
}

func LoadConfigWithClient(configFilePath string, httpClient klhttp.Client) (*Config, error) {
	if len(configFilePath) == 0 {
		configFilePath = "file://" + DefaultConfigFilePath()
	}

	store := konfig.New(konfig.DefaultConfig())
	var loader konfig.LoaderWatcher

	var fileParser parser.Func
	if strings.HasSuffix(configFilePath, "json") {
		fileParser = kpjson.Parser
	} else if strings.HasSuffix(configFilePath, "yaml") {
		fileParser = kpyaml.Parser
	} else {
		return nil, fmt.Errorf("config file path must have a json or yaml extension")
	}

	if strings.HasPrefix(configFilePath, "file://") {
		absolutePath := strings.Replace(configFilePath, "file://", "", 1)
		loader = klfile.New(
			&klfile.Config{
				Files: []klfile.File{
					{
						Parser: fileParser,
						Path:   absolutePath,
					},
				},
			},
		)
	} else if strings.HasPrefix(configFilePath, "http://") {
		loader = klhttp.New(
			&klhttp.Config{
				Client: httpClient,
				Sources: []klhttp.Source{
					{
						URL:    configFilePath,
						Method: "GET",
						Parser: fileParser,
					},
				},
			},
		)
	} else {
		return nil, fmt.Errorf("config file must start with file:// or http://")
	}

	store.RegisterLoaderWatcher(loader)

	// Override file config with env config
	// KNOWN BUG: This doesn't work for configs keys that are camel cased
	// e.g. The env var APP_BROKER_CONSUMER_AUTOOFFSETRESET does not override broker.consumer.autoOffsetReset
	store.RegisterLoader(
		klenv.New(&klenv.Config{
			Regexp:         "^APP_.*",
			SliceSeparator: ",",
			Replacer: nstrings.ReplacerFunc(func(s string) string {
				return strings.ToLower(strings.Replace(strings.TrimPrefix(s, "APP_"), "_", ".", -1))
			}),
		}),
	)

	if err := store.Load(); err != nil {
		return nil, fmt.Errorf("failed to load config file from path '%s'. Reason: %w", configFilePath, err)
	}

	var (
		serverConfig   ServerConfig
		consumerConfig consumerConfig
		brokerConfig   BrokerConfig
		dbConfig       DBConfig
		err            error
	)
	if serverConfig, err = NewServerConfigBuilder().
		SetPort(store.Int("server.port")).
		SetReadTimeout(store.Duration("server.readTimeout") * time.Second).
		SetWriteTimeout(store.Duration("server.writeTimeout") * time.Second).
		SetMaxHeaderBytes(store.Int("server.maxHeaderBytes")).
		SetShutdownGracePeriod(store.Duration("server.shutdownGracePeriod") * time.Second).
		Build(); err != nil {
		return nil, fmt.Errorf("failed to load server config: %w", err)
	}

	if consumerConfig, err = NewConsumerConfig(
		store.String("broker.consumer.groupId"),
		store.String("broker.consumer.autoOffsetReset"),
	); err != nil {
		return nil, fmt.Errorf("failed to create consumer config: %w", err)
	}

	if brokerConfig, err = NewBrokerConfig(
		store.StringSlice("broker.bootstrapServers"),
		store.String("broker.securityProtocol"),
		consumerConfig,
	); err != nil {
		return nil, fmt.Errorf("failed to load broker config: %w", err)
	}

	if dbConfig, err = NewDBConfigBuilder().
		SetUsername(store.String("database.username")).
		SetPassword(store.String("database.password")).
		SetHost(store.String("database.host")).
		SetPort(store.Int("database.port")).
		SetName(store.String("database.name")).
		SetSSLMode(store.String("database.sslmode")).
		SetMigrationDirectory(store.String("database.migrationDir")).
		Build(); err != nil {
		return nil, fmt.Errorf("failed to load server config: %w", err)
	}

	return &Config{serverConfig, brokerConfig, dbConfig}, nil
}

func Must(config *Config, err error) *Config {
	if err != nil {
		log.Fatalf("failed to load config file. Reason: %s", err)
	}
	return config
}
