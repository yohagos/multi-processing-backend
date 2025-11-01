package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"multi-processing-backend/internal/api"
	"multi-processing-backend/internal/configs"
	"multi-processing-backend/internal/db"
	"multi-processing-backend/internal/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

func main() {
	cfg := configs.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool := db.ConnectDatabase(ctx, cfg.DatabaseURL)
	defer pool.Close()

	userRepo := db.NewUserRepository(pool)
	userService := services.NewUserService(userRepo)
	userHandler := api.NewUserHandler(userService)

	router := gin.New()

	router.Use(
		gin.Recovery(),
		api.LoggingMiddleware(slog.Default()),
		api.CORSMiddleware(cfg.AllowedOrigins),
	)

	v1 := router.Group("/api")
	{
		api.RegisterUserRoutes(v1.Group("/user"), userHandler)
	}

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		slog.Info("starting server", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	} else {
		slog.Info("server stopped gracefully")
	}
}
