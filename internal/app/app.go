package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/config"
	httpHandlers "github.com/iselldonuts/ai-for-developers-project-386/internal/http/handlers"
	httpMiddleware "github.com/iselldonuts/ai-for-developers-project-386/internal/http/middleware"
	httpRouter "github.com/iselldonuts/ai-for-developers-project-386/internal/http/router"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/platform/logger"
	platformPostgres "github.com/iselldonuts/ai-for-developers-project-386/internal/platform/postgres"
	postgresMigrations "github.com/iselldonuts/ai-for-developers-project-386/internal/platform/postgres/migrations"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/platform/postgres/txmanager"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/service"
	memoryStore "github.com/iselldonuts/ai-for-developers-project-386/internal/store/memory"
	postgresStore "github.com/iselldonuts/ai-for-developers-project-386/internal/store/postgres"
	"github.com/rs/zerolog"
)

type App struct {
	config     config.Config
	logger     zerolog.Logger
	httpServer *http.Server
}

type eventTypeStore interface {
	service.EventTypeStore
	service.BookingEventTypeStore
}

func New(ctx context.Context) (*App, error) {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	ownerStore, availabilityStore, bookingStore, eventTypeStore, txManager, err := stores(ctx, cfg)
	if err != nil {
		return nil, err
	}

	availabilityService := service.NewAvailabilityService(ownerStore, availabilityStore, txManager)
	bookingService := service.NewBookingService(bookingStore, eventTypeStore, ownerStore, availabilityStore, txManager)
	eventTypeService := service.NewEventTypeService(eventTypeStore, txManager)

	router := httpRouter.New(
		httpHandlers.NewHealthHandler(),
		httpHandlers.NewAPIHandler(availabilityService, bookingService, eventTypeService),
		httpMiddleware.CORS(cfg.CORSAllowedOrigins),
		httpMiddleware.RequestLogger(log),
	)

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	return &App{
		config:     cfg,
		logger:     log,
		httpServer: server,
	}, nil
}

func stores(ctx context.Context, cfg config.Config) (
	service.OwnerStore,
	service.AvailabilityStore,
	service.BookingStore,
	eventTypeStore,
	service.TxManager,
	error,
) {
	switch cfg.Storage {
	case "memory":
		store := memoryStore.New()
		return store, store, store, store, service.NoopTxManager{}, nil
	case "postgres":
		if err := postgresMigrations.Up(ctx, cfg.DatabaseURL, "migrations"); err != nil {
			return nil, nil, nil, nil, nil, err
		}

		pool, err := platformPostgres.NewPool(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}

		txManager := txmanager.New(pool)
		return postgresStore.NewOwnerStore(txManager),
			postgresStore.NewAvailabilityStore(txManager),
			postgresStore.NewBookingStore(txManager),
			postgresStore.NewEventTypeStore(txManager),
			txManager,
			nil
	default:
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported storage: %s", cfg.Storage)
	}
}

func (a *App) Run(_ context.Context) error {
	a.logger.Info().
		Str("env", a.config.Env).
		Str("addr", a.config.Addr).
		Msg("starting HTTP server")

	return a.httpServer.ListenAndServe()
}
