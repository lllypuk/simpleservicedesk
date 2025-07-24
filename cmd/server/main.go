package main

import (
	"context"
	"log/slog"
	"os"

	"simpleservicedesk/internal"
)

func main() {
	logger := slog.Default()

	cfg, err := internal.LoadConfig()
	if err != nil {
		logger.ErrorContext(context.Background(), "Could not load config", "err", err)
		os.Exit(1)
	}

	err = internal.Run(cfg)
	if err != nil {
		logger.ErrorContext(context.Background(), "Failed to run server", "err", err)
		os.Exit(1)
	}
}
