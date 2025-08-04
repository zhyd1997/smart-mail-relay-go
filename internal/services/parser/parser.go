package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"smart-mail-relay-go/internal/models"
	"smart-mail-relay-go/internal/repository"
)

// EmailParser handles parsing and matching of email subjects
type EmailParser struct {
	repo *repository.Repository
}

// NewEmailParser creates a new email parser
func NewEmailParser(repo *repository.Repository) *EmailParser {
	return &EmailParser{repo: repo}
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
	rule, err := p.repo.FindMatchingRule(keyword)
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

// GetAllRules returns all forwarding rules
func (p *EmailParser) GetAllRules() ([]models.ForwardRule, error) {
	return p.repo.GetAllRules()
}

// GetEnabledRules returns all enabled forwarding rules
func (p *EmailParser) GetEnabledRules() ([]models.ForwardRule, error) {
	return p.repo.GetEnabledRules()
}

// IsEmailProcessed checks if an email has already been processed
func (p *EmailParser) IsEmailProcessed(messageID string) (bool, error) {
	return p.repo.IsEmailProcessed(messageID)
}

// MarkEmailAsProcessed marks an email as processed
func (p *EmailParser) MarkEmailAsProcessed(messageID string) error {
	return p.repo.MarkEmailAsProcessed(messageID)
}

// LogForwardAttempt logs a forwarding attempt
func (p *EmailParser) LogForwardAttempt(messageID string, ruleID *uint, status string, errorMsg string) error {
	return p.repo.LogForwardAttempt(messageID, ruleID, status, errorMsg)
}
