package config

import (
	"errors"
	"strings"
	"testing"
)

const testJWTSecret = "test-secret-that-is-long-enough-32ch"

func TestLoadDefaultsWhenUnset(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "")
	got := Load()
	if got.Port != defaultPort {
		t.Errorf("Load().Port = %q, want %q", got.Port, defaultPort)
	}
	if err := got.Validate(); !errors.Is(err, ErrMissingDatabaseURL) {
		t.Errorf("Validate() = %v, want ErrMissingDatabaseURL", err)
	}
}

func TestLoadUsesEnvWhenSet(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://miyano:dev@localhost:5432/miyano")
	t.Setenv("JWT_SECRET", testJWTSecret)
	got := Load()
	if got.Port != "9090" {
		t.Errorf("Load().Port = %q, want %q", got.Port, "9090")
	}
	if got.JWTSecret != testJWTSecret {
		t.Errorf("Load().JWTSecret = %q, want the env value", got.JWTSecret)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestValidateRequiresStrongJWTSecret(t *testing.T) {
	validDB := "postgres://miyano:dev@localhost:5432/miyano"

	cases := []struct {
		name   string
		secret string
	}{
		{"absent", ""},
		{"too short", strings.Repeat("a", 31)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{Port: defaultPort, DatabaseURL: validDB, JWTSecret: tc.secret}
			if err := cfg.Validate(); !errors.Is(err, ErrWeakJWTSecret) {
				t.Errorf("Validate() with secret %q = %v, want ErrWeakJWTSecret", tc.secret, err)
			}
		})
	}

	cfg := Config{Port: defaultPort, DatabaseURL: validDB, JWTSecret: strings.Repeat("a", 32)}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() with 32-char secret = %v, want nil", err)
	}
}
