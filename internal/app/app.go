package app

import (
	"context"
	"errors"
	"github.com/Olegsandrik/Exponenta/config"
	"github.com/Olegsandrik/Exponenta/internal/adapters/postgres"
	"github.com/Olegsandrik/Exponenta/internal/delivery"
	"github.com/Olegsandrik/Exponenta/internal/middleware"
	"github.com/Olegsandrik/Exponenta/internal/repository"
	"github.com/Olegsandrik/Exponenta/internal/usecase"
	"github.com/gorilla/mux"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

func InitServer(router *mux.Router) *http.Server {
	return &http.Server{
		Addr:         ":8080",
		Handler:      router,
		WriteTimeout: config.SERVER_TIMEOUT,
		ReadTimeout:  config.SERVER_TIMEOUT,
	}
}

func InitApp() *App {

	// Router

	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware)

	apiRouter := r.PathPrefix("/api").Subrouter()

	// Logger

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	logger := slog.Default()

	// Postgres

	postgresAdapter, err := postgres.NewPostgresAdapter()
	if err != nil {
		panic(err)
	}

	server := InitServer(r)

	// Cooking recipe

	cookingRecipeRepo := repository.NewCookingRecipeRepo(postgresAdapter)
	cookingRecipeUsecase := usecase.NewCookingRecipeUsecase(cookingRecipeRepo)
	cookingRecipeHandler := delivery.NewCookingRecipeHandler(cookingRecipeUsecase)
	cookingRecipeHandler.InitRouter(apiRouter)

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

	ctx, cancel := context.WithTimeout(context.Background(), config.SERVER_TIMEOUT)
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
