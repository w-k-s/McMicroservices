package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	msg "github.com/w-k-s/McMicroservices/kitchen-service/internal/messages"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

type App struct {
	config         *cfg.Config
	mux            *mux.Router
	pool           *sql.DB
	amqpConnection *amqp.Connection
	amqpConsumer   *amqp.Channel
	amqpProducer   *amqp.Channel
}

var (
	defaultStockHandler stockHandler
	defaultOrderHandler OrderHandler
)

func (app *App) Config() *cfg.Config {
	return app.config
}

func Init(config *cfg.Config) (*App, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is required. Got %v", nil)
	}

	pool := db.MustOpenPool(config.Database())
	db.MustRunMigrations(pool, config.Database())

	amqpConnection, amqpConsumer, amqpProducer := msg.Must(msg.NewAmqpConnection(config.Broker()))
	msg.MustDeclareExchanges(amqpConsumer, amqpProducer)

	app := &App{
		config:         config,
		mux:            mux.NewRouter(),
		pool:           pool,
		amqpConnection: amqpConnection,
		amqpConsumer:   amqpConsumer,
		amqpProducer:   amqpProducer,
	}

	app.registerHealthEndpoint()
	app.registerStockEndpoint()
	app.registerOrderEndpoint()

	log.Printf("--- Application Initialized ---")
	return app, nil
}

func Must(app *App, err error) *App {
	if err != nil {
		log.Fatalf("failed to init application. Reason: %s", err)
	}
	return app
}

func (app *App) ListenAndServe() {
	server := &http.Server{
		Addr:           app.config.Server().ListenAddress(),
		Handler:        app.mux,
		ReadTimeout:    app.config.Server().ReadTimeout(),
		WriteTimeout:   app.config.Server().WriteTimeout(),
		MaxHeaderBytes: app.config.Server().MaxHeaderBytes(),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Error while listening and serving. Details: %s", err)
		}
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (kill -2)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), app.config.Server().ShutdownGracePeriod())
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Failed to shut down server gracefully in %s", app.config.Server().ShutdownGracePeriod())
	}
}

func (app *App) Router() *mux.Router {
	return app.mux
}

func (app *App) Close() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Failed to close application. Reason: %v\n", r)
		}
	}()
	var err error

	defaultStockHandler.Close()

	if err = app.amqpConnection.Close(); err != nil {
		log.Printf("Failed to close consumer. Reason: %q", err.Error())
	}

	if err = app.amqpConsumer.Close(); err != nil {
		log.Printf("Failed to close consumer. Reason: %q", err.Error())
	}

	if err = app.amqpProducer.Close(); err != nil {
		log.Printf("Failed to close consumer. Reason: %q", err.Error())
	}

	if err = app.pool.Close(); err != nil {
		log.Printf("Failed to close connection pool. Reason: %q", err.Error())
	}
}

func (app *App) registerHealthEndpoint() {
	healthHandler := MustHealthHandler(app.pool)

	app.mux.HandleFunc("/health", healthHandler.CheckHealth).
		Methods("GET")
}

func (app *App) registerStockEndpoint() {
	stockDao := db.MustOpenStockDao(app.pool)
	stockService := svc.MustStockService(stockDao)
	defaultStockHandler = NewStockHandler(stockService, app.amqpConsumer)

	stockRouter := app.mux.PathPrefix("/kitchen/api/v1/stock").Subrouter()
	stockRouter.HandleFunc("", defaultStockHandler.GetStock).
		Methods("GET")

}

func (app *App) registerOrderEndpoint() {
	stockDao := db.MustOpenStockDao(app.pool)
	orderService := svc.MustOrderService(stockDao)
	defaultOrderHandler = NewOrderHandler(orderService, app.amqpConsumer, app.amqpProducer)
}
