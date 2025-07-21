package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"simpleservicedesk/internal/application"
	usersInfra "simpleservicedesk/internal/infrastructure/users"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"golang.org/x/sync/errgroup"
)

func Run(cfg Config) error {
	g, ctx := errgroup.WithContext(context.Background())
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		return fmt.Errorf("failed to connect to mongo: %w", err)
	}
	g.Go(func() error {
		<-ctx.Done()
		slog.Info("shutting down mongo client")
		disconnectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(disconnectCtx); err != nil {
			slog.Error("failed to disconnect mongo client", "error", err)
		}
		return nil
	})

	db := mongoClient.Database(cfg.Mongo.Database)

	startServer(ctx, g, cfg, db)

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("server exited with error: %w", err)
	}
	return nil
}

func startServer(ctx context.Context, g *errgroup.Group, cfg Config, db *mongo.Database) {
	userRepo := usersInfra.NewMongoRepo(db)

	httpServer := application.SetupHTTPServer(userRepo)

	address := "0.0.0.0:" + cfg.Server.Port
	server := &http.Server{
		Addr: address,
		Handler: h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpServer.ServeHTTP(w, r)
		}), &http2.Server{}),
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
	}

	g.Go(func() error {
		slog.Info("Starting server http server at " + address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		slog.Info("Http server shut down gracefully")
		return nil
	})
	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.InterruptTimeout)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			return err
		}
		return nil
	})
}
