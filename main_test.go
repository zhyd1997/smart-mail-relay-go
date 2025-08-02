package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidation(t *testing.T) {
	// Test valid configuration
	config := &Config{
		Server: ServerConfig{
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host:   "localhost",
			User:   "test",
			DBName: "test",
		},
		Gmail: GmailConfig{
			ClientID:     "test",
			ClientSecret: "test",
			RefreshToken: "test",
		},
		Scheduler: SchedulerConfig{
			IntervalMinutes: 5,
		},
	}

	err := config.Validate()
	assert.NoError(t, err)

	// Test invalid configuration
	invalidConfig := &Config{
		Server: ServerConfig{
			Port: "",
		},
	}

	err = invalidConfig.Validate()
	assert.Error(t, err)
}

func TestDatabaseDSN(t *testing.T) {
	config := DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
	}

	dsn := config.GetDSN()
	expected := "testuser:testpass@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	assert.Equal(t, expected, dsn)
}

func TestEmailParserExtractKeyword(t *testing.T) {
	parser := &EmailParser{}

	// Test valid subject format
	keyword, err := parser.extractKeyword("urgent - John Doe")
	assert.NoError(t, err)
	assert.Equal(t, "urgent", keyword)

	// Test subject without dash
	keyword, err = parser.extractKeyword("urgent message")
	assert.NoError(t, err)
	assert.Equal(t, "urgent", keyword)

	// Test empty subject
	keyword, err = parser.extractKeyword("")
	assert.NoError(t, err)
	assert.Equal(t, "", keyword)
}

func TestForwardRuleValidation(t *testing.T) {
	rule := ForwardRule{
		Keyword:     "test",
		TargetEmail: "test@example.com",
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.NotEmpty(t, rule.Keyword)
	assert.NotEmpty(t, rule.TargetEmail)
	assert.True(t, rule.Enabled)
}

func TestEmailMessageStructure(t *testing.T) {
	email := EmailMessage{
		ID:       "test-id",
		Subject:  "Test Subject",
		From:     "sender@example.com",
		To:       []string{"recipient@example.com"},
		Body:     "Test body",
		HTMLBody: "<p>Test body</p>",
		Headers:  make(map[string]string),
	}

	assert.Equal(t, "test-id", email.ID)
	assert.Equal(t, "Test Subject", email.Subject)
	assert.Equal(t, "sender@example.com", email.From)
	assert.Len(t, email.To, 1)
	assert.Equal(t, "recipient@example.com", email.To[0])
	assert.Equal(t, "Test body", email.Body)
	assert.Equal(t, "<p>Test body</p>", email.HTMLBody)
	assert.NotNil(t, email.Headers)
}
