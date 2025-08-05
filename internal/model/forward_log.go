package model

import (
	"time"

	"gorm.io/gorm"
)

// ForwardLog represents a log entry for email forwarding attempts
type ForwardLog struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	MessageID string         `json:"message_id" gorm:"type:varchar(255);not null;index"`
	RuleID    *uint          `json:"rule_id" gorm:"index"`
	Status    string         `json:"status" gorm:"type:varchar(50);not null"`
	ErrorMsg  string         `json:"error_msg" gorm:"type:text"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Rule *ForwardRule `json:"rule,omitempty" gorm:"foreignKey:RuleID"`
}

// TableName specifies the table name for ForwardLog
func (ForwardLog) TableName() string {
	return "forward_logs"
}
