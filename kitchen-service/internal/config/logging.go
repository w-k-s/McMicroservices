package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/w-k-s/McMicroservices/kitchen-service/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

func ConfigureLogging() (log.Logger, error) {
	var err error
	path := filepath.Join(defaultLogsDirectoryPath(), "server.log")
	if err = os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return nil, fmt.Errorf("failed to create temporary directory %q. Reason: %w", path, err)
	}

	if _, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		return nil, fmt.Errorf("failed to open log file. Reason: %w", err)
	}
	lumberjackLogger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    5, // MB
		MaxBackups: 10,
		MaxAge:     30,   // days
		Compress:   true, // disabled by default
	}
	multiWriter := io.MultiWriter(os.Stderr, lumberjackLogger)

	return log.NewLogger(multiWriter), nil
}
