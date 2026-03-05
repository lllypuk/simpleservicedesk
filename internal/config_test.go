package internal_test

import (
	"encoding/base64"
	"os"
	"testing"
	"time"

	"simpleservicedesk/internal"
	"simpleservicedesk/pkg/environment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"ENV_TYPE",
		"SERVER_PORT",
		"INTERRUPT_TIMEOUT",
		"READ_HEADER_TIMEOUT",
		"CORS_ALLOWED_ORIGINS",
		"RATE_LIMIT_RPS",
		"MONGO_URI",
		"MONGO_DATABASE",
		"JWT_SECRET",
		"JWT_EXPIRATION",
		"BOOTSTRAP_ADMIN_NAME",
		"BOOTSTRAP_ADMIN_EMAIL",
		"BOOTSTRAP_ADMIN_PASSWORD",
	}

	for _, key := range envVars {
		if val, exists := os.LookupEnv(key); exists {
			originalEnv[key] = val
		}
		os.Unsetenv(key)
	}

	// Restore environment after test
	defer func() {
		for _, key := range envVars {
			os.Unsetenv(key)
		}
		for key, val := range originalEnv {
			t.Setenv(key, val)
		}
	}()

	t.Run("default configuration", func(t *testing.T) {
		config, err := internal.LoadConfig()
		require.NoError(t, err)

		// Test server defaults
		assert.Equal(t, environment.Testing, config.Server.Environment)
		assert.Equal(t, "8080", config.Server.Port)
		assert.Equal(t, 2*time.Second, config.Server.InterruptTimeout)
		assert.Equal(t, 5*time.Second, config.Server.ReadHeaderTimeout)
		assert.Equal(t, []string{"*"}, config.Server.CORSAllowedOrigins)
		assert.Equal(t, 100, config.Server.RateLimitRPS)

		// Test mongo defaults
		assert.Equal(t, "mongodb://localhost:27017", config.Mongo.URI)
		assert.Equal(t, "servicedesk", config.Mongo.Database)

		// Test auth defaults
		assert.NotEmpty(t, config.Auth.JWTSigningKey)
		assert.Equal(t, 24*time.Hour, config.Auth.JWTExpiration)
	})

	t.Run("production requires jwt secret", func(t *testing.T) {
		t.Setenv("ENV_TYPE", "production")
		t.Setenv("JWT_SECRET", "")

		_, err := internal.LoadConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "jwt secret is required in production environment")
	})

	t.Run("custom configuration from environment", func(t *testing.T) {
		// Set custom environment variables
		t.Setenv("ENV_TYPE", "production")
		t.Setenv("SERVER_PORT", "9090")
		t.Setenv("INTERRUPT_TIMEOUT", "10s")
		t.Setenv("READ_HEADER_TIMEOUT", "30s")
		t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com,https://admin.example.com")
		t.Setenv("RATE_LIMIT_RPS", "150")
		t.Setenv("MONGO_URI", "mongodb://custom-host:27017")
		t.Setenv("MONGO_DATABASE", "custom_db")
		t.Setenv("JWT_SECRET", "custom-jwt-secret")
		t.Setenv("JWT_EXPIRATION", "12h")
		t.Setenv("BOOTSTRAP_ADMIN_NAME", "Bootstrap Root")
		t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "root@example.com")
		t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "bootstrap-password")

		config, err := internal.LoadConfig()
		require.NoError(t, err)

		// Test server custom values
		assert.Equal(t, environment.Type("production"), config.Server.Environment)
		assert.Equal(t, "9090", config.Server.Port)
		assert.Equal(t, 10*time.Second, config.Server.InterruptTimeout)
		assert.Equal(t, 30*time.Second, config.Server.ReadHeaderTimeout)
		assert.Equal(
			t,
			[]string{"https://app.example.com", "https://admin.example.com"},
			config.Server.CORSAllowedOrigins,
		)
		assert.Equal(t, 150, config.Server.RateLimitRPS)

		// Test mongo custom values
		assert.Equal(t, "mongodb://custom-host:27017", config.Mongo.URI)
		assert.Equal(t, "custom_db", config.Mongo.Database)

		// Test auth custom values
		assert.Equal(t, "custom-jwt-secret", config.Auth.JWTSigningKey)
		assert.Equal(t, 12*time.Hour, config.Auth.JWTExpiration)
		assert.Equal(t, "Bootstrap Root", config.Auth.BootstrapAdminName)
		assert.Equal(t, "root@example.com", config.Auth.BootstrapAdminEmail)
		assert.Equal(t, "bootstrap-password", config.Auth.BootstrapAdminPassword)
	})

	t.Run("invalid interrupt timeout", func(t *testing.T) {
		t.Setenv("INTERRUPT_TIMEOUT", "invalid-duration")

		_, err := internal.LoadConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not parse interrupt timeout")
	})

	t.Run("invalid read header timeout", func(t *testing.T) {
		// Set a valid interrupt timeout so we get to the read header timeout validation
		t.Setenv("INTERRUPT_TIMEOUT", "2s")
		t.Setenv("READ_HEADER_TIMEOUT", "invalid-duration")

		_, err := internal.LoadConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not parse read header timeout")
	})

	t.Run("invalid jwt expiration", func(t *testing.T) {
		t.Setenv("JWT_EXPIRATION", "invalid-duration")

		_, err := internal.LoadConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not parse jwt expiration")
	})
}

