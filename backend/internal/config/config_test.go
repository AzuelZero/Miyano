package config

import (
	"errors"
	"testing"
)

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
	got := Load()
	if got.Port != "9090" {
		t.Errorf("Load().Port = %q, want %q", got.Port, "9090")
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}
