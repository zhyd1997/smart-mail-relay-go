package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Configure logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	logrus.Info("Starting Smart Mail Relay Service")

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		logrus.Fatalf("Configuration validation failed: %v", err)
	}

	// Initialize database
	db, err := initDatabase(config.Database)
	if err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize metrics
	metrics := NewMetrics()

	// Initialize email fetcher
	var fetcher EmailFetcher
	if config.Gmail.UseIMAP {
		fetcher, err = NewIMAPFetcher(&config.Gmail)
		if err != nil {
			logrus.Fatalf("Failed to create IMAP fetcher: %v", err)
		}
		logrus.Info("Using IMAP for email fetching")
	} else {
		fetcher, err = NewGmailAPIFetcher(&config.Gmail)
		if err != nil {
			logrus.Fatalf("Failed to create Gmail API fetcher: %v", err)
		}
		logrus.Info("Using Gmail API for email fetching")
	}

	// Initialize email parser
	parser := NewEmailParser(db)

	// Initialize email forwarder
	forwarder, err := NewEmailForwarder(&config.Gmail)
	if err != nil {
		logrus.Fatalf("Failed to create email forwarder: %v", err)
	}

	// Initialize scheduler
	scheduler := NewScheduler(&config.Scheduler, fetcher, parser, forwarder, metrics)

	// Initialize HTTP handlers
	handlers := NewHandlers(db, parser, scheduler, metrics)

	// Setup HTTP server
	router := setupRouter(handlers)
	server := &http.Server{
		Addr:         ":" + config.Server.Port,
		Handler:      router,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
	}

	// Start scheduler
	if err := scheduler.Start(); err != nil {
		logrus.Fatalf("Failed to start scheduler: %v", err)
	}

	// Start HTTP server in a goroutine
	go func() {
		logrus.Infof("Starting HTTP server on port %s", config.Server.Port)
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

// initDatabase initializes the database connection and runs migrations
func initDatabase(config DatabaseConfig) (*gorm.DB, error) {
	// Configure GORM logger
	gormLogger := logger.New(
		logrus.StandardLogger(),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Connect to database
	db, err := gorm.Open(mysql.Open(config.GetDSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logrus.Info("Database initialized successfully")
	return db, nil
}

// runMigrations runs database migrations
func runMigrations(db *gorm.DB) error {
	logrus.Info("Running database migrations...")

	// Auto migrate all models
	if err := db.AutoMigrate(&ForwardRule{}, &ProcessedEmail{}, &ForwardLog{}); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	logrus.Info("Database migrations completed")
	return nil
}

// setupRouter sets up the HTTP router with middleware
func setupRouter(handlers *Handlers) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(loggerMiddleware())

	// Setup routes
	handlers.SetupRoutes(router)

	return router
}

// loggerMiddleware adds logging middleware
func loggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}
