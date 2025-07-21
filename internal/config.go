package internal

import (
	"fmt"
	"os"
	"simpleservicedesk/pkg/environment"
	"time"
)

type Config struct {
	Server Server
	Mongo  Mongo
}

type Mongo struct {
	URI      string
	Database string
}

func LoadConfig() (Config, error) {
	var (
		config Config
		err    error
	)

	config.Server, err = loadServer()
	if err != nil {
		return config, fmt.Errorf("could not load server config: %w", err)
	}

	config.Mongo, err = loadMongo()
	if err != nil {
		return config, fmt.Errorf("could not load mongo config: %w", err)
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

func loadServer() (Server, error) {
	var server Server

	server.Environment = environment.Type(getEnv("ENV_TYPE", string(environment.Testing)))
	server.Port = getEnv("SERVER_PORT", "8080")
	interruptTimeout, err := time.ParseDuration(getEnv("INTERRUPT_TIMEOUT", "2s"))
	if err != nil {
		return server, fmt.Errorf("could not parse interrupt timeout: %w", err)
	}
	server.InterruptTimeout = interruptTimeout
	readHeaderTimeout, err := time.ParseDuration(getEnv("READ_HEADER_TIMEOUT", "5s"))
	if err != nil {
		return server, fmt.Errorf("could not parse read header timeout: %w", err)
	}
	server.ReadHeaderTimeout = readHeaderTimeout

	return server, nil
}

func loadMongo() (Mongo, error) {
	var mongo Mongo
	mongo.URI = getEnv("MONGO_URI", "mongodb://localhost:27017")
	mongo.Database = getEnv("MONGO_DATABASE", "servicedesk")
	return mongo, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
