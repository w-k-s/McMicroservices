package config

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigHTTPTestSuite struct {
	suite.Suite
	mockServer *MockClient
}

func TestConfigHTTPTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigHTTPTestSuite))
}

// -- Mocks

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type MockClient struct {
	Status  int
	Content string
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte(m.Content)))
	return &http.Response{
		StatusCode: m.Status,
		Body:       r,
	}, nil
}

// -- SETUP
const uri = "http://localhost:8888/kitchen-service-default.yaml"

func (suite *ConfigHTTPTestSuite) setupMockServer(status int, content string) {
	suite.mockServer = &MockClient{
		Status:  status,
		Content: content,
	}
}

func (suite *ConfigHTTPTestSuite) SetupTest() {
	suite.setupMockServer(200, configFileContents)
}

// -- TEARDOWN

func (suite *ConfigHTTPTestSuite) TearDownTest() {
	//
}

// -- SUITE

func (suite *ConfigHTTPTestSuite) Test_GIVEN_httpUriIsProvided_WHEN_configFileDoesNotExistAtProvidedPath_THEN_errorIsReturned() {
	// GIVEN

	suite.setupMockServer(404, `{"timestamp":"2022-04-13T18:50:27.115+00:00","status":404,"error":"Not Found","path":"/kitchen-service/"}`)

	// WHEN
	config, err := LoadConfigWithClient(uri, suite.mockServer)

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Contains(suite.T(), err.Error(), "failed to load config file from path 'http://localhost:8888/kitchen-service-default.yaml'")
	assert.Contains(suite.T(), err.Error(), "status code: 404")
}

func (suite *ConfigHTTPTestSuite) Test_GIVEN_httpUriIsProvided_WHEN_configFileDoesExistAtProvidedPath_THEN_configsParsedCorrectly() {

	// WHEN
	config, err := LoadConfigWithClient(uri, suite.mockServer)

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), config)
	assert.Equal(suite.T(), 8080, config.Server().Port())
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

func (suite *ConfigHTTPTestSuite) Test_GIVEN_defaultHttpConfig_WHEN_environmentVariableForSameConfig_THEN_localFileConfigOverridenWithEnvironmentVariableConfig() {
	// GIVEN
	os.Setenv("APP_DATABASE_PASSWORD", "MySecretPassword")
	defer os.Unsetenv("APP_DATABASE_PASSWORD")

	os.Setenv("APP_BROKER_CONSUMER_AUTOOFFSETRESET", "newest")
	defer os.Unsetenv("APP_BROKER_CONSUMER_AUTOOFFSETRESET")

	// WHEN
	config, err := LoadConfigWithClient(uri, suite.mockServer)

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "jack.torrence", config.Database().Username())
	assert.Equal(suite.T(), "MySecretPassword", config.Database().Password())

	// KNOWN ISSUE
	// assert.Equal(suite.T(), Newest, config.Broker().ConsumerConfig().AutoOffsetReset())
}

func (suite *ConfigHTTPTestSuite) Test_GIVEN_httpUriIsProvided_WHEN_configFileIsEmpty_THEN_errorIsReturned() {
	// GIVEN
	suite.setupMockServer(200, "")

	// WHEN
	config, err := LoadConfigWithClient(uri, suite.mockServer)

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Contains(suite.T(), err.Error(), "failed to load config file from path 'http://localhost:8888/kitchen-service-default.yaml")
	assert.Contains(suite.T(), err.Error(), "EOF")
}

func (suite *ConfigHTTPTestSuite) Test_GIVEN_httpUriIsProvided_WHEN_configFileDoesNotContainValidContent_THEN_errorIsReturned() {
	// GIVEN
	suite.setupMockServer(200, `"database.port":8080`)

	// WHEN
	config, err := LoadConfigWithClient(uri, suite.mockServer)

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Contains(suite.T(), err.Error(), "failed to load config file from path 'http://localhost:8888/kitchen-service-default.yaml'")
	assert.Contains(suite.T(), err.Error(), "yaml: unmarshal errors")
}
