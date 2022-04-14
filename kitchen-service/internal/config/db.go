package config

import (
	"fmt"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type DBConfig interface {
	Username() string
	Password() string
	Host() string
	Port() int
	Name() string
	SslMode() string
	MigrationDirectory() string
	DriverName() string
	Schema() string
	ConnectionString() string
}

type defaultDBConfig struct {
	username     string
	password     string
	host         string
	port         int
	name         string
	sslMode      string
	migrationDir string
}

func makeDBConfig(b *dbConfigBuilder) (DBConfig, error) {
	errors := validate.Validate(
		&validators.StringLengthInRange{Name: "Database Username", Field: b.username, Min: 1, Max: 0, Message: "Database username is required"},
		&validators.StringLengthInRange{Name: "Database Password", Field: b.password, Min: 1, Max: 0, Message: "Database password is required"},
		&validators.StringLengthInRange{Name: "Database Host", Field: b.host, Min: 1, Max: 0, Message: "Database host is required"},
		&validators.IntIsGreaterThan{Name: "Database Port", Field: int(b.port), Compared: 0, Message: "Database port is required"},
		&validators.StringLengthInRange{Name: "Database Name", Field: b.host, Min: 1, Max: 0, Message: "Database name is required"},
		&validators.StringInclusion{Name: "Database SSL Mode", Field: b.sslMode, List: []string{"disable", "require", "verify-ca", "verify-full"}, Message: "Database SSL Mode is required"},
		&validators.StringLengthInRange{Name: "Migration Directory", Field: b.host, Min: 1, Max: 0, Message: "Migration Directory path is required"},
	)

	if errors.HasAny() {
		return nil, errors
	}

	return defaultDBConfig{
		b.username,
		b.password,
		b.host,
		b.port,
		b.name,
		b.sslMode,
		b.migrationDir,
	}, nil
}

func (d defaultDBConfig) Username() string {
	return d.username
}

func (d defaultDBConfig) Password() string {
	return d.password
}

func (d defaultDBConfig) Host() string {
	return d.host
}

func (d defaultDBConfig) Port() int {
	return d.port
}

func (d defaultDBConfig) Name() string {
	return d.name
}

func (d defaultDBConfig) Schema() string {
	return "kitchen"
}

func (d defaultDBConfig) SslMode() string {
	return d.sslMode
}

func (d defaultDBConfig) DriverName() string {
	return "postgres"
}

func (d defaultDBConfig) MigrationDirectory() string {
	if len(d.migrationDir) == 0 {
		return DefaultMigrationsDirectoryPath()
	}
	return d.migrationDir
}

func (d defaultDBConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host(),
		d.Port(),
		d.Username(),
		d.Password(),
		d.Name(),
		d.SslMode(),
	)
}

type dbConfigBuilder struct {
	username     string
	password     string
	host         string
	port         int
	name         string
	sslMode      string
	migrationDir string
}

func NewDBConfigBuilder() *dbConfigBuilder {
	return &dbConfigBuilder{
		sslMode:      "disable",
		migrationDir: DefaultMigrationsDirectoryPath(),
	}
}

func (b *dbConfigBuilder) SetUsername(username string) *dbConfigBuilder {
	b.username = username
	return b
}

func (b *dbConfigBuilder) SetPassword(password string) *dbConfigBuilder {
	b.password = password
	return b
}

func (b *dbConfigBuilder) SetHost(host string) *dbConfigBuilder {
	b.host = host
	return b
}

func (b *dbConfigBuilder) SetPort(port int) *dbConfigBuilder {
	b.port = port
	return b
}

func (b *dbConfigBuilder) SetName(name string) *dbConfigBuilder {
	b.name = name
	return b
}

func (b *dbConfigBuilder) SetSSLMode(sslMode string) *dbConfigBuilder {
	b.sslMode = sslMode
	return b
}

func (b *dbConfigBuilder) SetMigrationDirectory(migrationDir string) *dbConfigBuilder {
	b.migrationDir = migrationDir
	return b
}

func (b *dbConfigBuilder) Build() (DBConfig, error) {
	return makeDBConfig(b)
}
