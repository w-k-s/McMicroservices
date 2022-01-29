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
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	msg "github.com/w-k-s/McMicroservices/kitchen-service/internal/messages"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

type appBuilder struct {
	config          *cfg.Config
	consumerFactory msg.ConsumerFactory
	producerFactory msg.ProducerFactory
}

func NewAppBuilder(config *cfg.Config) *appBuilder {
	return &appBuilder{
		config: config,
	}
}

func (b *appBuilder) SetConsumerFactory(cf msg.ConsumerFactory) *appBuilder {
	b.consumerFactory = cf
	return b
}

func (b appBuilder) GetConsumerFactory() msg.ConsumerFactory {
	if b.consumerFactory == nil {
		return msg.NewConsumer
	}
	return b.consumerFactory
}

func (b *appBuilder) SetProducerFactory(pf msg.ProducerFactory) *appBuilder {
	b.producerFactory = pf
	return b
}

func (b appBuilder) GetProducerFactory() msg.ProducerFactory {
	if b.producerFactory == nil {
		return msg.NewProducer
	}
	return b.producerFactory
}

func (b *appBuilder) Build() (*App, error) {
	return newApp(b)
}

var (
	defaultOrderHandler OrderHandler
	defaultStockHandler stockHandler
)

type App struct {
	config          *cfg.Config
	consumerFactory msg.ConsumerFactory
	producerFactory msg.ProducerFactory
	mux             *mux.Router
	pool            *sql.DB
}

func (app *App) Config() *cfg.Config {
	return app.config
}

func newApp(b *appBuilder) (*App, error) {
	if b.config == nil {
		return nil, fmt.Errorf("configuration is required. Got %v", nil)
	}

	pool := db.MustOpenPool(b.config.Database())
	db.MustRunMigrations(pool, b.config.Database())

	app := &App{
		config: b.config,
		mux:    mux.NewRouter(),
		pool:   pool,
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
	if err := defaultOrderHandler.Close(); err != nil {
		log.Printf("Error while closing order handler: %q", err)
	}
	if err := defaultStockHandler.Close(); err != nil {
		log.Printf("Error while closing stock handler: %q", err)
	}
	if err := app.pool.Close(); err != nil {
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
	defaultStockHandler = NewStockHandler(
		stockService,
		msg.MustConsumer(app.consumerFactory(app.config.Broker())),
	)

	stockRouter := app.mux.PathPrefix("/kitchen/api/v1/stock").Subrouter()
	stockRouter.HandleFunc("", defaultStockHandler.GetStock).
		Methods("GET")
}

func (app *App) registerOrderEndpoint() {
	stockDao := db.MustOpenStockDao(app.pool)
	orderService := svc.MustOrderService(stockDao)
	defaultOrderHandler = NewOrderHandler(
		orderService,
		msg.MustConsumer(app.consumerFactory(app.config.Broker())),
		msg.MustProducer(app.producerFactory(app.config.Broker())),
	)
}
