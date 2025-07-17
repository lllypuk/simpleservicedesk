package main

import (
	"log/slog"
	"os"

	"simpleservicedesk/internal"
)

func main() {
	logger := slog.Default()

	cfg, err := internal.LoadConfig()
	if err != nil {
		logger.Error("Could not load config", "err", err)
		os.Exit(1)
	}

	err = internal.Run(cfg)
	if err != nil {
		logger.Error("Failed to run server", "err", err)
		os.Exit(1)
	}
}
