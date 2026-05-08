package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadReadsConfigEnvAndYaml(t *testing.T) {
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	unsetEnv(t, "APP_HOST")
	unsetEnv(t, "POSTGRES_HOST")
	unsetEnv(t, "POSTGRES_PORT")
	unsetEnv(t, "POSTGRES_USER")
	unsetEnv(t, "POSTGRES_PASSWORD")
	unsetEnv(t, "POSTGRES_DB")
	unsetEnv(t, "POSTGRES_SSLMODE")

	dir := t.TempDir()
	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
		viper.Reset()
	})

	if err := os.Mkdir(filepath.Join(dir, "configs"), 0o755); err != nil {
		t.Fatal(err)
	}

	configYAML := []byte(`app:
  port: "9090"
  shutdown_timeout: "10s"
  wallet_queue_size: "32"
db:
  max_conns: "4"
  min_conns: "1"
`)
	if err := os.WriteFile(filepath.Join(dir, "configs", "config.yaml"), configYAML, 0o644); err != nil {
		t.Fatal(err)
	}

	configEnv := []byte(`APP_HOST=127.0.0.1
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=wallet
POSTGRES_PASSWORD=secret
POSTGRES_DB=wallet_test
POSTGRES_SSLMODE=disable
`)
	envPath := filepath.Join(dir, "config.env")
	if err := os.WriteFile(envPath, configEnv, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	viper.Reset()

	if err := Load(envPath); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := HTTPAddr(), "127.0.0.1:9090"; got != want {
		t.Fatalf("HTTPAddr() = %q, want %q", got, want)
	}

	wantDSN := "host=postgres port=5432 user=wallet password=secret dbname=wallet_test sslmode=disable"
	if got := DSN(); got != wantDSN {
		t.Fatalf("DSN() = %q, want %q", got, wantDSN)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	value, ok := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if ok {
			_ = os.Setenv(key, value)
			return
		}
		_ = os.Unsetenv(key)
	})
}
