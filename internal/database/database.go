package database

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"smart-mail-relay-go/config"
	"smart-mail-relay-go/internal/model"
)

// InitDatabase initializes the database connection and runs migrations
func InitDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
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
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logrus.Info("Database initialized successfully")
	return db, nil
}

func runMigrations(db *gorm.DB) error {
	logrus.Info("Running database migrations...")

	if err := db.AutoMigrate(&model.ForwardRule{}, &model.ProcessedEmail{}, &model.ForwardLog{}); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	logrus.Info("Database migrations completed")
	return nil
}
