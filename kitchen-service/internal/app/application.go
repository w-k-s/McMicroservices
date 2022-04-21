package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/w-k-s/McMicroservices/kitchen-service/internal/broker"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	"github.com/w-k-s/McMicroservices/kitchen-service/log"
)

type App struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	config     *cfg.Config
	consumer   broker.Consumer
	producer   broker.Producer
	router     *mux.Router
	pool       *sql.DB
	logger     log.Logger
	encoder    Encoder
	decoder    Decoder
	serde      JsonSerde
}

type Builder struct {
	Config   *cfg.Config
	Pool     *sql.DB
	Consumer broker.Consumer
	Producer broker.Producer
}

func New(b Builder) (*App, error) {
	if b.Config == nil {
		return nil, fmt.Errorf("server config is required")
	}

	logger, err := cfg.ConfigureLogging()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(logger.WithContext(context.Background()))

	db.MustRunMigrations(b.Pool, b.Config.Database())

	router := mux.NewRouter()
	router.Use(loggingMiddleware(logger))

	app := &App{
		ctx:        ctx,
		cancelFunc: cancel,
		config:     b.Config,
		router:     router,
		consumer:   b.Consumer,
		producer:   b.Producer,
		pool:       b.Pool,
		logger:     logger,
		encoder:    JSONEncoder{},
		decoder:    JSONDecoder{},
		serde:      DefaultJsonSerde,
	}

	app.configureRoutes()
	app.configureEventHandlers()
	app.consumer.Start(app.ctx)

	logger.Printf("--- Application Initialized ---")
	return app, nil
}

func Must(app *App, err error) *App {
	if err != nil {
		log.Fatalf("failed to init application. Reason: %s", err)
	}
	return app
}

func (app *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	app.router.ServeHTTP(w, req)
}

func (app *App) Close() {
	defer func() {
		if r := recover(); r != nil {
			app.logger.Printf("Failed to close application. Reason: %v\n", r)
		}
	}()

	app.cancelFunc()

	if err := app.consumer.Close(); err != nil {
		app.logger.Printf("Error while closing consumer: %q", err)
	}
	if err := app.producer.Close(); err != nil {
		app.logger.Printf("Error while closing producer: %q", err)
	}
	if err := app.pool.Close(); err != nil {
		app.logger.Printf("Failed to close connection pool. Reason: %q", err.Error())
	}
}
