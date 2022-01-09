package server

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/mux"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	msg "github.com/w-k-s/McMicroservices/kitchen-service/internal/messages"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

type App struct {
	config      *cfg.Config
	mux         *mux.Router
	topicRouter msg.TopicRouter
	pool        *sql.DB
	consumer    *kafka.Consumer
	producer    *kafka.Producer
}

func (app *App) Config() *cfg.Config {
	return app.config
}

func Init(config *cfg.Config) (*App, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is required. Got %v", nil)
	}

	pool := db.MustOpenPool(config.Database())
	consumer, producer := msg.MustNewConsumerProducerPair(config.Broker())

	db.MustRunMigrations(pool, config.Database())

	app := &App{
		config:      config,
		mux:         mux.NewRouter(),
		topicRouter: msg.TopicRouter{},
		pool:        pool,
		consumer:    consumer,
		producer:    producer,
	}

	app.registerHealthEndpoint()
	app.registerStockEndpoint()
	app.registerOrderEndpoint()

	msg.Start(app.consumer, app.producer, app.topicRouter)

	log.Printf("--- Application Initialized ---")
	return app, nil
}

func (app *App) Close() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Failed to close application. Reason: %v\n", r)
		}
	}()
	var err error

	// TODO: Closing Kafka causes usually causes panics in tests. Why?
	app.producer.Flush(15 * 1000)
	app.producer.Close()
	if err = app.consumer.Close(); err != nil {
		log.Printf("Failed to close consumer. Reason: %q", err.Error())
	}

	if err = app.pool.Close(); err != nil {
		log.Printf("Failed to close connection pool. Reason: %q", err.Error())
	}
}

func (app *App) Router() *mux.Router {
	return app.mux
}

func (app *App) registerHealthEndpoint() {
	healthHandler := MustHealthHandler(app.pool)

	app.mux.HandleFunc("/health", healthHandler.CheckHealth).
		Methods("GET")
}

func (app *App) registerStockEndpoint() {
	stockDao := db.MustOpenStockDao(app.pool)
	stockService := svc.MustStockService(stockDao)
	stockHandler := NewStockHandler(stockService)

	stock := app.mux.PathPrefix("/kitchen/api/v1/stock").Subrouter()
	stock.HandleFunc("", stockHandler.GetStock).
		Methods("GET")

	app.topicRouter[InventoryDelivery] = stockHandler.ReceiveInventory
}

func (app *App) registerOrderEndpoint() {
	stockDao := db.MustOpenStockDao(app.pool)
	orderService := svc.MustOrderService(stockDao)
	orderHandler := NewOrderHandler(orderService)

	app.topicRouter[CreateOrder] = orderHandler.HandleOrderMessage
}
