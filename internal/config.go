package internal

import (
	"fmt"
	"os"
	"time"

	"simpleservicedesk/pkg/environment"
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

	config.Server, err = LoadServer()
	if err != nil {
		return config, fmt.Errorf("could not load server config: %w", err)
	}

	config.Mongo = LoadMongo()
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

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
