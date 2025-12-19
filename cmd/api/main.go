package main

import (
	"context"
	"log"
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

	db.ApplyMigrations(ctx, pool)
	seeder := db.NewSeeder(pool)
	if err := seeder.SeedAll(ctx, "migrations/json/employment"); err != nil {
		log.Fatal("Seeding failed", err)
		os.Exit(100)
	}

	userRepo := db.NewUserRepository(pool)
	userService := services.NewUserService(userRepo)
	userHandler := api.NewUserHandler(userService)

	departmentRepo := db.NewDepartmentRepository(pool)
	departmentService := services.NewDepartmentService(departmentRepo)
	departmentHandler := api.NewDepartmentHandler(departmentService)

	salaryRepo := db.NewSalaryRepository(pool)
	salaryService := services.NewSalaryService(salaryRepo)
	salaryHandler := api.NewSalaryHandler(salaryService)

	addressRepo := db.NewAddressRepository(pool)
	addressService := services.NewAddressService(addressRepo)
	addressHandler := api.NewAddressHandler(addressService)

	skillRepo := db.NewSkillRepository(pool)
	skillService := services.NewSkillService(skillRepo)
	skillHandler := api.NewSkillHandler(skillService)

	positionRepo := db.NewPositionRepository(pool)
	positionService := services.NewPositionService(positionRepo)
	positionHandler := api.NewPositionHandler(positionService)

	cryptoRepo := db.NewCryptoRepository(pool)
	cryptoRepo.SeedCryptosIfEmpty(ctx, "migrations/json/crypto/generated_cryptos.json")
	cryptoService := services.NewCryptoService(cryptoRepo)

	forumRepo := db.NewForumUserRepository(pool)
	forumService := services.NewForumUserService(forumRepo)
	forumHandler := api.NewForumUserHandler(forumService)

	go cryptoService.StartPriceTicker(ctx)

	cryptoHandler := api.NewCryptoHandler(cryptoService)

	router := gin.New()

	router.Use(
		gin.Recovery(),
		// api.LoggingMiddleware(slog.Default()),
		api.CORSMiddleware(cfg.AllowedOrigins),
	)

	v1 := router.Group("/api")
	{
		api.RegisterUserRoutes(v1.Group("/user"), userHandler)
		api.RegisterCryptoRoutes(v1.Group("/crypto"), cryptoHandler)
		api.RegisterDepartmentRoutes(v1.Group("/department"), departmentHandler)
		api.RegisterSalaryRoutes(v1.Group("/salary"), salaryHandler)
		api.RegisterSkillRoutes(v1.Group("/skill"), skillHandler)
		api.RegisterPositionRoutes(v1.Group("/position"), positionHandler)
		api.RegisterAddressRoutes(v1.Group("/address"), addressHandler)
		api.RegisterForumUserRoutes(v1.Group("/forum"), forumHandler)
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

	slog.Warn("Dropping Tables for Employments")
	seeder.DeleteDevData(context.Background())
	cryptoRepo.DeleteDevData(context.Background())
	forumRepo.DeleteForumTables(context.Background())

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	} else {
		slog.Info("server stopped gracefully")
	}
}
