package server

import (
	"fmt"
	"log"
	"net/http"

	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	svc "github.com/w-k-s/McMicroservices/kitchen-service/pkg/services"
)

type App struct {
	config *cfg.Config
}

func (app *App) Config() *cfg.Config {
	return app.config
}

func Init(config *cfg.Config) (*App, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is required. Got %v", nil)
	}

	db.MustRunMigrations(
		config.Database(),
	)

	log.Printf("--- Application Initialized ---")
	return &App{
		config: config,
	}, nil
}

func (app *App) Router() *http.ServeMux {
	mux := http.NewServeMux()

	app.registerHealthEndpoint(mux)
	app.registerStockEndpoint(mux)

	return mux
}

func (app *App) registerHealthEndpoint(mux *http.ServeMux) {
	healthHandler := MustHealthHandler(app.config)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
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

func (app *App) registerStockEndpoint(mux *http.ServeMux) {
	dao := db.MustOpenDao(
		app.config.Database().DriverName(),
		app.config.Database().ConnectionString(),
	)

	stockService := svc.MustStockService(dao)
	stockHandler := NewStockHandler(stockService)

	mux.HandleFunc("/kitchen/api/v1/stock", func(w http.ResponseWriter, r *http.Request) {
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
