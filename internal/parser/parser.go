package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"smart-mail-relay-go/internal/models"
)

// EmailParser handles parsing and matching of email subjects
type EmailParser struct {
	db *gorm.DB
}

// NewEmailParser creates a new email parser
func NewEmailParser(db *gorm.DB) *EmailParser {
	return &EmailParser{
		db: db,
	}
}

// ParseAndMatchEmail parses an email and finds matching forwarding rules
func (p *EmailParser) ParseAndMatchEmail(email models.EmailMessage) (*models.ForwardRule, error) {
	// Extract keyword from subject
	keyword, err := p.extractKeyword(email.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to extract keyword from subject: %w", err)
	}

	if keyword == "" {
		logrus.Debugf("No keyword found in subject: %s", email.Subject)
		return nil, nil
	}

	// Find matching rule
	rule, err := p.findMatchingRule(keyword)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching rule: %w", err)
	}

	if rule == nil {
		logrus.Debugf("No matching rule found for keyword: %s", keyword)
		return nil, nil
	}

	logrus.Infof("Found matching rule for keyword '%s': %s -> %s", keyword, rule.Keyword, rule.TargetEmail)
	return rule, nil
}

// extractKeyword extracts the keyword from email subject
// Expected format: "<keyword> - <recipient_name>"
func (p *EmailParser) extractKeyword(subject string) (string, error) {
	if subject == "" {
		return "", nil
	}

	// Clean up the subject
	subject = strings.TrimSpace(subject)

	// Try to match the pattern "<keyword> - <recipient_name>"
	// This regex matches: word/words followed by " - " followed by any text
	pattern := `^([^-]+)\s*-\s*(.+)$`
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(subject)
	if len(matches) < 3 {
		// If the pattern doesn't match, try to extract just the first word as keyword
		words := strings.Fields(subject)
		if len(words) > 0 {
			return strings.TrimSpace(words[0]), nil
		}
		return "", nil
	}

	keyword := strings.TrimSpace(matches[1])
	return keyword, nil
}

// findMatchingRule finds a forwarding rule that matches the given keyword
func (p *EmailParser) findMatchingRule(keyword string) (*models.ForwardRule, error) {
	var rule models.ForwardRule

	// First try exact match
	result := p.db.Where("keyword = ? AND enabled = ?", keyword, true).First(&rule)
	if result.Error == nil {
		return &rule, nil
	}

	if result.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	// If no exact match, try case-insensitive match
	result = p.db.Where("LOWER(keyword) = LOWER(?) AND enabled = ?", keyword, true).First(&rule)
	if result.Error == nil {
		return &rule, nil
	}

	if result.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	// If still no match, try partial match (keyword contains the search term)
	result = p.db.Where("keyword LIKE ? AND enabled = ?", "%"+keyword+"%", true).First(&rule)
	if result.Error == nil {
		return &rule, nil
	}

	if result.Error != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	// No matching rule found
	return nil, nil
}

// GetAllRules returns all forwarding rules
func (p *EmailParser) GetAllRules() ([]models.ForwardRule, error) {
	var rules []models.ForwardRule
	result := p.db.Find(&rules)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get rules: %w", result.Error)
	}
	return rules, nil
}

// GetEnabledRules returns all enabled forwarding rules
func (p *EmailParser) GetEnabledRules() ([]models.ForwardRule, error) {
	var rules []models.ForwardRule
	result := p.db.Where("enabled = ?", true).Find(&rules)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get enabled rules: %w", result.Error)
	}
	return rules, nil
}

// IsEmailProcessed checks if an email has already been processed
func (p *EmailParser) IsEmailProcessed(messageID string) (bool, error) {
	var processed models.ProcessedEmail
	result := p.db.Where("message_id = ?", messageID).First(&processed)

	if result.Error == nil {
		return true, nil // Email has been processed
	}

	if result.Error == gorm.ErrRecordNotFound {
		return false, nil // Email has not been processed
	}

	return false, fmt.Errorf("database error checking processed email: %w", result.Error)
}

// MarkEmailAsProcessed marks an email as processed
func (p *EmailParser) MarkEmailAsProcessed(messageID string) error {
	processed := models.ProcessedEmail{
		MessageID:   messageID,
		ProcessedAt: time.Now(),
	}

	result := p.db.Create(&processed)
	if result.Error != nil {
		return fmt.Errorf("failed to mark email as processed: %w", result.Error)
	}

	return nil
}

// LogForwardAttempt logs a forwarding attempt
func (p *EmailParser) LogForwardAttempt(messageID string, ruleID *uint, status string, errorMsg string) error {
	log := models.ForwardLog{
		MessageID: messageID,
		RuleID:    ruleID,
		Status:    status,
		ErrorMsg:  errorMsg,
		CreatedAt: time.Now(),
	}

	result := p.db.Create(&log)
	if result.Error != nil {
		return fmt.Errorf("failed to log forward attempt: %w", result.Error)
	}

	return nil
}
