package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"simpleservicedesk/internal/application"
	userdomain "simpleservicedesk/internal/domain/users"
	categoriesInfra "simpleservicedesk/internal/infrastructure/categories"
	healthInfra "simpleservicedesk/internal/infrastructure/health"
	organizationsInfra "simpleservicedesk/internal/infrastructure/organizations"
	ticketsInfra "simpleservicedesk/internal/infrastructure/tickets"
	usersInfra "simpleservicedesk/internal/infrastructure/users"
	"simpleservicedesk/internal/queries"
	"simpleservicedesk/pkg/environment"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"golang.org/x/sync/errgroup"
)

const disconnectTimeout = 5 * time.Second
const insecureBootstrapAdminEmail = "admin@example.com"
const insecureBootstrapAdminPassword = "change-me"

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

	if err = startServer(ctx, g, cfg, db, mongoClient); err != nil {
		return err
	}

	if err = g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("server exited with error: %w", err)
	}
	return nil
}

func startServer(
	ctx context.Context,
	g *errgroup.Group,
	cfg Config,
	db *mongo.Database,
	mongoClient *mongo.Client,
) error {
	userRepo := usersInfra.NewMongoRepo(db)
	ticketRepo := ticketsInfra.NewMongoRepo(db)
	organizationRepo := organizationsInfra.NewMongoRepo(db)
	categoryRepo := categoriesInfra.NewMongoRepo(db)
	pinger := healthInfra.NewMongoPinger(mongoClient)
	if err := ensureBootstrapAdminUser(ctx, userRepo, cfg.Server.Environment, cfg.Auth); err != nil {
		return err
	}

	httpServer, err := application.SetupHTTPServer(
		userRepo,
		ticketRepo,
		organizationRepo,
		categoryRepo,
		pinger,
		cfg.Auth.JWTSigningKey,
		cfg.Auth.JWTExpiration,
		cfg.Server.CORSAllowedOrigins,
		cfg.Server.RateLimitRPS,
	)
	if err != nil {
		return fmt.Errorf("failed to set up http server: %w", err)
	}

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
		if listenErr := server.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			return listenErr
		}
		slog.InfoContext(ctx, "Http server shut down gracefully")
		return nil
	})
	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.InterruptTimeout)
		defer cancel()
		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			return shutdownErr
		}
		return nil
	})

	return nil
}

func ensureBootstrapAdminUser(
	ctx context.Context,
	userRepo *usersInfra.MongoRepo,
	envType environment.Type,
	authCfg Auth,
) error {
	logger := slog.Default()

	bootstrapName := strings.TrimSpace(authCfg.BootstrapAdminName)
	bootstrapEmail := strings.ToLower(strings.TrimSpace(authCfg.BootstrapAdminEmail))
	bootstrapPassword := strings.TrimSpace(authCfg.BootstrapAdminPassword)

	if bootstrapName == "" && bootstrapEmail == "" && bootstrapPassword == "" {
		return nil
	}
	if bootstrapEmail == "" || bootstrapPassword == "" {
		return errors.New("bootstrap admin requires both BOOTSTRAP_ADMIN_EMAIL and BOOTSTRAP_ADMIN_PASSWORD")
	}
	if envType != environment.Testing {
		if strings.EqualFold(bootstrapEmail, insecureBootstrapAdminEmail) &&
			bootstrapPassword == insecureBootstrapAdminPassword {
			return errors.New("bootstrap admin uses insecure default credentials")
		}
	}
	if bootstrapName == "" {
		bootstrapName = "Bootstrap Admin"
	}

	adminRole := string(userdomain.RoleAdmin)
	adminUsersCount, err := userRepo.CountUsers(ctx, queries.UserFilter{Role: &adminRole})
	if err != nil {
		return fmt.Errorf("failed to count admin users before bootstrap admin creation: %w", err)
	}
	if adminUsersCount > 0 {
		logger.InfoContext(ctx, "bootstrap admin skipped because admin users already exist")
		return nil
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(bootstrapPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash bootstrap admin password: %w", err)
	}

	_, err = userRepo.CreateUser(ctx, bootstrapEmail, passwordHash, func() (*userdomain.User, error) {
		now := time.Now().UTC()
		return userdomain.NewUserWithDetails(
			uuid.New(),
			bootstrapName,
			bootstrapEmail,
			passwordHash,
			userdomain.RoleAdmin,
			nil,
			true,
			now,
			now,
		)
	})
	if err != nil {
		if errors.Is(err, userdomain.ErrUserAlreadyExist) {
			logger.InfoContext(ctx, "bootstrap admin already exists")
			return nil
		}
		return fmt.Errorf("failed to create bootstrap admin user: %w", err)
	}

	logger.InfoContext(ctx, "bootstrap admin user created", "email", bootstrapEmail)
	return nil
}
