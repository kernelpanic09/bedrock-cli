package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	DefaultModel  = "sonnet"
	DefaultRegion = "us-east-1"
)

// Config holds all user-facing configuration values.
type Config struct {
	// DefaultModel is the model alias used when none is specified (e.g. "sonnet").
	DefaultModel string `mapstructure:"default-model"`
	// Region is the AWS region for Bedrock API calls.
	Region string `mapstructure:"region"`
	// MaxTokens caps the number of tokens the model may return per response.
	MaxTokens int `mapstructure:"max-tokens"`
	// Temperature controls sampling randomness (0.0 = deterministic, 1.0 = maximum).
	Temperature float64 `mapstructure:"temperature"`
	// CacheTTL is the response cache lifetime in seconds; 0 = cache forever, -1 = disabled.
	CacheTTL int `mapstructure:"cache-ttl"`
	// NoColor suppresses ANSI color codes in terminal output.
	NoColor bool `mapstructure:"no-color"`
	// NoStream buffers the full response before printing instead of streaming tokens.
	NoStream bool `mapstructure:"no-stream"`
	// ShowCost prints the estimated invocation cost after each response.
	ShowCost bool `mapstructure:"show-cost"`
}

// Dir returns the config directory, creating it if needed.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving config dir: %w", err)
	}
	dir := filepath.Join(base, "bedrock-cli")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("creating config dir %s: %w", dir, err)
	}
	return dir, nil
}

// TemplatesDir returns the directory where user templates are stored.
func TemplatesDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	tdir := filepath.Join(dir, "templates")
	if err := os.MkdirAll(tdir, 0o700); err != nil {
		return "", fmt.Errorf("creating templates dir: %w", err)
	}
	return tdir, nil
}

// CacheDir returns the cache directory for responses and the usage database.
func CacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolving cache dir: %w", err)
	}
	dir := filepath.Join(base, "bedrock-cli")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("creating cache dir %s: %w", dir, err)
	}
	return dir, nil
}

// ResponseCacheDir returns the subdirectory used for cached response files.
func ResponseCacheDir() (string, error) {
	base, err := CacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "responses")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("creating response cache dir: %w", err)
	}
	return dir, nil
}

// Load reads the config file and returns a Config, applying defaults.
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	viper.AddConfigPath(dir)

	// Env var overrides - BEDROCK_CLI_DEFAULT_MODEL etc.
	viper.SetEnvPrefix("BEDROCK_CLI")
	viper.AutomaticEnv()

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		// No config file is fine - we use defaults.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// Save writes the in-memory viper config back to disk.
func Save() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "config.toml")
	return viper.WriteConfigAs(path)
}

// Set updates a single key and persists the config.
func Set(key, value string) error {
	viper.Set(key, value)
	return Save()
}

// Get returns the string value for a config key.
func Get(key string) string {
	return viper.GetString(key)
}

// Path returns the path to the config file.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

func setDefaults() {
	viper.SetDefault("default-model", DefaultModel)
	viper.SetDefault("region", DefaultRegion)
	viper.SetDefault("max-tokens", 4096)
	viper.SetDefault("temperature", 0.7)
	viper.SetDefault("cache-ttl", 0)
	viper.SetDefault("no-color", false)
	viper.SetDefault("no-stream", false)
	viper.SetDefault("show-cost", true)
}
