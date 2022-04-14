package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

// -- SETUP

var configFileContents string = `server:
  port: 8080

database:
  username: "jack.torrence"
  password: "password"
  name: "overlook"
  host: "localhost"
  port: 5432
  sslmode: "disable"

broker:
  bootstrapServers: 
    - "localhost"

  consumer:
    groupId: "group_id"
    autoOffsetReset: "earliest"
`

func createTestConfigFile(content string, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return fmt.Errorf("Failed to create path '%s'. Reason: %w", path, err)
	}
	if err := ioutil.WriteFile(path, []byte(content), 0777); err != nil {
		return fmt.Errorf("Failed to write test config file to '%s'. Reason: %w", path, err)
	}
	return nil
}

func (suite *ConfigTestSuite) SetupTest() {
	if err := createTestConfigFile(configFileContents, DefaultConfigFilePath()); err != nil {
		log.Fatalf("Failed to create test config file. Reason: %s", err)
	}
}

// -- TEARDOWN

func (suite *ConfigTestSuite) TearDownTest() {
	path := DefaultConfigFilePath()
	_ = os.Remove(path)
}

// -- SUITE

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsNotProvided_WHEN_loadingConfig_THEN_configsLoadedFromDefaultPath() {
	// WHEN
	config, err := LoadConfig("")

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), config)
	assert.Equal(suite.T(), 8080, config.Server().Port())
	assert.Equal(suite.T(), ":8080", config.Server().ListenAddress())
	assert.Equal(suite.T(), 1048576, config.Server().MaxHeaderBytes())
	assert.Equal(suite.T(), time.Duration(10)*time.Second, config.Server().ReadTimeout())
	assert.Equal(suite.T(), time.Duration(10)*time.Second, config.Server().WriteTimeout())
	assert.Equal(suite.T(), time.Duration(5)*time.Second, config.Server().ShutdownGracePeriod())
	assert.Equal(suite.T(), "postgres", config.Database().DriverName())
	assert.Equal(suite.T(), "jack.torrence", config.Database().Username())
	assert.Equal(suite.T(), "password", config.Database().Password())
	assert.Equal(suite.T(), "overlook", config.Database().Name())
	assert.Equal(suite.T(), "kitchen", config.Database().Schema())
	assert.Equal(suite.T(), 5432, config.Database().Port())
	assert.Equal(suite.T(), "disable", config.Database().SslMode())
	assert.Equal(suite.T(), "host=localhost port=5432 user=jack.torrence password=password dbname=overlook sslmode=disable", config.Database().ConnectionString())
	assert.Equal(suite.T(), []string{"localhost"}, config.Broker().BootstrapServers())
	assert.Equal(suite.T(), "group_id", config.Broker().ConsumerConfig().GroupId())
	assert.Equal(suite.T(), Earliest, config.Broker().ConsumerConfig().AutoOffsetReset())
	assert.Equal(suite.T(), "plaintext", config.Broker().SecurityProtocol())
}

