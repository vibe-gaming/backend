package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apiHttp "github.com/vibe-gaming/backend/internal/api/http"
	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/internal/db"
	"github.com/vibe-gaming/backend/internal/repository"
	"github.com/vibe-gaming/backend/internal/server"
	"github.com/vibe-gaming/backend/internal/service"
	"github.com/vibe-gaming/backend/pkg/auth"
	"github.com/vibe-gaming/backend/pkg/email/smtp"
	"github.com/vibe-gaming/backend/pkg/hash"
	logger "github.com/vibe-gaming/backend/pkg/logger"
	"github.com/vibe-gaming/backend/pkg/otp"
)

func main() {
	// Init cfg from environment variables
	cfg := config.MustLoad()

	// Dependencies
	appLogger := logger.SetupLogger(cfg.Env)

	appLogger.Info("starting backend api", "env", cfg.Env)
	appLogger.Debug("debug messages are enabled")

	// Init database
	dbMySQL, err := db.New(cfg.Database)
	if err != nil {
		appLogger.Error("mysql connect problem", "error", err)
		os.Exit(1)
	}
	defer func() {
		err = dbMySQL.Close()
		if err != nil {
			appLogger.Error("error when closing", "error", err)
		}
	}()
	appLogger.Info("mysql connection done")

	hasher := hash.NewSHA1Hasher(cfg.Auth.PasswordSalt)

	emailSender, err := smtp.NewSMTPSender(cfg.SMTP.From, cfg.SMTP.Pass, cfg.SMTP.Host, cfg.SMTP.Port)
	if err != nil {
		appLogger.Error("smtp sender creation failed", "error", err)
		return
	}

	tokenManager, err := auth.NewManager(cfg.Auth.JWT)
	if err != nil {
		appLogger.Error("auth manager creation err", "error", err)
		return
	}

	otpGenerator := otp.NewGOTPGenerator()

	// Services, Repos & API Handlers
	repos := repository.NewRepositories(dbMySQL)
	services := service.NewServices(service.Deps{
		Logger:       appLogger,
		Config:       cfg,
		Hasher:       hasher,
		TokenManager: tokenManager,
		OtpGenerator: otpGenerator,
		EmailSender:  emailSender,
		Repos:        repos,
	})
	handlers := apiHttp.NewHandlers(services, appLogger, tokenManager)

	// HTTP Server
	srv := server.NewServer(cfg, handlers.Init(cfg))
	go func() {
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("error occurred while running http server", "error", err)
		}
	}()
	appLogger.Info("server started")

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second

	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := srv.Stop(ctx); err != nil {
		appLogger.Error("failed to stop server", "error", err)
	}

	appLogger.Info("app stopped")
}