func TestLoadServer(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"ENV_TYPE",
		"SERVER_PORT",
		"INTERRUPT_TIMEOUT",
		"READ_HEADER_TIMEOUT",
		"CORS_ALLOWED_ORIGINS",
		"RATE_LIMIT_RPS",
	}

	for _, key := range envVars {
		if val, exists := os.LookupEnv(key); exists {
			originalEnv[key] = val
		}
		os.Unsetenv(key)
	}

	// Restore environment after test
	defer func() {
		for _, key := range envVars {
			os.Unsetenv(key)
		}
		for key, val := range originalEnv {
			t.Setenv(key, val)
		}
	}()

	tests := []struct {
		name          string
		envVars       map[string]string
		expectedError bool
		expected      internal.Server
	}{
		{
			name:    "default values",
			envVars: map[string]string{},
			expected: internal.Server{
				Environment:        environment.Testing,
				Port:               "8080",
				InterruptTimeout:   2 * time.Second,
				ReadHeaderTimeout:  5 * time.Second,
				CORSAllowedOrigins: []string{"*"},
				RateLimitRPS:       100,
			},
		},
		{
			name: "production environment",
			envVars: map[string]string{
				"ENV_TYPE": "production",
			},
			expected: internal.Server{
				Environment:        environment.Production,
				Port:               "8080",
				InterruptTimeout:   2 * time.Second,
				ReadHeaderTimeout:  5 * time.Second,
				CORSAllowedOrigins: []string{"*"},
				RateLimitRPS:       100,
			},
		},
		{
			name: "custom port",
			envVars: map[string]string{
				"SERVER_PORT": "3000",
			},
			expected: internal.Server{
				Environment:        environment.Testing,
				Port:               "3000",
				InterruptTimeout:   2 * time.Second,
				ReadHeaderTimeout:  5 * time.Second,
				CORSAllowedOrigins: []string{"*"},
				RateLimitRPS:       100,
			},
		},
		{
			name: "custom timeouts",
			envVars: map[string]string{
				"INTERRUPT_TIMEOUT":   "30s",
				"READ_HEADER_TIMEOUT": "60s",
			},
			expected: internal.Server{
				Environment:        environment.Testing,
				Port:               "8080",
				InterruptTimeout:   30 * time.Second,
				ReadHeaderTimeout:  60 * time.Second,
				CORSAllowedOrigins: []string{"*"},
				RateLimitRPS:       100,
			},
		},
		{
			name: "custom cors origins",
			envVars: map[string]string{
				"CORS_ALLOWED_ORIGINS": "https://app.example.com, https://admin.example.com",
			},
			expected: internal.Server{
				Environment:        environment.Testing,
				Port:               "8080",
				InterruptTimeout:   2 * time.Second,
				ReadHeaderTimeout:  5 * time.Second,
				CORSAllowedOrigins: []string{"https://app.example.com", "https://admin.example.com"},
				RateLimitRPS:       100,
			},
		},
		{
			name: "empty cors origins uses default wildcard",
			envVars: map[string]string{
				"CORS_ALLOWED_ORIGINS": "",
			},
			expected: internal.Server{
				Environment:        environment.Testing,
				Port:               "8080",
				InterruptTimeout:   2 * time.Second,
				ReadHeaderTimeout:  5 * time.Second,
				CORSAllowedOrigins: []string{"*"},
				RateLimitRPS:       100,
			},
		},
		{
			name: "custom rate limit rps",
			envVars: map[string]string{
				"RATE_LIMIT_RPS": "250",
			},
			expected: internal.Server{
				Environment:        environment.Testing,
				Port:               "8080",
				InterruptTimeout:   2 * time.Second,
				ReadHeaderTimeout:  5 * time.Second,
				CORSAllowedOrigins: []string{"*"},
				RateLimitRPS:       250,
			},
		},
		{
			name: "invalid interrupt timeout",
			envVars: map[string]string{
				"INTERRUPT_TIMEOUT": "not-a-duration",
			},
			expectedError: true,
		},
		{
			name: "invalid read header timeout",
			envVars: map[string]string{
				"READ_HEADER_TIMEOUT": "not-a-duration",
			},
			expectedError: true,
		},
		{
			name: "invalid rate limit rps",
			envVars: map[string]string{
				"RATE_LIMIT_RPS": "invalid",
			},
			expectedError: true,
		},
		{
			name: "non-positive rate limit rps",
			envVars: map[string]string{
				"RATE_LIMIT_RPS": "0",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			for _, key := range envVars {
				os.Unsetenv(key)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			server, err := internal.LoadServer()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, server)
			}
		})
	}
}

