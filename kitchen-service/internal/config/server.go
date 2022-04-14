package config

import (
	"fmt"
	"time"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type ServerConfig interface {
	Port() int
	ReadTimeout() time.Duration
	WriteTimeout() time.Duration
	MaxHeaderBytes() int
	ShutdownGracePeriod() time.Duration
	ListenAddress() string
}

type defaultServerConfig struct {
	port                int
	readTimeout         time.Duration
	writeTimeout        time.Duration
	maxHeaderBytes      int
	shutdownGracePeriod time.Duration
}

func makeServerConfig(b *serverConfigBuilder) (ServerConfig, error) {

	errors := validate.Validate(
		&validators.IntIsGreaterThan{Name: "Server Port", Field: int(b.port), Compared: 1023, Message: "Server port must be at least 1023"},
	)

	if errors.HasAny() {
		return nil, errors
	}

	return defaultServerConfig{
		b.port,
		b.readTimeout,
		b.writeTimeout,
		b.maxHeaderBytes,
		b.shutdownGracePeriod,
	}, nil
}

func (s defaultServerConfig) Port() int {
	return s.port
}

func (s defaultServerConfig) MaxHeaderBytes() int {
	if s.maxHeaderBytes <= 0 {
		return 1 << 20 // 1MB
	}
	return s.maxHeaderBytes
}

func (s defaultServerConfig) ReadTimeout() time.Duration {
	if s.readTimeout == 0 {
		return 10 * time.Second
	}
	return s.readTimeout
}

func (s defaultServerConfig) WriteTimeout() time.Duration {
	if s.writeTimeout == 0 {
		return 10 * time.Second
	}
	return s.writeTimeout
}

func (s defaultServerConfig) ShutdownGracePeriod() time.Duration {
	if s.shutdownGracePeriod <= 0 {
		return 5 * time.Second
	}
	return s.shutdownGracePeriod
}

func (s defaultServerConfig) ListenAddress() string {
	return fmt.Sprintf(":%d", s.port)
}

type serverConfigBuilder struct {
	port                int
	readTimeout         time.Duration
	writeTimeout        time.Duration
	maxHeaderBytes      int
	shutdownGracePeriod time.Duration
}

func NewServerConfigBuilder() *serverConfigBuilder {
	return &serverConfigBuilder{
		port:                8080,
		readTimeout:         time.Duration(0),
		writeTimeout:        time.Duration(0),
		shutdownGracePeriod: time.Duration(0),
	}
}

func (b *serverConfigBuilder) SetPort(port int) *serverConfigBuilder {
	b.port = port
	return b
}

func (b *serverConfigBuilder) SetReadTimeout(timeout time.Duration) *serverConfigBuilder {
	b.readTimeout = timeout
	return b
}

func (b *serverConfigBuilder) SetWriteTimeout(timeout time.Duration) *serverConfigBuilder {
	b.writeTimeout = timeout
	return b
}

func (b *serverConfigBuilder) SetMaxHeaderBytes(maxHeaderBytes int) *serverConfigBuilder {
	b.maxHeaderBytes = maxHeaderBytes
	return b
}

func (b *serverConfigBuilder) SetShutdownGracePeriod(shutdownGracePeriod time.Duration) *serverConfigBuilder {
	b.shutdownGracePeriod = shutdownGracePeriod
	return b
}

func (b *serverConfigBuilder) Build() (ServerConfig, error) {
	return makeServerConfig(b)
}
