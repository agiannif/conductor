package config

import (
	"os"
	"strings"
)

// defaultsFilePath is the path to the env-var defaults file (overridable in tests).
var defaultsFilePath = "/etc/default/conductor"

type Config struct {
	DBPath       string
	SecureCookie bool
}

func Load() Config {
	cfg := Config{
		DBPath:       "/data/conductor.db",
		SecureCookie: true,
	}

	defaults := readDefaultsFile(defaultsFilePath)

	if v := os.Getenv("CONDUCTOR_DB_PATH"); v != "" {
		cfg.DBPath = v
	} else if v := defaults["CONDUCTOR_DB_PATH"]; v != "" {
		cfg.DBPath = v
	}

	if os.Getenv("CONDUCTOR_SECURE_COOKIE") == "false" {
		cfg.SecureCookie = false
	} else if defaults["CONDUCTOR_SECURE_COOKIE"] == "false" {
		cfg.SecureCookie = false
	}

	return cfg
}

// readDefaultsFile parses a KEY=VALUE file (shell-style, comments and blank
// lines ignored, optional "export " prefix and surrounding quotes stripped).
func readDefaultsFile(path string) map[string]string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	vars := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		vars[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"'`)
	}
	return vars
}
