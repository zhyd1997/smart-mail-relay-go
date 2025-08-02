package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"smart-mail-relay-go/internal/config"
	"smart-mail-relay-go/internal/db"
	"smart-mail-relay-go/internal/fetcher"
	"smart-mail-relay-go/internal/forwarder"
	"smart-mail-relay-go/internal/handlers"
	"smart-mail-relay-go/internal/metrics"
	"smart-mail-relay-go/internal/parser"
	"smart-mail-relay-go/internal/scheduler"
	"smart-mail-relay-go/internal/server"
)

// Run initializes and starts the application
func Run() error {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	logrus.Info("Starting Smart Mail Relay Service")

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	dbConn, err := db.Init(cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	m := metrics.NewMetrics()

	var f fetcher.EmailFetcher
	if cfg.Gmail.UseIMAP {
		f, err = fetcher.NewIMAPFetcher(&cfg.Gmail)
		if err != nil {
			return fmt.Errorf("failed to create IMAP fetcher: %w", err)
		}
		logrus.Info("Using IMAP for email fetching")
	} else {
		f, err = fetcher.NewGmailAPIFetcher(&cfg.Gmail)
		if err != nil {
			return fmt.Errorf("failed to create Gmail API fetcher: %w", err)
		}
		logrus.Info("Using Gmail API for email fetching")
	}

	p := parser.NewEmailParser(dbConn)

	fw, err := forwarder.NewEmailForwarder(&cfg.Gmail)
	if err != nil {
		return fmt.Errorf("failed to create email forwarder: %w", err)
	}

	sched := scheduler.NewScheduler(&cfg.Scheduler, f, p, fw, m)

	h := handlers.NewHandlers(dbConn, p, sched, m)
	router := server.SetupRouter(h)
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	if err := sched.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	go func() {
		logrus.Infof("Starting HTTP server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := sched.Stop(); err != nil {
		logrus.Errorf("Failed to stop scheduler: %v", err)
	}
	sched.Wait()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("HTTP server shutdown error: %v", err)
	}

	if err := f.Close(); err != nil {
		logrus.Errorf("Failed to close fetcher: %v", err)
	}
	if err := fw.Close(); err != nil {
		logrus.Errorf("Failed to close forwarder: %v", err)
	}

	logrus.Info("Server stopped gracefully")
	return nil
}
