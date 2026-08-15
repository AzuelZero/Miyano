package config

import "testing"

func TestLoadDefaultsWhenUnset(t *testing.T) {
	t.Setenv("PORT", "")
	got := Load()
	if got.Port != defaultPort {
		t.Errorf("Load().Port = %q, want %q", got.Port, defaultPort)
	}
}

func TestLoadUsesEnvWhenSet(t *testing.T) {
	t.Setenv("PORT", "9090")
	got := Load()
	if got.Port != "9090" {
		t.Errorf("Load().Port = %q, want %q", got.Port, "9090")
	}
}
