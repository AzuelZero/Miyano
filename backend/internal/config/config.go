// Package config loads the application configuration from the environment.
package config

import (
	"errors"
	"os"
	"strings"
)

const (
	defaultPort      = "8080"
	defaultMediaDir  = "media"
	minJWTSecretSize = 32
)

// forbiddenJWTSecrets are public values that must never sign tokens however
// long they are (audit 03/11: the .env.example value passed the size check).
var forbiddenJWTSecrets = map[string]struct{}{
	"change-me-minimum-32-characters!!": {},
}

// isTrivialSecret reports whether the secret is a single character repeated
// (e.g. 32 x "a") — syntactically long, practically public.
func isTrivialSecret(secret string) bool {
	return secret != "" && strings.Repeat(secret[:1], len(secret)) == secret
}

// ErrMissingDatabaseURL is returned by Validate when DATABASE_URL is not set.
var ErrMissingDatabaseURL = errors.New("DATABASE_URL is required (e.g. postgres://miyano:dev@localhost:5432/miyano)")

// ErrWeakJWTSecret is returned by Validate when JWT_SECRET is absent,
// shorter than the minimum size for HS256, or a known/trivial value.
var ErrWeakJWTSecret = errors.New("JWT_SECRET is required, must be at least 32 characters and not a known or trivial value")

// Config holds the application settings.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	MediaDir    string
}

// Load reads configuration from environment variables, applying sane defaults.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	mediaDir := os.Getenv("MEDIA_DIR")
	if mediaDir == "" {
		mediaDir = defaultMediaDir
	}
	return Config{
		Port:        port,
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		MediaDir:    mediaDir,
	}
}

// Validate returns an error if the configuration cannot boot the application.
func (c Config) Validate() error {
	if c.DatabaseURL == "" {
		return ErrMissingDatabaseURL
	}
	if len(c.JWTSecret) < minJWTSecretSize {
		return ErrWeakJWTSecret
	}
	if _, banned := forbiddenJWTSecrets[c.JWTSecret]; banned {
		return ErrWeakJWTSecret
	}
	if isTrivialSecret(c.JWTSecret) {
		return ErrWeakJWTSecret
	}
	return nil
}
