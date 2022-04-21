package main

import (
	"flag"

	a "github.com/w-k-s/McMicroservices/kitchen-service/internal/app"
	"github.com/w-k-s/McMicroservices/kitchen-service/internal/broker"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	"github.com/w-k-s/McMicroservices/kitchen-service/internal/server"
)

var (
	configFileUrl string
	config        *cfg.Config
)

func init() {
	const (
		configFileUrlUsage = "URI to download the config file"
	)
	flag.StringVar(&configFileUrl, "uri", "", configFileUrlUsage)
}

func main() {
	// LoadConfig must be called in the main function and not in the init function because
	// the init function is called in tests but the config file does not exist.
	// This results in a panic.
	flag.Parse()

	config = cfg.Must(cfg.LoadConfig(configFileUrl))
	pool := db.Must(db.OpenPool(config.Database()))
	consumer := broker.MustConsumer(broker.KafkaConsumer(config.Broker(), config.Broker().ConsumerConfig()))
	producer := broker.MustProducer(broker.KafkaProducer(config.Broker()))

	app := a.Must(a.New(a.Builder{
		Config:   config,
		Pool:     pool,
		Consumer: consumer,
		Producer: producer,
	}))

	srv := server.New(config.Server(), app)
	srv.RegisterOnShutdown(app.Close)

	srv.ListenAndServe()
}
