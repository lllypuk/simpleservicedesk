package internal

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
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
	config.Auth, err = LoadAuth()
	if err != nil {
		return config, fmt.Errorf("could not load auth config: %w", err)
	}

	return config, nil
}

type Server struct {
	Environment       environment.Type
	Port              string
	InterruptTimeout  time.Duration
	ReadHeaderTimeout time.Duration
	PprofPort         string
}

func LoadServer() (Server, error) {
	var server Server

	server.Environment = environment.Type(GetEnv("ENV_TYPE", string(environment.Testing)))
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

	return server, nil
}

func LoadMongo() Mongo {
	var mongo Mongo
	mongo.URI = GetEnv("MONGO_URI", "mongodb://localhost:27017")
	mongo.Database = GetEnv("MONGO_DATABASE", "servicedesk")
	return mongo
}

func LoadAuth() (Auth, error) {
	var auth Auth

	secret := strings.TrimSpace(GetEnv("JWT_SECRET", ""))
	if secret == "" {
		generatedSecret, err := generateDefaultJWTSecret()
		if err != nil {
			return auth, fmt.Errorf("could not generate default jwt secret: %w", err)
		}
		secret = generatedSecret
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

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
