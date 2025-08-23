package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"simpleservicedesk/internal/application"
	organizationsInfra "simpleservicedesk/internal/infrastructure/organizations"
	ticketsInfra "simpleservicedesk/internal/infrastructure/tickets"
	usersInfra "simpleservicedesk/internal/infrastructure/users"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"golang.org/x/sync/errgroup"
)

const disconnectTimeout = 5 * time.Second

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
		slog.InfoContext(ctx, "shutting down mongo client")
		disconnectCtx, cancel := context.WithTimeout(context.Background(), disconnectTimeout)
		defer cancel()
		if err = mongoClient.Disconnect(disconnectCtx); err != nil {
			slog.ErrorContext(disconnectCtx, "failed to disconnect mongo client", "error", err)
		}
		return nil
	})

	db := mongoClient.Database(cfg.Mongo.Database)

	startServer(ctx, g, cfg, db)

	if err = g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("server exited with error: %w", err)
	}
	return nil
}

func startServer(ctx context.Context, g *errgroup.Group, cfg Config, db *mongo.Database) {
	userRepo := usersInfra.NewMongoRepo(db)
	ticketRepo := ticketsInfra.NewMongoRepo(db)
	organizationRepo := organizationsInfra.NewMongoRepo(db)

	httpServer := application.SetupHTTPServer(userRepo, ticketRepo, organizationRepo)

	address := "0.0.0.0:" + cfg.Server.Port
	server := &http.Server{
		Addr: address,
		Handler: h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpServer.ServeHTTP(w, r)
		}), &http2.Server{}),
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
	}

	g.Go(func() error {
		slog.InfoContext(ctx, "Starting server http server at "+address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		slog.InfoContext(ctx, "Http server shut down gracefully")
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
