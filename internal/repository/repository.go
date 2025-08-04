package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"smart-mail-relay-go/internal/models"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindMatchingRule(keyword string) (*models.ForwardRule, error) {
	var rule models.ForwardRule
	result := r.db.Where("keyword = ? AND enabled = ?", keyword, true).First(&rule)
	if result.Error == nil {
		return &rule, nil
	}
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	result = r.db.Where("LOWER(keyword) = LOWER(?) AND enabled = ?", keyword, true).First(&rule)
	if result.Error == nil {
		return &rule, nil
	}
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	result = r.db.Where("keyword LIKE ? AND enabled = ?", "%"+keyword+"%", true).First(&rule)
	if result.Error == nil {
		return &rule, nil
	}
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	return nil, nil
}

func (r *Repository) GetAllRules() ([]models.ForwardRule, error) {
	var rules []models.ForwardRule
	result := r.db.Find(&rules)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get rules: %w", result.Error)
	}
	return rules, nil
}

func (r *Repository) GetEnabledRules() ([]models.ForwardRule, error) {
	var rules []models.ForwardRule
	result := r.db.Where("enabled = ?", true).Find(&rules)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get enabled rules: %w", result.Error)
	}
	return rules, nil
}

func (r *Repository) IsEmailProcessed(messageID string) (bool, error) {
	var processed models.ProcessedEmail
	result := r.db.Where("message_id = ?", messageID).First(&processed)
	if result.Error == nil {
		return true, nil
	}
	if result.Error == gorm.ErrRecordNotFound {
		return false, nil
	}
	return false, fmt.Errorf("database error checking processed email: %w", result.Error)
}

func (r *Repository) MarkEmailAsProcessed(messageID string) error {
	processed := models.ProcessedEmail{
		MessageID:   messageID,
		ProcessedAt: time.Now(),
	}
	result := r.db.Create(&processed)
	if result.Error != nil {
		return fmt.Errorf("failed to mark email as processed: %w", result.Error)
	}
	return nil
}

func (r *Repository) LogForwardAttempt(messageID string, ruleID *uint, status, errorMsg string) error {
	log := models.ForwardLog{
		MessageID: messageID,
		RuleID:    ruleID,
		Status:    status,
		ErrorMsg:  errorMsg,
		CreatedAt: time.Now(),
	}
	result := r.db.Create(&log)
	if result.Error != nil {
		return fmt.Errorf("failed to log forward attempt: %w", result.Error)
	}
	return nil
}
