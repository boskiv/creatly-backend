package app

import (
	"context"
	"github.com/zhashkevych/courses-backend/pkg/email/smtp"
	"github.com/zhashkevych/courses-backend/pkg/otp"
	"github.com/zhashkevych/courses-backend/pkg/payment/fondy"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/zhashkevych/courses-backend/internal/config"
	"github.com/zhashkevych/courses-backend/internal/delivery/http"
	"github.com/zhashkevych/courses-backend/internal/repository"
	"github.com/zhashkevych/courses-backend/internal/server"
	"github.com/zhashkevych/courses-backend/internal/service"
	"github.com/zhashkevych/courses-backend/pkg/auth"
	"github.com/zhashkevych/courses-backend/pkg/cache"
	"github.com/zhashkevych/courses-backend/pkg/database/mongodb"
	"github.com/zhashkevych/courses-backend/pkg/email/sendpulse"
	"github.com/zhashkevych/courses-backend/pkg/hash"
	"github.com/zhashkevych/courses-backend/pkg/logger"
)

// @title Course Platform API
// @version 1.0
// @description API Server for Course Platform

// @host localhost:8000
// @BasePath /api/v1/

// @securityDefinitions.apikey AdminAuth
// @in header
// @name Authorization

// @securityDefinitions.apikey StudentsAuth
// @in header
// @name Authorization

// Run initializes whole application
func Run(configPath string) {
	cfg, err := config.Init(configPath)
	if err != nil {
		logger.Error(err)
		return
	}

	// Dependencies
	mongoClient, err := mongodb.NewClient(cfg.Mongo.URI, cfg.Mongo.User, cfg.Mongo.Password)
	if err != nil {
		logger.Error(err)
		return
	}

	db := mongoClient.Database(cfg.Mongo.Name)

	memCache := cache.NewMemoryCache()
	hasher := hash.NewSHA1Hasher(cfg.Auth.PasswordSalt)
	emailProvider := sendpulse.NewClient(cfg.Email.ClientID, cfg.Email.ClientSecret, memCache)
	emailSender := smtp.NewSMTPSender(cfg.SMTP.From, cfg.SMTP.Pass, cfg.SMTP.Host, cfg.SMTP.Port)
	paymentProvider := fondy.NewFondyClient(cfg.Payment.Fondy.MerchantId, cfg.Payment.Fondy.MerchantPassword)
	tokenManager, err := auth.NewManager(cfg.Auth.JWT.SigningKey)
	if err != nil {
		logger.Error(err)
		return
	}
	otpGenerator := otp.NewGOTPGenerator()

	// Services, Repos & API Handlers
	repos := repository.NewRepositories(db)
	services := service.NewServices(service.Deps{
		Repos:                  repos,
		Cache:                  memCache,
		Hasher:                 hasher,
		TokenManager:           tokenManager,
		EmailProvider:          emailProvider,
		EmailSender:            emailSender,
		EmailListId:            cfg.Email.ListID,
		PaymentProvider:        paymentProvider,
		AccessTokenTTL:         cfg.Auth.JWT.AccessTokenTTL,
		RefreshTokenTTL:        cfg.Auth.JWT.RefreshTokenTTL,
		PaymentResponseURL:     cfg.Payment.ResponseURL,
		PaymentCallbackURL:     cfg.Payment.CallbackURL,
		CacheTTL:               int64(cfg.CacheTTL.Seconds()),
		OtpGenerator:           otpGenerator,
		VerificationCodeLength: cfg.Auth.VerificationCodeLength,
		FrontendURL:            cfg.FrontendURL,
	})
	handlers := http.NewHandler(services, tokenManager)

	// HTTP Server
	srv := server.NewServer(cfg, handlers.Init(cfg.HTTP.Host, cfg.HTTP.Port, cfg.Limiter, cfg.Cors))
	go func() {
		if err := srv.Run(); err != nil {
			logrus.Errorf("error occurred while running http server: %s\n", err.Error())
		}
	}()

	logger.Info("Server started")

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	if err := mongoClient.Disconnect(context.Background()); err != nil {
		logger.Error(err.Error())
	}
}
