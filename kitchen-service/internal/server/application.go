package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	msg "github.com/w-k-s/McMicroservices/kitchen-service/internal/messages"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

type App struct {
	config      *cfg.Config
	mux         *http.ServeMux
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

	log.Printf("--- Application Initialized ---")
	app := &App{
		config:      config,
		mux:         http.NewServeMux(),
		topicRouter: msg.TopicRouter{},
		pool:        pool,
		consumer:    consumer,
		producer:    producer,
	}

	app.registerHealthEndpoint()
	app.registerStockEndpoint()
	app.registerOrderEndpoint()

	msg.Start(app.consumer, app.producer, app.topicRouter)

	return app, nil
}

func (app *App) Close() {
	var err error
	app.producer.Flush(15 * 1000)
	app.producer.Close()
	if err = app.consumer.Close(); err != nil {
		log.Printf("Failed to close consumer. Reason: %q", err.Error())
	}

	if err = app.pool.Close(); err != nil {
		log.Printf("Failed to close connection pool. Reason: %q", err.Error())
	}
}

func (app *App) Router() *http.ServeMux {
	return app.mux
}

func (app *App) registerHealthEndpoint() {
	healthHandler := MustHealthHandler(app.config)

	app.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		var handle http.HandlerFunc
		switch r.Method {
		case http.MethodGet:
			handle = healthHandler.CheckHealth
		default:
			http.Error(w, "method not found", http.StatusMethodNotAllowed)
			return
		}
		handle(w, r)
	})
}

func (app *App) registerStockEndpoint() {
	stockDao := db.MustOpenStockDao(app.pool)
	stockService := svc.MustStockService(stockDao)
	stockHandler := NewStockHandler(stockService)

	app.mux.HandleFunc("/kitchen/api/v1/stock", func(w http.ResponseWriter, r *http.Request) {
		var handle http.HandlerFunc
		switch r.Method {
		case http.MethodGet:
			handle = stockHandler.GetStock
		default:
			http.Error(w, "method not found", http.StatusMethodNotAllowed)
			return
		}
		handle(w, r)
	})
}

func (app *App) registerOrderEndpoint() {
	stockDao := db.MustOpenStockDao(app.pool)
	orderService := svc.MustOrderService(stockDao)
	orderHandler := NewOrderHandler(orderService)

	app.topicRouter[CreateOrder] = orderHandler.HandleOrderMessage
}
