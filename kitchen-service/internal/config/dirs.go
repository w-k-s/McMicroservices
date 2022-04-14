package config

import (
	"log"
	"os"
	"path/filepath"
)

func DefaultApplicationRootDirectory() string {
	return filepath.Join(mustUserHomeDir(), ".kitchen")
}

func DefaultConfigFilePath() string {
	return filepath.Join(DefaultApplicationRootDirectory(), "config.yaml")
}

func DefaultMigrationsDirectoryPath() string {
	return filepath.Join(DefaultApplicationRootDirectory(), "migrations")
}

func defaultLogsDirectoryPath() string {
	return filepath.Join(DefaultApplicationRootDirectory(), "logs")
}

func mustUserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Unable to access user's home directory")
	}
	return homeDir
}
