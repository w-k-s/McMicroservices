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

var configFileContents string = `
[server]
port = 8080

[database]
username = "jack.torrence"
password = "password"
name     = "overlook"
host     = "localhost"
port     = 5432
sslmode  = "disable"

[broker]
bootstrap_servers = ["localhost"]

[broker.consumer]
group_id = "group_id"
auto_offset_reset = "earliest"
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
	config, err := LoadConfig("", "", "", "")

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

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsNotProvided_WHEN_configFileDoesNotExistAtDefaultPath_THEN_errorIsReturned() {
	// GIVEN
	_ = os.Remove(DefaultConfigFilePath())

	// WHEN
	config, err := LoadConfig("", "", "", "")

	// THEN
	assert.Nil(suite.T(), config)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), fmt.Sprintf("failed to load config file from local path '%s'. Reason: open %s: no such file or directory", DefaultConfigFilePath(), DefaultConfigFilePath()), err.Error())
}

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsProvided_WHEN_configFileDoesNotExistAtProvidedPath_THEN_errorIsReturned() {
	// GIVEN
	path := "file://" + filepath.Join("/.kitchen", "test.d", "config.toml")

	// WHEN
	config, err := LoadConfig(path, "", "", "")

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Equal(suite.T(), "failed to load config file from local path '/.kitchen/test.d/config.toml'. Reason: open /.kitchen/test.d/config.toml: no such file or directory", err.Error())
}

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsProvided_WHEN_configFileDoesExistAtProvidedPath_THEN_configsParsedCorrectly() {
	// GIVEN
	var customConfigFileContents string = `
[server]
port = 8085
read_timeout = 5
write_timeout = 3
max_header_bytes = 2097152
shutdown_grace_period = 10

[database]
username = "danny.torrence"
password = "password"
name     = "tony"
host     = "localhost"
port     = 5432
sslmode  = "disable"

[broker]
bootstrap_servers = ["localhost"]
security_protocol = "ssl"

[broker.consumer]
group_id = "group_id"
auto_offset_reset = "earliest"
`
	assert.Nil(suite.T(), createTestConfigFile(customConfigFileContents, DefaultConfigFilePath()))

	// WHEN
	config, err := LoadConfig("", "", "", "")

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
	config, err := LoadConfig("", "", "", "")

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Contains(suite.T(), err.Error(), "Kafka Consumer Auto offset is required")
	assert.Contains(suite.T(), err.Error(), "Kafka Consumer Auto offset must either be 'earliest' or 'newest'")
}

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsProvided_WHEN_configFileDoesNotContainValidToml_THEN_errorIsReturned() {
	// GIVEN
	invalidToml := `{
		"database":{
			"port":8080
		}
	}`
	assert.Nil(suite.T(), createTestConfigFile(invalidToml, DefaultConfigFilePath()))

	// WHEN
	config, err := LoadConfig("", "", "", "")

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Equal(suite.T(), fmt.Sprintf("failed to load config file from local path '%s'. Reason: Near line 1 (last key parsed ''): expected '.' or '=', but got '{' instead", DefaultConfigFilePath()), err.Error())
}

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsNotPrefixedWithFileOrS3Protocol_WHEN_configFileIsLoaded_THEN_errorIsReturned() {
	// GIVEN
	uri := "http://" + filepath.Join("/.kitchen", "test.d", "config.toml")

	// WHEN
	config, err := LoadConfig(uri, "", "", "")

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Equal(suite.T(), "Config file must start with file:// or s3://", err.Error())
}