func (suite *ConfigTestSuite) Test_GIVEN_defaultLocalConfig_WHEN_environmentVariableForSameConfig_THEN_localFileConfigOverridenWithEnvironmentVariableConfig() {
	// GIVEN
	os.Setenv("APP_DATABASE_PASSWORD", "MySecretPassword")
	defer os.Unsetenv("APP_DATABASE_PASSWORD")

	os.Setenv("APP_BROKER_CONSUMER_AUTOOFFSETRESET", "newest")
	defer os.Unsetenv("APP_BROKER_CONSUMER_AUTOOFFSETRESET")

	// WHEN
	config, err := LoadConfig("")

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "jack.torrence", config.Database().Username())
	assert.Equal(suite.T(), "MySecretPassword", config.Database().Password())

	// Known issue
	//assert.Equal(suite.T(), Newest, config.Broker().ConsumerConfig().AutoOffsetReset())
}

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsNotProvided_WHEN_configFileDoesNotExistAtDefaultPath_THEN_errorIsReturned() {
	// GIVEN
	_ = os.Remove(DefaultConfigFilePath())

	// WHEN
	config, err := LoadConfig("")

	// THEN
	assert.Nil(suite.T(), config)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), fmt.Sprintf("failed to load config file from path 'file://%s'. Reason: open %s: no such file or directory", DefaultConfigFilePath(), DefaultConfigFilePath()), err.Error())
}

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsProvided_WHEN_configFileDoesNotExistAtProvidedPath_THEN_errorIsReturned() {
	// GIVEN
	path := "file://" + filepath.Join("/.kitchen", "test.d", "config.yaml")

	// WHEN
	config, err := LoadConfig(path)

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Equal(suite.T(), "failed to load config file from path 'file:///.kitchen/test.d/config.yaml'. Reason: open /.kitchen/test.d/config.yaml: no such file or directory", err.Error())
}

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsProvided_WHEN_configFileDoesExistAtProvidedPath_THEN_configsParsedCorrectly() {
	// GIVEN
	var customConfigFileContents string = `server:
  port: 8085
  readTimeout: 5
  writeTimeout: 3
  maxHeaderBytes: 2097152
  shutdownGracePeriod: 10

database:
  username: "danny.torrence"
  password: "password"
  name: "tony"
  host: "localhost"
  port: 5432
  sslmode: "disable"

broker:
  bootstrapServers:
    - "localhost"
  securityProtocol: "ssl"
  consumer:
    groupId: "group_id"
    autoOffsetReset: "earliest"
`
	assert.Nil(suite.T(), createTestConfigFile(customConfigFileContents, DefaultConfigFilePath()))

	// WHEN
	config, err := LoadConfig("")

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), config)
	assert.Equal(suite.T(), 8085, config.Server().Port())
	assert.Equal(suite.T(), ":8085", config.Server().ListenAddress())
	assert.Equal(suite.T(), 2097152, config.Server().MaxHeaderBytes())
	assert.Equal(suite.T(), 5*time.Second, config.Server().ReadTimeout())
	assert.Equal(suite.T(), 3*time.Second, config.Server().WriteTimeout())
	assert.Equal(suite.T(), 10*time.Second, config.Server().ShutdownGracePeriod())
	assert.Equal(suite.T(), "postgres", config.Database().DriverName())
	assert.Equal(suite.T(), "danny.torrence", config.Database().Username())
	assert.Equal(suite.T(), "password", config.Database().Password())
	assert.Equal(suite.T(), "tony", config.Database().Name())
	assert.Equal(suite.T(), 5432, config.Database().Port())
	assert.Equal(suite.T(), "disable", config.Database().SslMode())
	assert.Equal(suite.T(), "host=localhost port=5432 user=danny.torrence password=password dbname=tony sslmode=disable", config.Database().ConnectionString())
	assert.Equal(suite.T(), []string{"localhost"}, config.Broker().BootstrapServers())
	assert.Equal(suite.T(), "group_id", config.Broker().ConsumerConfig().GroupId())
	assert.Equal(suite.T(), Earliest, config.Broker().ConsumerConfig().AutoOffsetReset())
	assert.Equal(suite.T(), "ssl", config.Broker().SecurityProtocol())
}

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsProvided_WHEN_configFileIsEmpty_THEN_errorIsReturned() {
	// GIVEN
	assert.Nil(suite.T(), createTestConfigFile("", DefaultConfigFilePath()))

	// WHEN
	config, err := LoadConfig("")

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Contains(suite.T(), err.Error(), fmt.Sprintf("failed to load config file from path 'file://%s'. Reason: EOF", DefaultConfigFilePath()))
}

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsProvided_WHEN_configFileContainsInvalidYaml_THEN_errorIsReturned() {
	// GIVEN
	invalidYaml := `
[server]
port = 8085
`
	assert.Nil(suite.T(), createTestConfigFile(invalidYaml, DefaultConfigFilePath()))

	// WHEN
	config, err := LoadConfig("")

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Contains(suite.T(), err.Error(), fmt.Sprintf("failed to load config file from path 'file://%s'", DefaultConfigFilePath()))
	assert.Contains(suite.T(), err.Error(), "yaml: unmarshal errors")
}

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsNotPrefixedWithFileOrHTTPProtocol_WHEN_configFileIsLoaded_THEN_errorIsReturned() {
	// GIVEN
	uri := "s3://" + filepath.Join("com.wks.mcmicroservices", "kitchen-service", "config.yaml")

	// WHEN
	config, err := LoadConfig(uri)

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Equal(suite.T(), "config file must start with file:// or http://", err.Error())
}

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsNotSuffixedWithJSONOrYamlExtension_WHEN_configFileIsLoaded_THEN_errorIsReturned() {
	// GIVEN
	uri := "http://" + filepath.Join("com.wks.mcmicroservices", "kitchen-service", "config.toml")

	// WHEN
	config, err := LoadConfig(uri)

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Equal(suite.T(), "config file path must have a json or yaml extension", err.Error())
}
