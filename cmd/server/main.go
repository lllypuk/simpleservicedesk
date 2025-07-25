package main

import (
	"context"
	"log/slog"
	"os"

	"simpleservicedesk/internal"
	"simpleservicedesk/pkg/logger"
)

func main() {
	logger.Setup()

	loggerInstance := slog.Default()

	cfg, err := internal.LoadConfig()
	if err != nil {
		loggerInstance.ErrorContext(context.Background(), "Could not load config", "err", err)
		os.Exit(1)
	}

	err = internal.Run(cfg)
	if err != nil {
		loggerInstance.ErrorContext(context.Background(), "Failed to run server", "err", err)
		os.Exit(1)
	}
}
