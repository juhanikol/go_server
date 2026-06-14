package serverapp

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func chdirTemp(t *testing.T) string {
	t.Helper()

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	return tempDir
}

func TestLoadAndMergeConfigMalformedJSONReturnsError(t *testing.T) {
	dir := chdirTemp(t)

	if err := os.WriteFile(filepath.Join(dir, "server.json"), []byte(`{"server_address":`), 0644); err != nil {
		t.Fatalf("write malformed config: %v", err)
	}

	_, err := loadAndMergeConfig(AppConfig{})
	if err == nil {
		t.Fatal("expected malformed JSON to return an error")
	}
	if !strings.Contains(err.Error(), "parse server.json") {
		t.Fatalf("expected parse error to include config path, got %q", err.Error())
	}
}

func TestLoadAndMergeConfigAcceptsStringLogLevel(t *testing.T) {
	dir := chdirTemp(t)

	configJSON := []byte(`{
		"server_address": ":8081",
		"env": "test",
		"root_path": "./",
		"read_timeout_sec": 15,
		"write_timeout_sec": 20,
		"shutdown_timeout_sec": 5,
		"log_file_name": "app.log",
		"log_max_size_mb": 12,
		"log_level": "INFO",
		"template_dir": "web/templates",
		"static_dir": "web/static"
	}`)
	if err := os.WriteFile(filepath.Join(dir, "server.json"), configJSON, 0644); err != nil {
		t.Fatalf("write valid config: %v", err)
	}

	config, err := loadAndMergeConfig(AppConfig{
		LogLevel: slog.LevelDebug,
	})
	if err != nil {
		t.Fatalf("expected valid config to load: %v", err)
	}

	if config.LogLevel != slog.LevelInfo {
		t.Fatalf("expected INFO log level, got %v", config.LogLevel)
	}
	if config.LogFileName != "app.log" {
		t.Fatalf("expected log file name to merge, got %q", config.LogFileName)
	}
	if config.LogMaxSizeMB != 12 {
		t.Fatalf("expected log max size to merge, got %d", config.LogMaxSizeMB)
	}
	if config.TemplateDir != "web/templates" {
		t.Fatalf("expected template dir to merge, got %q", config.TemplateDir)
	}
}

func TestLoadAndMergeConfigMissingFileIsNonFatal(t *testing.T) {
	chdirTemp(t)

	config, err := loadAndMergeConfig(AppConfig{
		ServerAddress: ":8081",
		TemplateDir:   "web/templates",
	})
	if err != nil {
		t.Fatalf("expected missing config to be non-fatal: %v", err)
	}

	if config.Env != "development" {
		t.Fatalf("expected default env, got %q", config.Env)
	}
	if config.LogFileName != "goserver.log" {
		t.Fatalf("expected default log file name, got %q", config.LogFileName)
	}
	if config.LogMaxSizeMB != 10 {
		t.Fatalf("expected default log max size, got %d", config.LogMaxSizeMB)
	}
}
