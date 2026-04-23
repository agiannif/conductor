package config_test

import (
	"os"
	"testing"

	"conductor/internal/config"
)

func TestDefaults(t *testing.T) {
	os.Unsetenv("CONDUCTOR_DB_PATH")
	os.Unsetenv("CONDUCTOR_SECURE_COOKIE")

	cfg := config.Load()

	if cfg.DBPath != "/data/conductor.db" {
		t.Errorf("want default DB path, got %q", cfg.DBPath)
	}
	if !cfg.SecureCookie {
		t.Error("want SecureCookie=true by default")
	}
}

func TestOverrides(t *testing.T) {
	os.Setenv("CONDUCTOR_DB_PATH", "/tmp/test.db")
	os.Setenv("CONDUCTOR_SECURE_COOKIE", "false")
	t.Cleanup(func() {
		os.Unsetenv("CONDUCTOR_DB_PATH")
		os.Unsetenv("CONDUCTOR_SECURE_COOKIE")
	})

	cfg := config.Load()

	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("want /tmp/test.db, got %q", cfg.DBPath)
	}
	if cfg.SecureCookie {
		t.Error("want SecureCookie=false when env is 'false'")
	}
}
