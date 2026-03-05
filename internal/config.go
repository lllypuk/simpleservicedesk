package internal

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"simpleservicedesk/pkg/environment"
)

type Config struct {
	Server Server
	Mongo  Mongo
	Auth   Auth
}

type Mongo struct {
	URI      string
	Database string
}

type Auth struct {
	JWTSigningKey          string
	JWTExpiration          time.Duration
	BootstrapAdminName     string
	BootstrapAdminEmail    string
	BootstrapAdminPassword string
}

const generatedJWTSecretLength = 32
const minProductionJWTSecretLength = 32
const insecureDefaultJWTSecret = "change-me-in-production"

func LoadConfig() (Config, error) {
	var (
		config Config
		err    error
	)

	config.Server, err = LoadServer()
	if err != nil {
		return config, fmt.Errorf("could not load server config: %w", err)
	}

	config.Mongo = LoadMongo()
	config.Auth, err = LoadAuth(config.Server.Environment)
	if err != nil {
		return config, fmt.Errorf("could not load auth config: %w", err)
	}

	return config, nil
}

type Server struct {
	Environment        environment.Type
	Port               string
	InterruptTimeout   time.Duration
	ReadHeaderTimeout  time.Duration
	PprofPort          string
	CORSAllowedOrigins []string
	RateLimitRPS       int
}

func LoadServer() (Server, error) {
	var server Server

	serverEnvironment := strings.TrimSpace(GetEnv("ENV_TYPE", string(environment.Testing)))
	serverEnvironment = strings.ToLower(serverEnvironment)
	server.Environment = environment.Type(serverEnvironment)
	server.Port = GetEnv("SERVER_PORT", "8080")
	interruptTimeout, err := time.ParseDuration(GetEnv("INTERRUPT_TIMEOUT", "2s"))
	if err != nil {
		return server, fmt.Errorf("could not parse interrupt timeout: %w", err)
	}
	server.InterruptTimeout = interruptTimeout
	readHeaderTimeout, err := time.ParseDuration(GetEnv("READ_HEADER_TIMEOUT", "5s"))
	if err != nil {
		return server, fmt.Errorf("could not parse read header timeout: %w", err)
	}
	server.ReadHeaderTimeout = readHeaderTimeout
	server.CORSAllowedOrigins = loadCORSAllowedOrigins()
	rateLimitRPS, err := loadRateLimitRPS()
	if err != nil {
		return server, err
	}
	server.RateLimitRPS = rateLimitRPS

	return server, nil
}

func loadRateLimitRPS() (int, error) {
	rawRate := strings.TrimSpace(GetEnv("RATE_LIMIT_RPS", "100"))
	if rawRate == "" {
		rawRate = "100"
	}

	rateLimitRPS, err := strconv.Atoi(rawRate)
	if err != nil {
		return 0, fmt.Errorf("could not parse rate limit rps: %w", err)
	}
	if rateLimitRPS <= 0 {
		return 0, errors.New("rate limit rps must be greater than zero")
	}

	return rateLimitRPS, nil
}

func loadCORSAllowedOrigins() []string {
	rawOrigins := strings.TrimSpace(GetEnv("CORS_ALLOWED_ORIGINS", "*"))
	if rawOrigins == "" {
		return []string{"*"}
	}

	parts := strings.Split(rawOrigins, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		origins = append(origins, origin)
	}

	if len(origins) == 0 {
		return []string{"*"}
	}

	return origins
}

func LoadMongo() Mongo {
	var mongo Mongo
	mongo.URI = GetEnv("MONGO_URI", "mongodb://localhost:27017")
	mongo.Database = GetEnv("MONGO_DATABASE", "servicedesk")
	return mongo
}

func LoadAuth(envType environment.Type) (Auth, error) {
	var auth Auth

	secret := strings.TrimSpace(GetEnv("JWT_SECRET", ""))
	if secret == "" {
		if envType == environment.Production {
			return auth, errors.New("jwt secret is required in production environment")
		}

		generatedSecret, err := generateDefaultJWTSecret()
		if err != nil {
			return auth, fmt.Errorf("could not generate default jwt secret: %w", err)
		}
		secret = generatedSecret
	}
	if envType == environment.Production {
		if err := validateProductionJWTSecret(secret); err != nil {
			return auth, err
		}
	}

	expiration, err := time.ParseDuration(GetEnv("JWT_EXPIRATION", "24h"))
	if err != nil {
		return auth, fmt.Errorf("could not parse jwt expiration: %w", err)
	}
	if expiration <= 0 {
		return auth, errors.New("jwt expiration must be greater than zero")
	}

	auth.JWTSigningKey = secret
	auth.JWTExpiration = expiration
	auth.BootstrapAdminName = strings.TrimSpace(GetEnv("BOOTSTRAP_ADMIN_NAME", ""))
	auth.BootstrapAdminEmail = strings.TrimSpace(GetEnv("BOOTSTRAP_ADMIN_EMAIL", ""))
	auth.BootstrapAdminPassword = GetEnv("BOOTSTRAP_ADMIN_PASSWORD", "")

	return auth, nil
}

func generateDefaultJWTSecret() (string, error) {
	secret := make([]byte, generatedJWTSecretLength)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(secret), nil
}

func validateProductionJWTSecret(secret string) error {
	normalizedSecret := strings.TrimSpace(secret)
	if strings.EqualFold(normalizedSecret, insecureDefaultJWTSecret) {
		return errors.New("jwt secret uses an insecure default value in production environment")
	}
	if len(normalizedSecret) < minProductionJWTSecretLength {
		return fmt.Errorf(
			"jwt secret must be at least %d characters in production environment",
			minProductionJWTSecretLength,
		)
	}

	return nil
}

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