func TestLoadMongo(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{"MONGO_URI", "MONGO_DATABASE"}

	for _, key := range envVars {
		if val, exists := os.LookupEnv(key); exists {
			originalEnv[key] = val
		}
		os.Unsetenv(key)
	}

	// Restore environment after test
	defer func() {
		for _, key := range envVars {
			os.Unsetenv(key)
		}
		for key, val := range originalEnv {
			t.Setenv(key, val)
		}
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected internal.Mongo
	}{
		{
			name:    "default values",
			envVars: map[string]string{},
			expected: internal.Mongo{
				URI:      "mongodb://localhost:27017",
				Database: "servicedesk",
			},
		},
		{
			name: "custom URI",
			envVars: map[string]string{
				"MONGO_URI": "mongodb://prod-db:27017",
			},
			expected: internal.Mongo{
				URI:      "mongodb://prod-db:27017",
				Database: "servicedesk",
			},
		},
		{
			name: "custom database",
			envVars: map[string]string{
				"MONGO_DATABASE": "test_db",
			},
			expected: internal.Mongo{
				URI:      "mongodb://localhost:27017",
				Database: "test_db",
			},
		},
		{
			name: "both custom",
			envVars: map[string]string{
				"MONGO_URI":      "mongodb://remote:27017",
				"MONGO_DATABASE": "prod_servicedesk",
			},
			expected: internal.Mongo{
				URI:      "mongodb://remote:27017",
				Database: "prod_servicedesk",
			},
		},
		{
			name: "empty values use defaults",
			envVars: map[string]string{
				"MONGO_URI":      "",
				"MONGO_DATABASE": "",
			},
			expected: internal.Mongo{
				URI:      "mongodb://localhost:27017",
				Database: "servicedesk",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			for _, key := range envVars {
				os.Unsetenv(key)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				if value != "" {
					t.Setenv(key, value)
				}
			}

			mongo := internal.LoadMongo()
			assert.Equal(t, tt.expected, mongo)
		})
	}
}

