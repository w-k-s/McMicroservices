package main

import (
	"flag"

	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	app "github.com/w-k-s/McMicroservices/kitchen-service/internal/server"
)

var (
	configFileUrl string
	config        *cfg.Config
	handler       *app.App
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
	handler := app.Must(app.NewAppBuilder(config).Build())
	defer handler.Close()

	handler.ListenAndServe()
}
