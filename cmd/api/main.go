package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	cfgPkg "smart-mail-relay-go/config"
	"smart-mail-relay-go/internal/database"
	handlerPkg "smart-mail-relay-go/internal/handler"
	metricsPkg "smart-mail-relay-go/internal/metrics"
	"smart-mail-relay-go/internal/router"
	"smart-mail-relay-go/internal/service"
)

func main() {
	// Configure logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	logrus.Info("Starting Smart Mail Relay Service")

	// Load configuration
	cfg, err := cfgPkg.LoadConfig()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logrus.Fatalf("Configuration validation failed: %v", err)
	}

	// Initialize database
	db, err := database.InitDatabase(cfg.Database)
	if err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize metrics
	metrics := metricsPkg.NewMetrics()

	// Initialize email fetcher
	var fetcher service.EmailFetcher
	if cfg.Gmail.UseIMAP {
		fetcher, err = service.NewIMAPFetcher(&cfg.Gmail)
		if err != nil {
			logrus.Fatalf("Failed to create IMAP fetcher: %v", err)
		}
		logrus.Info("Using IMAP for email fetching")
	} else {
		fetcher, err = service.NewGmailAPIFetcher(&cfg.Gmail)
		if err != nil {
			logrus.Fatalf("Failed to create Gmail API fetcher: %v", err)
		}
		logrus.Info("Using Gmail API for email fetching")
	}

	// Initialize email parser
	parser := service.NewEmailParser(db)

	// Initialize email forwarder
	forwarder, err := service.NewEmailForwarder(&cfg.Gmail)
	if err != nil {
		logrus.Fatalf("Failed to create email forwarder: %v", err)
	}

	// Initialize scheduler
	scheduler := service.NewScheduler(&cfg.Scheduler, fetcher, parser, forwarder, metrics)

	// Initialize HTTP handlers
	handlers := handlerPkg.NewHandlers(db, parser, scheduler, metrics)

	// Setup HTTP server
	r := router.SetupRouter(handlers)
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start scheduler
	if err := scheduler.Start(); err != nil {
		logrus.Fatalf("Failed to start scheduler: %v", err)
	}

	// Start HTTP server in a goroutine
	go func() {
		logrus.Infof("Starting HTTP server on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop scheduler
	if err := scheduler.Stop(); err != nil {
		logrus.Errorf("Failed to stop scheduler: %v", err)
	}

	// Wait for scheduler to finish
	scheduler.Wait()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logrus.Errorf("HTTP server shutdown error: %v", err)
	}

	// Close fetcher
	if err := fetcher.Close(); err != nil {
		logrus.Errorf("Failed to close fetcher: %v", err)
	}

	// Close forwarder
	if err := forwarder.Close(); err != nil {
		logrus.Errorf("Failed to close forwarder: %v", err)
	}

	logrus.Info("Server stopped gracefully")
}
