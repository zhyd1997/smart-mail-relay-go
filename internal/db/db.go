package db

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/sirupsen/logrus"

	"smart-mail-relay-go/internal/config"
	"smart-mail-relay-go/internal/models"
)

// Init initializes the database connection and runs migrations
func Init(cfg config.DatabaseConfig) (*gorm.DB, error) {
	gormLogger := logger.New(
		logrus.StandardLogger(),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(mysql.Open(cfg.GetDSN()), &gorm.Config{Logger: gormLogger})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := runMigrations(db); err != nil {
		return nil, err
	}

	logrus.Info("Database initialized successfully")
	return db, nil
}

func runMigrations(db *gorm.DB) error {
	logrus.Info("Running database migrations...")
	if err := db.AutoMigrate(&models.ForwardRule{}, &models.ProcessedEmail{}, &models.ForwardLog{}); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}
	logrus.Info("Database migrations completed")
	return nil
}
