// Package config loads the application configuration from the environment.
package config

import "errors"

import "os"

const defaultPort = "8080"

// ErrMissingDatabaseURL is returned by Validate when DATABASE_URL is not set.
var ErrMissingDatabaseURL = errors.New("DATABASE_URL is required (e.g. postgres://miyano:dev@localhost:5432/miyano)")

// Config holds the application settings.
type Config struct {
	Port        string
	DatabaseURL string
}

// Load reads configuration from environment variables, applying sane defaults.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	return Config{
		Port:        port,
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

// Validate returns an error if the configuration cannot boot the application.
func (c Config) Validate() error {
	if c.DatabaseURL == "" {
		return ErrMissingDatabaseURL
	}
	return nil
}
