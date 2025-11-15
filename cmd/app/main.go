package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	apiHttp "github.com/vibe-gaming/backend/internal/api/http"
	"github.com/vibe-gaming/backend/internal/cache"
	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/internal/db"
	"github.com/vibe-gaming/backend/internal/esia"
	"github.com/vibe-gaming/backend/internal/queue/asynqserver"
	"github.com/vibe-gaming/backend/internal/queue/client"
	"github.com/vibe-gaming/backend/internal/repository"
	"github.com/vibe-gaming/backend/internal/server"
	"github.com/vibe-gaming/backend/internal/service"
	"github.com/vibe-gaming/backend/internal/worker"
	"github.com/vibe-gaming/backend/pkg/auth"
	"github.com/vibe-gaming/backend/pkg/email/smtp"
	"github.com/vibe-gaming/backend/pkg/hash"
	logger "github.com/vibe-gaming/backend/pkg/logger"
	"github.com/vibe-gaming/backend/pkg/otp"
	"go.uber.org/zap"
)

func main() {
	// Init cfg from environment variables
	cfg := config.MustLoad()

	// Delete "" from cfg.SMTP.Pass if env is local
	if cfg.Env == "local" {
		cfg.SMTP.Pass = strings.ReplaceAll(cfg.SMTP.Pass, "\"", "")
	}

	// Dependencies
	logger.Init(cfg.LogLevel)

	logger.Info("logger initialized")
	logger.Info("starting backend api", zap.String("env", cfg.Env))
	logger.Debug("debug messages are enabled")

	// Init database
	dbMySQL, err := db.New(cfg.Database)
	if err != nil {
		logger.Error("mysql connect problem", zap.Error(err))
		os.Exit(1)
	}
	defer func() {
		err = dbMySQL.Close()
		if err != nil {
			logger.Error("error when closing", zap.Error(err))
		}
	}()
	logger.Info("mysql connection done")

	// init redis cache
	redis, err := cache.NewRedis(cfg.Cache)
	if err != nil {
		logger.Error("redis init problem", zap.Error(err))
		os.Exit(1)
	}

	defer func() {
		err = redis.Close()
		if err != nil {
			logger.Error("error when closing", zap.Error(err))
		}
	}()
	logger.Info("redis connection done")

	hasher := hash.NewSHA1Hasher(cfg.Auth.PasswordSalt)

	emailSender, err := smtp.NewSMTPSender(cfg.SMTP.From, cfg.SMTP.Pass, cfg.SMTP.Host, cfg.SMTP.Port)
	if err != nil {
		logger.Error("smtp sender creation failed", zap.Error(err))
		return
	}

	tokenManager, err := auth.NewManager(cfg.Auth.JWT)
	if err != nil {
		logger.Error("auth manager creation err", zap.Error(err))
		return
	}

	otpGenerator := otp.NewGOTPGenerator()

	esiaClient := esia.NewClient(cfg.ESIA)

	// Services, Repos & API Handlers
	repos := repository.NewRepositories(dbMySQL)
	services := service.NewServices(service.Deps{
		Config:       cfg,
		Hasher:       hasher,
		TokenManager: tokenManager,
		OtpGenerator: otpGenerator,
		Repos:        repos,
		EsiaClient:   esiaClient,
	})
	workers := worker.NewWorkers(worker.Deps{
		Redis:         redis,
		Services:      services,
		EmailProvider: emailSender,
		Config:        cfg,
	})
	handlers := apiHttp.NewHandlers(services, tokenManager, cfg, esiaClient)

	// HTTP Server
	srv := server.NewServer(cfg, handlers.Init(cfg))
	go func() {
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("error occurred while running http server", zap.Error(err))
		}
	}()
	logger.Info("server started")

	// asynq server
	asynqServer, mux := asynqserver.New(cfg.Cache, workers)

	taskClient := asynq.NewClient(asynqserver.RedisOptions(cfg.Cache))
	defer func() {
		if err := taskClient.Close(); err != nil {
			logger.Error("asynq task client close problem", zap.Error(err))
		}
	}()
	client.SetClient(taskClient)

	if err = asynqServer.Start(mux); err != nil {
		logger.Fatal("asynq: start worker server failed", zap.Error(err))
	}

	logger.Info("asynq server started")

	logger.Info("app started")

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second

	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := srv.Stop(ctx); err != nil {
		logger.Error("failed to stop server", zap.Error(err))
	}

	logger.Info("app stopped")
}
