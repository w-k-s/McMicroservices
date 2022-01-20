package main

import (
	"flag"

	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	app "github.com/w-k-s/McMicroservices/kitchen-service/internal/server"
)

var (
	configFilePath string
	awsAccessKey   string
	awsSecretKey   string
	awsRegion      string
	config         *cfg.Config
	handler        *app.App
)

func init() {
	const (
		configFileUsage   = `Path to configFile. Must start with 'file://' (if file is in local filesystem) or 's3://' (if file is hosted on s3)`
		awsAccessKeyUsage = "AWS Access key; used to download config file. Only required if config file is hosted on s3"
		awsSecretKeyUsage = "AWS Secret key; used to download config file. Only required if config file is hosted on s3"
		awsRegionUsage    = "AWS Region; used to download config file. Only required if config file is hosted on s3"
	)
	flag.StringVar(&configFilePath, "file", "", configFileUsage)
	flag.StringVar(&awsAccessKey, "aws_access_key", "", awsAccessKeyUsage)
	flag.StringVar(&awsSecretKey, "aws_secret_key", "", awsSecretKeyUsage)
	flag.StringVar(&awsRegion, "aws_region", "", awsRegionUsage)
}

func main() {
	// LoadConfig must be called in the main function and not in the init function because
	// the init function is called in tests but the config file does not exist.
	// This results in a panic.
	flag.Parse()

	config = cfg.Must(cfg.LoadConfig(configFilePath, awsAccessKey, awsSecretKey, awsRegion))
	handler := app.Must(app.Init(config))
	defer handler.Close()

	handler.ListenAndServe()
}
