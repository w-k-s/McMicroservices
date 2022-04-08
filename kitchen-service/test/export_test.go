package test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	db "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
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

	testConfig *cfg.Config
	err        error
)

func init() {
	if testConfig, _ = cfg.NewConfig(
		cfg.NewServerConfigBuilder().
			SetPort(9898).
			Build(),
		requestKafkaTestContainer(),
		requestDatabaseTestContainer(),
	); err != nil {
		log.Fatalf("Failed to configure application for tests. Reason: %s", err)
	}

	testDB = db.Must(db.OpenPool(testConfig.Database()))
	db.MustRunMigrations(testDB, testConfig.Database())
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

func requestKafkaTestContainer() cfg.BrokerConfig {
	consumerConfig, _ := cfg.NewConsumerConfig("group_id", "earliest")
	return cfg.NewBrokerConfig(
		[]string{"localhost:9012"},
		"plaintext",
		consumerConfig,
	)
}

func TestMain(m *testing.M) {
	defer func(exitCode int) {
		defer func(exitCode int) {
			if r := recover(); r != nil {
				log.Printf("Panic while cleaning tests. Reason: %v\n", r)
			}
			os.Exit(exitCode)
		}(exitCode)

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