func TestLoadAuth(t *testing.T) {
	originalEnv := make(map[string]string)
	envVars := []string{
		"JWT_SECRET",
		"JWT_EXPIRATION",
		"BOOTSTRAP_ADMIN_NAME",
		"BOOTSTRAP_ADMIN_EMAIL",
		"BOOTSTRAP_ADMIN_PASSWORD",
	}

	for _, key := range envVars {
		if val, exists := os.LookupEnv(key); exists {
			originalEnv[key] = val
		}
		os.Unsetenv(key)
	}

	defer func() {
		for _, key := range envVars {
			os.Unsetenv(key)
		}
		for key, val := range originalEnv {
			t.Setenv(key, val)
		}
	}()

	t.Run("default values", func(t *testing.T) {
		auth, err := internal.LoadAuth(environment.Testing)
		require.NoError(t, err)
		assert.NotEmpty(t, auth.JWTSigningKey)
		assert.Equal(t, 24*time.Hour, auth.JWTExpiration)

		_, decodeErr := base64.RawStdEncoding.DecodeString(auth.JWTSigningKey)
		require.NoError(t, decodeErr)
	})

	t.Run("custom values", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "custom-secret")
		t.Setenv("JWT_EXPIRATION", "6h")
		t.Setenv("BOOTSTRAP_ADMIN_NAME", "Bootstrap Root")
		t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "root@example.com")
		t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "bootstrap-password")

		auth, err := internal.LoadAuth(environment.Production)
		require.NoError(t, err)
		assert.Equal(t, "custom-secret", auth.JWTSigningKey)
		assert.Equal(t, 6*time.Hour, auth.JWTExpiration)
		assert.Equal(t, "Bootstrap Root", auth.BootstrapAdminName)
		assert.Equal(t, "root@example.com", auth.BootstrapAdminEmail)
		assert.Equal(t, "bootstrap-password", auth.BootstrapAdminPassword)
	})

	t.Run("empty secret generates default", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "")

		auth, err := internal.LoadAuth(environment.Testing)
		require.NoError(t, err)
		assert.NotEmpty(t, auth.JWTSigningKey)
	})

	t.Run("production requires explicit secret", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "")

		_, err := internal.LoadAuth(environment.Production)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "jwt secret is required in production environment")
	})

	t.Run("invalid expiration", func(t *testing.T) {
		t.Setenv("JWT_EXPIRATION", "bad-value")

		_, err := internal.LoadAuth(environment.Testing)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not parse jwt expiration")
	})

	t.Run("zero expiration is invalid", func(t *testing.T) {
		t.Setenv("JWT_EXPIRATION", "0s")

		_, err := internal.LoadAuth(environment.Testing)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "jwt expiration must be greater than zero")
	})

	t.Run("negative expiration is invalid", func(t *testing.T) {
		t.Setenv("JWT_EXPIRATION", "-1s")

		_, err := internal.LoadAuth(environment.Testing)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "jwt expiration must be greater than zero")
	})
}

func TestGetEnv(t *testing.T) {
	testKey := "TEST_GET_ENV_KEY"

	// Ensure key is not set initially
	os.Unsetenv(testKey)
	defer os.Unsetenv(testKey)

	t.Run("returns fallback when env var not set", func(t *testing.T) {
		result := internal.GetEnv(testKey, "fallback_value")
		assert.Equal(t, "fallback_value", result)
	})

	t.Run("returns env var when set", func(t *testing.T) {
		t.Setenv(testKey, "env_value")
		result := internal.GetEnv(testKey, "fallback_value")
		assert.Equal(t, "env_value", result)
	})

	t.Run("returns empty string when env var is empty", func(t *testing.T) {
		t.Setenv(testKey, "")
		result := internal.GetEnv(testKey, "fallback_value")
		assert.Empty(t, result)
	})
}

