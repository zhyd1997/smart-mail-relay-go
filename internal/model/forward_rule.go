package model

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
