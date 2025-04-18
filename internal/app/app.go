package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Olegsandrik/Exponenta/config"
	"github.com/Olegsandrik/Exponenta/internal/adapters/elasticsearch"
	"github.com/Olegsandrik/Exponenta/internal/adapters/minio"
	"github.com/Olegsandrik/Exponenta/internal/adapters/postgres"
	"github.com/Olegsandrik/Exponenta/internal/delivery"
	"github.com/Olegsandrik/Exponenta/internal/middleware"
	"github.com/Olegsandrik/Exponenta/internal/repository"
	"github.com/Olegsandrik/Exponenta/internal/usecase"

	"github.com/gorilla/mux"
)

const (
	_timeout = 5 * time.Second
)

type App struct {
	router  *mux.Router
	server  *http.Server
	logger  *slog.Logger
	closers []io.Closer
}

func (app *App) StartServer() error {
	return app.server.ListenAndServe()
}

func (app *App) StopServer(ctx context.Context) error {
	return app.server.Shutdown(ctx)
}

func InitServer(router *mux.Router, config *config.Config) *http.Server {
	return &http.Server{
		Addr:         config.Port,
		Handler:      router,
		WriteTimeout: config.ServerTimeout,
		ReadTimeout:  config.ServerTimeout,
	}
}

func InitApp() *App {
	cfg := config.NewConfig()

	// Router

	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.PanicMiddleware)
	r.Use(middleware.CorsMiddleware)

	apiRouter := r.PathPrefix("/api").Subrouter()

	// Logger

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	logger := slog.Default()

	// Postgres

	postgresAdapter, err := postgres.NewPostgresAdapter(cfg)
	if err != nil {
		panic(err)
	}

	// Server

	server := InitServer(r, cfg)

	// Minio

	minioAdapter, err := minio.NewMinioAdapter(cfg)
	if err != nil {
		panic(err)
	}

	// ElasticSearch

	elasticsearchAdapter, err := elasticsearch.NewElasticsearchAdapter(cfg)

	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), _timeout)
	defer cancel()

	err = elasticsearch.InitElasticSearchData(ctx, elasticsearchAdapter, postgresAdapter)

	if err != nil {
		panic(err)
	}

	// Images

	imageRepo := repository.NewImageRepository(minioAdapter)
	imageUsecase := usecase.NewImageUsecase(imageRepo)
	imageHandler := delivery.NewImageHandler(imageUsecase)
	imageHandler.InitRouter(apiRouter)

	// Cooking recipe

	cookingRecipeRepo := repository.NewCookingRecipeRepo(postgresAdapter)
	cookingRecipeUsecase := usecase.NewCookingRecipeUsecase(cookingRecipeRepo)
	cookingRecipeHandler := delivery.NewCookingRecipeHandler(cookingRecipeUsecase)
	cookingRecipeHandler.InitRouter(apiRouter)

	// Generation recipe

	generationRecipeRepo := repository.NewGeneratedRecipeRepo(postgresAdapter, cfg)
	generationRecipeUsecase := usecase.NewGenerateUsecase(generationRecipeRepo, cookingRecipeRepo)
	generationRecipeHandler := delivery.NewGeneratedHandler(generationRecipeUsecase)
	generationRecipeHandler.InitRouter(apiRouter)

	// Search

	searchRepo := repository.NewSearchRepository(elasticsearchAdapter, postgresAdapter)
	searchUsecase := usecase.NewSearchUsecase(searchRepo)
	searchHandler := delivery.NewSearchHandler(searchUsecase)
	searchHandler.InitRouter(apiRouter)

	// Voice

	voiceHandler := delivery.NewVoiceHandler(cfg)
	voiceHandler.InitRouter(apiRouter)

	closers := []io.Closer{postgresAdapter}
	return &App{
		router:  r,
		server:  server,
		closers: closers,
		logger:  logger,
	}
}

func (app *App) Start() {
	go func() {
		app.logger.Info("Server is running...")
		if err := app.StartServer(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.logger.Error("HTTP server ListenAndServe error: " + err.Error())
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-exit

	ctx, cancel := context.WithTimeout(context.Background(), _timeout)
	defer cancel()

	app.logger.Info("Shutting down server...")
	if err := app.StopServer(ctx); err != nil {
		app.logger.Error("HTTP server shutdown error: " + err.Error())
	}

	for _, closer := range app.closers {
		if err := closer.Close(); err != nil {
			app.logger.Error("Error closing resource: " + err.Error())
		}
	}
	app.logger.Info("All resources closed.")
}
