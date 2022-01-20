package test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/streadway/amqp"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	msg "github.com/w-k-s/McMicroservices/kitchen-service/internal/messages"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
	app "github.com/w-k-s/McMicroservices/kitchen-service/internal/server"
)

const (
	testContainerPostgresUser       = "test"
	testContainerPostgresPassword   = "test"
	testContainerPostgresDB         = "kitchen"
	testContainerDatabaseDriverName = "postgres"
)

var (
	testContainerDatabaseContext context.Context
	testContainerPostgres        tc.Container
	testContainerDataSourceName  string
	testDB                       *sql.DB

	testContainerRabbitMQContext context.Context
	testRabbitMQContainer        tc.Container
	testAmqpConnection           *amqp.Connection
	testAmqpConsumer             *amqp.Channel
	testAmqpProducer             *amqp.Channel

	testConfig *cfg.Config
	testApp    *app.App
	err        error
)

func init() {
	if testConfig, _ = cfg.NewConfig(
		cfg.NewServerConfigBuilder().
			SetPort(9898).
			Build(),
		requestRabbitMqTestContainer(),
		requestDatabaseTestContainer(),
	); err != nil {
		log.Fatalf("Failed to configure application for tests. Reason: %s", err)
	}

	testDB = db.MustOpenPool(testConfig.Database())
	testAmqpConnection, testAmqpConsumer, testAmqpProducer = msg.Must(msg.NewAmqpConnection(testConfig.Broker()))

	if testApp, err = app.Init(testConfig); err != nil {
		log.Fatalf("Failed to initialize application for tests. Reason: %s", err)
	}
}

func requestDatabaseTestContainer() cfg.DBConfig {
	testContainerDatabaseReq := tc.ContainerRequest{
		Image:        "postgres:11.6-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     testContainerPostgresUser,
			"POSTGRES_PASSWORD": testContainerPostgresPassword,
			"POSTGRES_DB":       testContainerPostgresDB,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	testContainerDatabaseContext = context.Background()
	testContainerPostgres, err = tc.GenericContainer(testContainerDatabaseContext, tc.GenericContainerRequest{
		ContainerRequest: testContainerDatabaseReq,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to request postgres test container: %s", err)
	}

	postgresHost, _ := testContainerPostgres.Host(testContainerDatabaseContext)
	postgresPort, _ := testContainerPostgres.MappedPort(testContainerDatabaseContext, "5432")

	testContainerDataSourceName = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", postgresHost, postgresPort.Int(), testContainerPostgresUser, testContainerPostgresPassword, testContainerPostgresDB)

	return cfg.NewDBConfigBuilder().
		SetUsername(testContainerPostgresUser).
		SetPassword(testContainerPostgresPassword).
		SetHost(postgresHost).
		SetPort(postgresPort.Int()).
		SetName(testContainerDataSourceName).
		Build()
}

func requestRabbitMqTestContainer() cfg.BrokerConfig {

	testContainerRabbitMqReq := tc.ContainerRequest{
		Image:        "rabbitmq:3.8.11-management",
		ExposedPorts: []string{"5672/tcp"},
		Env:          map[string]string{},
		WaitingFor:   wait.ForLog("Server startup complete"),
	}

	testContainerRabbitMQContext = context.Background()
	testRabbitMQContainer, err = tc.GenericContainer(testContainerRabbitMQContext, tc.GenericContainerRequest{
		ContainerRequest: testContainerRabbitMqReq,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to request rabbitmq test container: %s", err)
	}

	rabbitMqHost, _ := testRabbitMQContainer.Host(testContainerRabbitMQContext)
	rabbitMqPort, _ := testRabbitMQContainer.MappedPort(testContainerRabbitMQContext, "5672")

	return cfg.NewBrokerConfig(fmt.Sprintf("amqp://%s:%s", rabbitMqHost, rabbitMqPort.Port()))
}

func TestMain(m *testing.M) {
	defer func(exitCode int) {
		defer func(exitCode int) {
			if r := recover(); r != nil {
				log.Printf("Panic while cleaning tests. Reason: %v\n", r)
			}
			os.Exit(exitCode)
		}(exitCode)

		log.Println("Cleaning up after tests")

		testApp.Close()
		if err := testAmqpConsumer.Close(); err != nil {
			log.Printf("Failed to close test amqp Consumer Channel. %s", err)
		}
		if err := testAmqpProducer.Close(); err != nil {
			log.Printf("Failed to close test amqp Producer Channel. %s", err)
		}
		if err := testAmqpConnection.Close(); err != nil {
			log.Printf("Failed to close test amqp Connection. %s", err)
		}
		if err := testRabbitMQContainer.Terminate(testContainerRabbitMQContext); err != nil {
			log.Printf("Error closing Test RabbitMQ Container: %s", err)
		}
		if err := testContainerPostgres.Terminate(testContainerDatabaseContext); err != nil {
			log.Printf("Error closing Test Postgres Container: %s", err)
		}

		log.Print("Cleanup complete\n\n\n")

		os.Exit(exitCode)
	}(m.Run())
}

func clearTables() {
	if _, err := testDB.Exec("DELETE FROM kitchen.stock"); err != nil {
		log.Print("Failed to delete stock table: %w", err)
	}
}
