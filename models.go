package main

import (
	"time"

	"gorm.io/gorm"
)

// ForwardRule represents a forwarding rule in the database
type ForwardRule struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Keyword     string         `json:"keyword" gorm:"type:varchar(255);not null;uniqueIndex"`
	TargetEmail string         `json:"target_email" gorm:"type:varchar(255);not null"`
	Enabled     bool           `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName specifies the table name for ForwardRule
func (ForwardRule) TableName() string {
	return "forward_rules"
}

// ProcessedEmail represents a processed email to ensure idempotency
type ProcessedEmail struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	MessageID   string         `json:"message_id" gorm:"type:varchar(255);not null;uniqueIndex"`
	ProcessedAt time.Time      `json:"processed_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName specifies the table name for ProcessedEmail
func (ProcessedEmail) TableName() string {
	return "processed_emails"
}

// ForwardLog represents a log entry for email forwarding attempts
type ForwardLog struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	MessageID string         `json:"message_id" gorm:"type:varchar(255);not null;index"`
	RuleID    *uint          `json:"rule_id" gorm:"index"`
	Status    string         `json:"status" gorm:"type:varchar(50);not null"` // success, failure, skipped
	ErrorMsg  string         `json:"error_msg" gorm:"type:text"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationship
	Rule *ForwardRule `json:"rule,omitempty" gorm:"foreignKey:RuleID"`
}

// TableName specifies the table name for ForwardLog
func (ForwardLog) TableName() string {
	return "forward_logs"
}

// EmailMessage represents an email message structure
type EmailMessage struct {
	ID          string            `json:"id"`
	Subject     string            `json:"subject"`
	From        string            `json:"from"`
	To          []string          `json:"to"`
	CC          []string          `json:"cc"`
	BCC         []string          `json:"bcc"`
	Body        string            `json:"body"`
	HTMLBody    string            `json:"html_body"`
	Headers     map[string]string `json:"headers"`
	Raw         []byte            `json:"raw"`
	Attachments []Attachment      `json:"attachments"`
}

// Attachment represents an email attachment
type Attachment struct {
	Filename string `json:"filename"`
	MIMEType string `json:"mime_type"`
	Data     []byte `json:"data"`
}

// ForwardRuleRequest represents the request structure for creating/updating forward rules
type ForwardRuleRequest struct {
	Keyword     string `json:"keyword" binding:"required"`
	TargetEmail string `json:"target_email" binding:"required,email"`
	Enabled     *bool  `json:"enabled"`
}

// ForwardRuleResponse represents the response structure for forward rules
type ForwardRuleResponse struct {
	ID          uint      `json:"id"`
	Keyword     string    `json:"keyword"`
	TargetEmail string    `json:"target_email"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ForwardLogResponse represents the response structure for forward logs
type ForwardLogResponse struct {
	ID        uint                 `json:"id"`
	MessageID string               `json:"message_id"`
	RuleID    *uint                `json:"rule_id"`
	Status    string               `json:"status"`
	ErrorMsg  string               `json:"error_msg"`
	CreatedAt time.Time            `json:"created_at"`
	Rule      *ForwardRuleResponse `json:"rule,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Database  string            `json:"database"`
	Gmail     string            `json:"gmail"`
	Metrics   map[string]string `json:"metrics,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
