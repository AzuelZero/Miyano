// Package config loads the application configuration from the environment.
package config

import "os"

const defaultPort = "8080"

// Config holds the application settings.
type Config struct {
	Port string
}

// Load reads configuration from environment variables, applying sane defaults.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	return Config{Port: port}
}
