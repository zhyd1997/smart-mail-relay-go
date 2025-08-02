package main

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Gmail     GmailConfig     `mapstructure:"gmail"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// GmailConfig holds Gmail API configuration
type GmailConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RefreshToken string `mapstructure:"refresh_token"`
	UserEmail    string `mapstructure:"user_email"`
	UseIMAP      bool   `mapstructure:"use_imap"`
	IMAPHost     string `mapstructure:"imap_host"`
	IMAPPort     int    `mapstructure:"imap_port"`
	IMAPUser     string `mapstructure:"imap_user"`
	IMAPPassword string `mapstructure:"imap_password"`
}

// SchedulerConfig holds scheduler configuration
type SchedulerConfig struct {
	IntervalMinutes int `mapstructure:"interval_minutes"`
	MaxRetries      int `mapstructure:"max_retries"`
}

// LoadConfig loads configuration from environment variables and config file
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	setDefaults()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Environment variables override config file
	viper.AutomaticEnv()

	// Bind environment variables
	bindEnvVars()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.sslmode", "disable")

	viper.SetDefault("gmail.use_imap", false)
	viper.SetDefault("gmail.imap_host", "imap.gmail.com")
	viper.SetDefault("gmail.imap_port", 993)

	viper.SetDefault("scheduler.interval_minutes", 5)
	viper.SetDefault("scheduler.max_retries", 3)
}

// bindEnvVars binds environment variables to configuration keys
func bindEnvVars() {
	// Server
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")

	// Database
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.dbname", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")

	// Gmail
	viper.BindEnv("gmail.client_id", "GMAIL_CLIENT_ID")
	viper.BindEnv("gmail.client_secret", "GMAIL_CLIENT_SECRET")
	viper.BindEnv("gmail.refresh_token", "GMAIL_REFRESH_TOKEN")
	viper.BindEnv("gmail.user_email", "GMAIL_USER_EMAIL")
	viper.BindEnv("gmail.use_imap", "GMAIL_USE_IMAP")
	viper.BindEnv("gmail.imap_host", "GMAIL_IMAP_HOST")
	viper.BindEnv("gmail.imap_port", "GMAIL_IMAP_PORT")
	viper.BindEnv("gmail.imap_user", "GMAIL_IMAP_USER")
	viper.BindEnv("gmail.imap_password", "GMAIL_IMAP_PASSWORD")

	// Scheduler
	viper.BindEnv("scheduler.interval_minutes", "SCHEDULER_INTERVAL_MINUTES")
	viper.BindEnv("scheduler.max_retries", "SCHEDULER_MAX_RETRIES")
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if c.Database.Host == "" || c.Database.User == "" || c.Database.DBName == "" {
		return fmt.Errorf("database host, user, and dbname are required")
	}

	if !c.Gmail.UseIMAP {
		if c.Gmail.ClientID == "" || c.Gmail.ClientSecret == "" || c.Gmail.RefreshToken == "" {
			return fmt.Errorf("Gmail OAuth2 credentials are required when not using IMAP")
		}
	} else {
		if c.Gmail.IMAPUser == "" || c.Gmail.IMAPPassword == "" {
			return fmt.Errorf("IMAP credentials are required when using IMAP")
		}
	}

	if c.Scheduler.IntervalMinutes <= 0 {
		return fmt.Errorf("scheduler interval must be greater than 0")
	}

	return nil
}
