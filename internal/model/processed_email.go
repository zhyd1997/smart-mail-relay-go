package model

import (
	"time"

	"gorm.io/gorm"
)

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
