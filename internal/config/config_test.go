package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Set required environment variable
	os.Setenv("HUGGINGFACE_API_KEY", "test-api-key")
	defer os.Unsetenv("HUGGINGFACE_API_KEY")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() unexpected error = %v", err)
	}

	// Test default values
	if config.Server.Port != 8080 {
		t.Errorf("Server.Port = %v, want %v", config.Server.Port, 8080)
	}
	if config.Server.Host != "localhost" {
		t.Errorf("Server.Host = %v, want %v", config.Server.Host, "localhost")
	}
	if config.HuggingFace.APIKey != "test-api-key" {
		t.Errorf("HuggingFace.APIKey = %v, want %v", config.HuggingFace.APIKey, "test-api-key")
	}
	if config.HuggingFace.DefaultModel != "gpt2" {
		t.Errorf("HuggingFace.DefaultModel = %v, want %v", config.HuggingFace.DefaultModel, "gpt2")
	}
	if config.Logger.Level != "info" {
		t.Errorf("Logger.Level = %v, want %v", config.Logger.Level, "info")
	}
}

func TestLoadConfigMissingAPIKey(t *testing.T) {
	// Unset API key
	os.Unsetenv("HUGGINGFACE_API_KEY")

	_, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() expected error for missing API key")
	}
	
	expectedMsg := "HUGGINGFACE_API_KEY environment variable is required"
	if err.Error() != expectedMsg {
		t.Errorf("LoadConfig() error = %v, want %v", err.Error(), expectedMsg)
	}
}

func TestLoadConfigWithCustomValues(t *testing.T) {
	// Set custom environment variables
	envVars := map[string]string{
		"HUGGINGFACE_API_KEY":            "custom-api-key",
		"SERVER_PORT":                    "9000",
		"SERVER_HOST":                    "0.0.0.0",
		"SERVER_READ_TIMEOUT":            "60s",
		"SERVER_WRITE_TIMEOUT":           "45s",
		"HUGGINGFACE_BASE_URL":           "https://custom-api.huggingface.co",
		"HUGGINGFACE_DEFAULT_MODEL":      "gpt2-large",
		"HUGGINGFACE_TIMEOUT":            "60s",
		"HUGGINGFACE_RETRY_ATTEMPTS":     "5",
		"HUGGINGFACE_MAX_TOKENS":         "200",
		"HUGGINGFACE_TEMPERATURE":        "0.9",
		"LOG_LEVEL":                      "debug",
		"LOG_FORMAT":                     "plain",
		"LOG_STRUCTURED":                 "false",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
	}
	defer func() {
		for k := range envVars {
			os.Unsetenv(k)
		}
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() unexpected error = %v", err)
	}

	// Verify custom values
	if config.Server.Port != 9000 {
		t.Errorf("Server.Port = %v, want %v", config.Server.Port, 9000)
	}
	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %v, want %v", config.Server.Host, "0.0.0.0")
	}
	if config.Server.ReadTimeout != 60*time.Second {
		t.Errorf("Server.ReadTimeout = %v, want %v", config.Server.ReadTimeout, 60*time.Second)
	}
	if config.HuggingFace.APIKey != "custom-api-key" {
		t.Errorf("HuggingFace.APIKey = %v, want %v", config.HuggingFace.APIKey, "custom-api-key")
	}
	if config.HuggingFace.DefaultModel != "gpt2-large" {
		t.Errorf("HuggingFace.DefaultModel = %v, want %v", config.HuggingFace.DefaultModel, "gpt2-large")
	}
	if config.HuggingFace.MaxTokens != 200 {
		t.Errorf("HuggingFace.MaxTokens = %v, want %v", config.HuggingFace.MaxTokens, 200)
	}
	if config.HuggingFace.Temperature != 0.9 {
		t.Errorf("HuggingFace.Temperature = %v, want %v", config.HuggingFace.Temperature, 0.9)
	}
	if config.Logger.Level != "debug" {
		t.Errorf("Logger.Level = %v, want %v", config.Logger.Level, "debug")
	}
	if config.Logger.Structured != false {
		t.Errorf("Logger.Structured = %v, want %v", config.Logger.Structured, false)
	}
}

