package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration values
type Config struct {
	Server     ServerConfig     `json:"server"`
	HuggingFace HuggingFaceConfig `json:"hugging_face"`
	Logger     LoggerConfig     `json:"logger"`
	Database   DatabaseConfig   `json:"database,omitempty"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port         int           `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	GracefulShutdownTimeout time.Duration `json:"graceful_shutdown_timeout"`
}

// HuggingFaceConfig holds Hugging Face API configuration
type HuggingFaceConfig struct {
	APIKey         string        `json:"-"` // Hidden in JSON for security
	BaseURL        string        `json:"base_url"`
	DefaultModel   string        `json:"default_model"`
	Timeout        time.Duration `json:"timeout"`
	RetryAttempts  int           `json:"retry_attempts"`
	RetryDelay     time.Duration `json:"retry_delay"`
	MaxTokens      int           `json:"max_tokens"`
	Temperature    float32       `json:"temperature"`
	RateLimitRPM   int           `json:"rate_limit_rpm"`
	RateLimitTPM   int           `json:"rate_limit_tpm"`
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	Structured bool   `json:"structured"`
}

// DatabaseConfig holds database configuration (optional for this project)
type DatabaseConfig struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"-"` // Hidden in JSON for security
}

// LoadConfig loads configuration from environment variables and defaults
func LoadConfig() (*Config, error) {
	config := &Config{}

	// Server configuration
	config.Server = ServerConfig{
		Port:                    getEnvAsInt("SERVER_PORT", 8080),
		Host:                    getEnv("SERVER_HOST", "localhost"),
		ReadTimeout:             getEnvAsDuration("SERVER_READ_TIMEOUT", "30s"),
		WriteTimeout:            getEnvAsDuration("SERVER_WRITE_TIMEOUT", "30s"),
		IdleTimeout:             getEnvAsDuration("SERVER_IDLE_TIMEOUT", "60s"),
		GracefulShutdownTimeout: getEnvAsDuration("SERVER_GRACEFUL_SHUTDOWN_TIMEOUT", "30s"),
	}

	// Hugging Face configuration
	apiKey := getEnv("HUGGINGFACE_API_KEY", "")
	if apiKey == "" {
		return nil, fmt.Errorf("HUGGINGFACE_API_KEY environment variable is required")
	}

	config.HuggingFace = HuggingFaceConfig{
		APIKey:        apiKey,
		BaseURL:       getEnv("HUGGINGFACE_BASE_URL", "https://api-inference.huggingface.co"),
		DefaultModel:  getEnv("HUGGINGFACE_DEFAULT_MODEL", "gpt2"),
		Timeout:       getEnvAsDuration("HUGGINGFACE_TIMEOUT", "30s"),
		RetryAttempts: getEnvAsInt("HUGGINGFACE_RETRY_ATTEMPTS", 3),
		RetryDelay:    getEnvAsDuration("HUGGINGFACE_RETRY_DELAY", "1s"),
		MaxTokens:     getEnvAsInt("HUGGINGFACE_MAX_TOKENS", 100),
		Temperature:   getEnvAsFloat32("HUGGINGFACE_TEMPERATURE", 0.7),
		RateLimitRPM:  getEnvAsInt("HUGGINGFACE_RATE_LIMIT_RPM", 60),
		RateLimitTPM:  getEnvAsInt("HUGGINGFACE_RATE_LIMIT_TPM", 10000),
	}

	// Logger configuration
	config.Logger = LoggerConfig{
		Level:      getEnv("LOG_LEVEL", "info"),
		Format:     getEnv("LOG_FORMAT", "json"),
		Output:     getEnv("LOG_OUTPUT", "stdout"),
		Structured: getEnvAsBool("LOG_STRUCTURED", true),
	}

	// Database configuration (optional)
	if getEnv("DATABASE_DRIVER", "") != "" {
		config.Database = DatabaseConfig{
			Driver:   getEnv("DATABASE_DRIVER", ""),
			Host:     getEnv("DATABASE_HOST", "localhost"),
			Port:     getEnvAsInt("DATABASE_PORT", 5432),
			Database: getEnv("DATABASE_NAME", ""),
			Username: getEnv("DATABASE_USERNAME", ""),
			Password: getEnv("DATABASE_PASSWORD", ""),
		}
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.HuggingFace.APIKey == "" {
		return fmt.Errorf("hugging face API key is required")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	if c.HuggingFace.MaxTokens <= 0 {
		return fmt.Errorf("max tokens must be positive")
	}
	if c.HuggingFace.Temperature < 0 || c.HuggingFace.Temperature > 1 {
		return fmt.Errorf("temperature must be between 0 and 1")
	}
	return nil
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat32(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return time.Second * 30 // fallback
}