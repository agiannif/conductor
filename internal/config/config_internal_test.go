package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDefaultsFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[string]string
	}{
		{
			name:    "basic key=value",
			content: "FOO=bar\nBAZ=qux\n",
			want:    map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name:    "export prefix stripped",
			content: "export FOO=bar\n",
			want:    map[string]string{"FOO": "bar"},
		},
		{
			name:    "quoted values stripped",
			content: `FOO="bar baz"` + "\n" + "BAZ='qux'\n",
			want:    map[string]string{"FOO": "bar baz", "BAZ": "qux"},
		},
		{
			name:    "comments and blank lines ignored",
			content: "# comment\n\nFOO=bar\n",
			want:    map[string]string{"FOO": "bar"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.CreateTemp(t.TempDir(), "defaults")
			if err != nil {
				t.Fatal(err)
			}
			f.WriteString(tc.content)
			f.Close()

			got := readDefaultsFile(f.Name())
			for k, want := range tc.want {
				if got[k] != want {
					t.Errorf("key %s: got %q, want %q", k, got[k], want)
				}
			}
		})
	}
}

func TestReadDefaultsFileMissing(t *testing.T) {
	got := readDefaultsFile(filepath.Join(t.TempDir(), "nonexistent"))
	if got != nil {
		t.Errorf("expected nil for missing file, got %v", got)
	}
}

func TestLoadFromDefaultsFile(t *testing.T) {
	os.Unsetenv("CONDUCTOR_DB_PATH")
	os.Unsetenv("CONDUCTOR_SECURE_COOKIE")
	os.Unsetenv("CONDUCTOR_LISTEN_ADDR")
	t.Cleanup(func() { defaultsFilePath = "/etc/default/conductor" })

	f, err := os.CreateTemp(t.TempDir(), "defaults")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("CONDUCTOR_DB_PATH=/srv/conductor.db\nCONDUCTOR_SECURE_COOKIE=false\nCONDUCTOR_LISTEN_ADDR=127.0.0.1:8080\n")
	f.Close()

	defaultsFilePath = f.Name()
	cfg := Load()

	if cfg.DBPath != "/srv/conductor.db" {
		t.Errorf("got DBPath %q, want /srv/conductor.db", cfg.DBPath)
	}
	if cfg.SecureCookie {
		t.Error("want SecureCookie=false from defaults file")
	}
	if cfg.ListenAddr != "127.0.0.1:8080" {
		t.Errorf("got ListenAddr %q, want 127.0.0.1:8080", cfg.ListenAddr)
	}
}

func TestEnvVarTakesPrecedenceOverDefaultsFile(t *testing.T) {
	os.Setenv("CONDUCTOR_DB_PATH", "/env/override.db")
	t.Cleanup(func() {
		os.Unsetenv("CONDUCTOR_DB_PATH")
		defaultsFilePath = "/etc/default/conductor"
	})

	f, err := os.CreateTemp(t.TempDir(), "defaults")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("CONDUCTOR_DB_PATH=/file/should-be-ignored.db\n")
	f.Close()

	defaultsFilePath = f.Name()
	cfg := Load()

	if cfg.DBPath != "/env/override.db" {
		t.Errorf("env var should take precedence, got %q", cfg.DBPath)
	}
}