func TestConfigurationValidation(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"ENV_TYPE",
		"SERVER_PORT",
		"INTERRUPT_TIMEOUT",
		"READ_HEADER_TIMEOUT",
		"CORS_ALLOWED_ORIGINS",
		"RATE_LIMIT_RPS",
		"MONGO_URI",
		"MONGO_DATABASE",
		"JWT_SECRET",
		"JWT_EXPIRATION",
		"BOOTSTRAP_ADMIN_NAME",
		"BOOTSTRAP_ADMIN_EMAIL",
		"BOOTSTRAP_ADMIN_PASSWORD",
	}

	for _, key := range envVars {
		if val, exists := os.LookupEnv(key); exists {
			originalEnv[key] = val
		}
		os.Unsetenv(key)
	}

	// Restore environment after test
	defer func() {
		for _, key := range envVars {
			os.Unsetenv(key)
		}
		for key, val := range originalEnv {
			t.Setenv(key, val)
		}
	}()

	t.Run("valid configuration", func(t *testing.T) {
		testConfigurationValidation(t, map[string]string{}, false, "")
	})

	t.Run("zero timeout values are valid", func(t *testing.T) {
		envVars := map[string]string{
			"INTERRUPT_TIMEOUT":   "0s",
			"READ_HEADER_TIMEOUT": "0s",
		}
		testConfigurationValidation(t, envVars, false, "")
	})

	t.Run("negative timeout values are valid", func(t *testing.T) {
		envVars := map[string]string{
			"INTERRUPT_TIMEOUT": "-1s",
		}
		testConfigurationValidation(t, envVars, false, "")
	})

	t.Run("malformed duration format", func(t *testing.T) {
		envVars := map[string]string{
			"INTERRUPT_TIMEOUT": "5 seconds",
		}
		testConfigurationValidation(t, envVars, true, "could not parse interrupt timeout")
	})

	t.Run("non-numeric port is allowed", func(t *testing.T) {
		envVars := map[string]string{
			"SERVER_PORT": "not-a-port",
		}
		testConfigurationValidation(t, envVars, false, "")
	})
}

func testConfigurationValidation(t *testing.T, envVars map[string]string, expectErr bool, errMsg string) {
	// Clear environment
	configEnvVars := []string{
		"ENV_TYPE",
		"SERVER_PORT",
		"INTERRUPT_TIMEOUT",
		"READ_HEADER_TIMEOUT",
		"CORS_ALLOWED_ORIGINS",
		"RATE_LIMIT_RPS",
		"MONGO_URI",
		"MONGO_DATABASE",
		"JWT_SECRET",
		"JWT_EXPIRATION",
		"BOOTSTRAP_ADMIN_NAME",
		"BOOTSTRAP_ADMIN_EMAIL",
		"BOOTSTRAP_ADMIN_PASSWORD",
	}

	for _, key := range configEnvVars {
		os.Unsetenv(key)
	}

	// Set test environment variables
	for key, value := range envVars {
		t.Setenv(key, value)
	}

	config, err := internal.LoadConfig()

	if expectErr {
		require.Error(t, err)
		if errMsg != "" {
			assert.Contains(t, err.Error(), errMsg)
		}
	} else {
		require.NoError(t, err)
		assert.NotEmpty(t, config.Server.Port)
		assert.NotEmpty(t, config.Mongo.URI)
		assert.NotEmpty(t, config.Mongo.Database)
		assert.NotEmpty(t, config.Auth.JWTSigningKey)
		assert.NotZero(t, config.Auth.JWTExpiration)
	}
}

func TestEnvironmentTypeHandling(t *testing.T) {
	testKey := "ENV_TYPE"
	os.Unsetenv(testKey)
	defer os.Unsetenv(testKey)

	tests := []struct {
		name     string
		envValue string
		expected environment.Type
	}{
		{
			name:     "default when not set",
			envValue: "",
			expected: environment.Testing,
		},
		{
			name:     "production environment",
			envValue: "production",
			expected: environment.Production,
		},
		{
			name:     "testing environment",
			envValue: "testing",
			expected: environment.Testing,
		},
		{
			name:     "custom environment (valid but not predefined)",
			envValue: "staging",
			expected: environment.Type("staging"),
		},
		{
			name:     "case sensitive",
			envValue: "PRODUCTION",
			expected: environment.Type("PRODUCTION"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(testKey)
			if tt.envValue != "" {
				t.Setenv(testKey, tt.envValue)
			}

			server, err := internal.LoadServer()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, server.Environment)
		})
	}
}
