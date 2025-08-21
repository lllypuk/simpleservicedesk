package internal_test

import (
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
		"MONGO_URI",
		"MONGO_DATABASE",
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

		// Test mongo defaults
		assert.Equal(t, "mongodb://localhost:27017", config.Mongo.URI)
		assert.Equal(t, "servicedesk", config.Mongo.Database)
	})

	t.Run("custom configuration from environment", func(t *testing.T) {
		// Set custom environment variables
		t.Setenv("ENV_TYPE", "production")
		t.Setenv("SERVER_PORT", "9090")
		t.Setenv("INTERRUPT_TIMEOUT", "10s")
		t.Setenv("READ_HEADER_TIMEOUT", "30s")
		t.Setenv("MONGO_URI", "mongodb://custom-host:27017")
		t.Setenv("MONGO_DATABASE", "custom_db")

		config, err := internal.LoadConfig()
		require.NoError(t, err)

		// Test server custom values
		assert.Equal(t, environment.Type("production"), config.Server.Environment)
		assert.Equal(t, "9090", config.Server.Port)
		assert.Equal(t, 10*time.Second, config.Server.InterruptTimeout)
		assert.Equal(t, 30*time.Second, config.Server.ReadHeaderTimeout)

		// Test mongo custom values
		assert.Equal(t, "mongodb://custom-host:27017", config.Mongo.URI)
		assert.Equal(t, "custom_db", config.Mongo.Database)
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
}

func TestLoadServer(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"ENV_TYPE",
		"SERVER_PORT",
		"INTERRUPT_TIMEOUT",
		"READ_HEADER_TIMEOUT",
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
				Environment:       environment.Testing,
				Port:              "8080",
				InterruptTimeout:  2 * time.Second,
				ReadHeaderTimeout: 5 * time.Second,
			},
		},
		{
			name: "production environment",
			envVars: map[string]string{
				"ENV_TYPE": "production",
			},
			expected: internal.Server{
				Environment:       environment.Production,
				Port:              "8080",
				InterruptTimeout:  2 * time.Second,
				ReadHeaderTimeout: 5 * time.Second,
			},
		},
		{
			name: "custom port",
			envVars: map[string]string{
				"SERVER_PORT": "3000",
			},
			expected: internal.Server{
				Environment:       environment.Testing,
				Port:              "3000",
				InterruptTimeout:  2 * time.Second,
				ReadHeaderTimeout: 5 * time.Second,
			},
		},
		{
			name: "custom timeouts",
			envVars: map[string]string{
				"INTERRUPT_TIMEOUT":   "30s",
				"READ_HEADER_TIMEOUT": "60s",
			},
			expected: internal.Server{
				Environment:       environment.Testing,
				Port:              "8080",
				InterruptTimeout:  30 * time.Second,
				ReadHeaderTimeout: 60 * time.Second,
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
		"MONGO_URI",
		"MONGO_DATABASE",
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
		"MONGO_URI",
		"MONGO_DATABASE",
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
