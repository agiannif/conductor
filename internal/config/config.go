package config

import "os"

type Config struct {
	DBPath       string
	SecureCookie bool
}

func Load() Config {
	cfg := Config{
		DBPath:       "/data/conductor.db",
		SecureCookie: true,
	}
	if v := os.Getenv("CONDUCTOR_DB_PATH"); v != "" {
		cfg.DBPath = v
	}
	if os.Getenv("CONDUCTOR_SECURE_COOKIE") == "false" {
		cfg.SecureCookie = false
	}
	return cfg
}