func TestLoadConfigWithDatabase(t *testing.T) {
	envVars := map[string]string{
		"HUGGINGFACE_API_KEY": "test-api-key",
		"DATABASE_DRIVER":     "postgres",
		"DATABASE_HOST":       "localhost",
		"DATABASE_PORT":       "5433",
		"DATABASE_NAME":       "testdb",
		"DATABASE_USERNAME":   "testuser",
		"DATABASE_PASSWORD":   "testpass",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
	}
	defer func() {
		for k := range envVars {
			os.Unsetenv(k)
		}
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() unexpected error = %v", err)
	}

	if config.Database.Driver != "postgres" {
		t.Errorf("Database.Driver = %v, want %v", config.Database.Driver, "postgres")
	}
	if config.Database.Port != 5433 {
		t.Errorf("Database.Port = %v, want %v", config.Database.Port, 5433)
	}
	if config.Database.Database != "testdb" {
		t.Errorf("Database.Database = %v, want %v", config.Database.Database, "testdb")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Server: ServerConfig{
					Port: 8080,
				},
				HuggingFace: HuggingFaceConfig{
					APIKey:      "test-key",
					MaxTokens:   100,
					Temperature: 0.7,
				},
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: Config{
				Server: ServerConfig{
					Port: 8080,
				},
				HuggingFace: HuggingFaceConfig{
					APIKey:      "",
					MaxTokens:   100,
					Temperature: 0.7,
				},
			},
			wantErr: true,
			errMsg:  "hugging face API key is required",
		},
		{
			name: "invalid port - zero",
			config: Config{
				Server: ServerConfig{
					Port: 0,
				},
				HuggingFace: HuggingFaceConfig{
					APIKey:      "test-key",
					MaxTokens:   100,
					Temperature: 0.7,
				},
			},
			wantErr: true,
			errMsg:  "invalid server port: 0",
		},
		{
			name: "invalid port - negative",
			config: Config{
				Server: ServerConfig{
					Port: -1,
				},
				HuggingFace: HuggingFaceConfig{
					APIKey:      "test-key",
					MaxTokens:   100,
					Temperature: 0.7,
				},
			},
			wantErr: true,
			errMsg:  "invalid server port: -1",
		},
		{
			name: "invalid port - too high",
			config: Config{
				Server: ServerConfig{
					Port: 70000,
				},
				HuggingFace: HuggingFaceConfig{
					APIKey:      "test-key",
					MaxTokens:   100,
					Temperature: 0.7,
				},
			},
			wantErr: true,
			errMsg:  "invalid server port: 70000",
		},
		{
			name: "invalid max tokens",
			config: Config{
				Server: ServerConfig{
					Port: 8080,
				},
				HuggingFace: HuggingFaceConfig{
					APIKey:      "test-key",
					MaxTokens:   -1,
					Temperature: 0.7,
				},
			},
			wantErr: true,
			errMsg:  "max tokens must be positive",
		},
		{
			name: "invalid temperature - too low",
			config: Config{
				Server: ServerConfig{
					Port: 8080,
				},
				HuggingFace: HuggingFaceConfig{
					APIKey:      "test-key",
					MaxTokens:   100,
					Temperature: -0.1,
				},
			},
			wantErr: true,
			errMsg:  "temperature must be between 0 and 1",
		},
		{
			name: "invalid temperature - too high",
			config: Config{
				Server: ServerConfig{
					Port: 8080,
				},
				HuggingFace: HuggingFaceConfig{
					APIKey:      "test-key",
					MaxTokens:   100,
					Temperature: 1.1,
				},
			},
			wantErr: true,
			errMsg:  "temperature must be between 0 and 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Config.Validate() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Config.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Config.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		want         string
	}{
		{
			name:         "env var exists",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "custom",
			want:         "custom",
		},
		{
			name:         "env var does not exist",
			key:          "NONEXISTENT_KEY",
			defaultValue: "default",
			envValue:     "",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnv(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		want         int
	}{
		{
			name:         "valid int",
			key:          "TEST_INT",
			defaultValue: 10,
			envValue:     "20",
			want:         20,
		},
		{
			name:         "invalid int",
			key:          "TEST_INT",
			defaultValue: 10,
			envValue:     "invalid",
			want:         10,
		},
		{
			name:         "no env var",
			key:          "NONEXISTENT_INT",
			defaultValue: 10,
			envValue:     "",
			want:         10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvAsInt(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvAsInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvAsBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		want         bool
	}{
		{
			name:         "true value",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "true",
			want:         true,
		},
		{
			name:         "false value",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "false",
			want:         false,
		},
		{
			name:         "invalid value",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "invalid",
			want:         true,
		},
		{
			name:         "no env var",
			key:          "NONEXISTENT_BOOL",
			defaultValue: false,
			envValue:     "",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvAsBool(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvAsBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		want         time.Duration
	}{
		{
			name:         "valid duration",
			key:          "TEST_DURATION",
			defaultValue: "10s",
			envValue:     "30s",
			want:         30 * time.Second,
		},
		{
			name:         "invalid duration",
			key:          "TEST_DURATION",
			defaultValue: "10s",
			envValue:     "invalid",
			want:         10 * time.Second,
		},
		{
			name:         "no env var",
			key:          "NONEXISTENT_DURATION",
			defaultValue: "15s",
			envValue:     "",
			want:         15 * time.Second,
		},
		{
			name:         "invalid default fallback",
			key:          "TEST_DURATION",
			defaultValue: "invalid",
			envValue:     "",
			want:         30 * time.Second, // fallback value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvAsDuration(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvAsDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvAsFloat32(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue float32
		envValue     string
		want         float32
	}{
		{
			name:         "valid float",
			key:          "TEST_FLOAT",
			defaultValue: 0.5,
			envValue:     "0.8",
			want:         0.8,
		},
		{
			name:         "invalid float",
			key:          "TEST_FLOAT",
			defaultValue: 0.5,
			envValue:     "invalid",
			want:         0.5,
		},
		{
			name:         "no env var",
			key:          "NONEXISTENT_FLOAT",
			defaultValue: 0.7,
			envValue:     "",
			want:         0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvAsFloat32(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvAsFloat32() = %v, want %v", got, tt.want)
			}
		})
	}
}