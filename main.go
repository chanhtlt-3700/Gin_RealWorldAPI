package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"conduit/config"
	"conduit/controllers"
	"conduit/pkg/token"
	"conduit/pkg/validation"
	"conduit/repositories"
	"conduit/routes"
	"conduit/services"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// ---------- Config & Database ----------
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := config.ConnectDatabase(cfg.Database)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	if err := config.MigrateDatabase(db); err != nil {
		return err
	}

	validation.RegisterJSONTagNames()

	// ---------- Dependency Injection (wiring thủ công, tường minh) ----------
	tokenManager := token.NewManager(cfg.JWT.Secret, cfg.JWT.ExpiresIn, cfg.JWT.Issuer)

	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo, tokenManager)
	userController := controllers.NewUserController(userService)

	router := routes.SetupRouter(cfg, tokenManager, routes.Handlers{
		User: userController,
	})

	// ---------- HTTP server + graceful shutdown ----------
	srv := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("server started", "addr", srv.Addr, "env", cfg.App.Env, "db", cfg.Database.Driver)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
	}

	slog.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
