package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Point the config somewhere that doesn't exist so we fall through to defaults.
	t.Setenv("HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.DefaultModel != DefaultModel {
		t.Errorf("DefaultModel = %q, want %q", cfg.DefaultModel, DefaultModel)
	}
	if cfg.Region != DefaultRegion {
		t.Errorf("Region = %q, want %q", cfg.Region, DefaultRegion)
	}
	if cfg.MaxTokens != 4096 {
		t.Errorf("MaxTokens = %d, want 4096", cfg.MaxTokens)
	}
}

func TestEnvOverride(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("BEDROCK_CLI_REGION", "eu-west-1")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg.Region != "eu-west-1" {
		t.Errorf("Region = %q, want %q via env override", cfg.Region, "eu-west-1")
	}
	os.Unsetenv("BEDROCK_CLI_REGION")
}

func TestDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error: %v", err)
	}
	if dir == "" {
		t.Error("Dir() returned empty string")
	}

	// Must exist after the call.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Dir() did not create directory at %s", dir)
	}
}

func TestCacheDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir, err := CacheDir()
	if err != nil {
		t.Fatalf("CacheDir() error: %v", err)
	}
	if dir == "" {
		t.Error("CacheDir() returned empty string")
	}
}
