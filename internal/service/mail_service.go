package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"smart-mail-relay-go/config"
	"smart-mail-relay-go/internal/model"
)

// EmailMessage represents an email message structure
type EmailMessage struct {
	ID       string            `json:"id"`
	Subject  string            `json:"subject"`
	From     string            `json:"from"`
	To       []string          `json:"to"`
	CC       []string          `json:"cc"`
	BCC      []string          `json:"bcc"`
	Body     string            `json:"body"`
	HTMLBody string            `json:"html_body"`
	Headers  map[string]string `json:"headers"`
	Raw      []byte            `json:"raw"`
}

// EmailFetcher interface for fetching emails
type EmailFetcher interface {
	FetchNewEmails(ctx context.Context) ([]EmailMessage, error)
	Close() error
}

// GmailAPIFetcher implements EmailFetcher using Gmail API
type GmailAPIFetcher struct {
	service   *gmail.Service
	userEmail string
	lastCheck time.Time
}

// IMAPFetcher implements EmailFetcher using IMAP
type IMAPFetcher struct {
	client    *client.Client
	lastCheck time.Time
}

// NewGmailAPIFetcher creates a new Gmail API fetcher
func NewGmailAPIFetcher(cfg *config.GmailConfig) (*GmailAPIFetcher, error) {
	ctx := context.Background()

	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       []string{gmail.GmailReadonlyScope},
		Endpoint:     google.Endpoint,
	}

	// Create token source from refresh token
	token := &oauth2.Token{
		RefreshToken: cfg.RefreshToken,
	}

	tokenSource := oauth2Config.TokenSource(ctx, token)

	// Create Gmail service
	service, err := gmail.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	return &GmailAPIFetcher{
		service:   service,
		userEmail: cfg.UserEmail,
		lastCheck: time.Now().Add(-24 * time.Hour), // Start with emails from last 24 hours
	}, nil
}

// NewIMAPFetcher creates a new IMAP fetcher
func NewIMAPFetcher(cfg *config.GmailConfig) (*IMAPFetcher, error) {
	// Connect to IMAP server
	c, err := client.DialTLS(fmt.Sprintf("%s:%d", cfg.IMAPHost, cfg.IMAPPort), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %w", err)
	}

	// Login
	if err := c.Login(cfg.IMAPUser, cfg.IMAPPassword); err != nil {
		c.Logout()
		return nil, fmt.Errorf("failed to login to IMAP server: %w", err)
	}

	return &IMAPFetcher{
		client:    c,
		lastCheck: time.Now().Add(-24 * time.Hour), // Start with emails from last 24 hours
	}, nil
}

// FetchNewEmails fetches new emails using Gmail API
func (f *GmailAPIFetcher) FetchNewEmails(ctx context.Context) ([]EmailMessage, error) {
	query := fmt.Sprintf("after:%d", f.lastCheck.Unix())

	// Search for new messages
	call := f.service.Users.Messages.List(f.userEmail).Q(query)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	var emails []EmailMessage

	for _, msg := range response.Messages {
		// Get full message details
		message, err := f.service.Users.Messages.Get(f.userEmail, msg.Id).Format("full").Do()
		if err != nil {
			logrus.Warnf("Failed to get message %s: %v", msg.Id, err)
			continue
		}

		email, err := f.parseGmailMessage(message)
		if err != nil {
			logrus.Warnf("Failed to parse message %s: %v", msg.Id, err)
			continue
		}

		emails = append(emails, email)
	}

	f.lastCheck = time.Now()
	return emails, nil
}

// parseGmailMessage parses a Gmail API message into EmailMessage
func (f *GmailAPIFetcher) parseGmailMessage(msg *gmail.Message) (EmailMessage, error) {
	email := EmailMessage{
		ID:      msg.Id,
		Headers: make(map[string]string),
	}

	// Parse headers
	for _, header := range msg.Payload.Headers {
		email.Headers[header.Name] = header.Value

		switch header.Name {
		case "Subject":
			email.Subject = header.Value
		case "From":
			email.From = header.Value
		case "To":
			email.To = strings.Split(header.Value, ",")
		case "Cc":
			email.CC = strings.Split(header.Value, ",")
		}
	}

	// Parse body
	if err := f.parseGmailBody(msg.Payload, &email); err != nil {
		return email, err
	}

	return email, nil
}

// parseGmailBody recursively parses Gmail message body parts
func (f *GmailAPIFetcher) parseGmailBody(part *gmail.MessagePart, email *EmailMessage) error {
	if part.Body != nil && part.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(part.Body.Data)
		if err != nil {
			return fmt.Errorf("failed to decode body data: %w", err)
		}

		content := string(data)

		switch part.MimeType {
		case "text/plain":
			email.Body = content
		case "text/html":
			email.HTMLBody = content
		}
	}

	// Handle multipart messages
	if part.Parts != nil {
		for _, subPart := range part.Parts {
			if err := f.parseGmailBody(subPart, email); err != nil {
				return err
			}
		}
	}

	return nil
}

// Close closes the Gmail API fetcher
func (f *GmailAPIFetcher) Close() error {
	// Gmail API service doesn't need explicit closing
	return nil
}

// FetchNewEmails fetches new emails using IMAP
func (f *IMAPFetcher) FetchNewEmails(ctx context.Context) ([]EmailMessage, error) {
	// Select INBOX
	_, err := f.client.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("failed to select INBOX: %w", err)
	}

	// Search for messages since last check
	criteria := imap.NewSearchCriteria()
	criteria.Since = f.lastCheck

	uids, err := f.client.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	if len(uids) == 0 {
		f.lastCheck = time.Now()
		return []EmailMessage{}, nil
	}

	// Fetch messages
	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	messages := make(chan *imap.Message, len(uids))
	done := make(chan error, 1)

	go func() {
		done <- f.client.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchBody, imap.FetchUid}, messages)
	}()

	var emails []EmailMessage

	for msg := range messages {
		email, err := f.parseIMAPMessage(msg)
		if err != nil {
			logrus.Warnf("Failed to parse IMAP message: %v", err)
			continue
		}
		emails = append(emails, email)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	f.lastCheck = time.Now()
	return emails, nil
}

// parseIMAPMessage parses an IMAP message into EmailMessage
func (f *IMAPFetcher) parseIMAPMessage(msg *imap.Message) (EmailMessage, error) {
	email := EmailMessage{
		Headers: make(map[string]string),
	}

	if msg.Envelope != nil {
		email.Subject = msg.Envelope.Subject
		if msg.Envelope.From != nil && len(msg.Envelope.From) > 0 {
			email.From = msg.Envelope.From[0].Address()
		}
		if msg.Envelope.To != nil && len(msg.Envelope.To) > 0 {
			for _, addr := range msg.Envelope.To {
				email.To = append(email.To, addr.Address())
			}
		}
	}

	// Parse body
	if err := f.parseIMAPBody(msg, &email); err != nil {
		return email, err
	}

	return email, nil
}

// parseIMAPBody parses IMAP message body
func (f *IMAPFetcher) parseIMAPBody(msg *imap.Message, email *EmailMessage) error {
	if msg.Body == nil {
		return nil
	}

	section := &imap.BodySectionName{}
	r := msg.GetBody(section)
	if r == nil {
		return fmt.Errorf("failed to get message body")
	}

	entity, err := message.Read(r)
	if err != nil {
		return fmt.Errorf("failed to read message: %w", err)
	}

	// Parse multipart message
	if mr := entity.MultipartReader(); mr != nil {
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read part: %w", err)
			}

			content, err := io.ReadAll(p.Body)
			if err != nil {
				return fmt.Errorf("failed to read part body: %w", err)
			}

			contentType := p.Header.Get("Content-Type")
			if strings.Contains(contentType, "text/plain") {
				email.Body = string(content)
			} else if strings.Contains(contentType, "text/html") {
				email.HTMLBody = string(content)
			}
		}
	} else {
		// Single part message
		content, err := io.ReadAll(entity.Body)
		if err != nil {
			return fmt.Errorf("failed to read message body: %w", err)
		}

		contentType := entity.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/plain") {
			email.Body = string(content)
		} else if strings.Contains(contentType, "text/html") {
			email.HTMLBody = string(content)
		}
	}

	return nil
}

// Close closes the IMAP fetcher
func (f *IMAPFetcher) Close() error {
	return f.client.Logout()
}

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
func (p *EmailParser) ParseAndMatchEmail(email EmailMessage) (*model.ForwardRule, error) {
	// Extract keyword from subject
	keyword, err := p.ExtractKeyword(email.Subject)
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

// ExtractKeyword extracts the keyword from email subject
// Expected format: "<keyword> - <recipient_name>"
func (p *EmailParser) ExtractKeyword(subject string) (string, error) {
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
func (p *EmailParser) findMatchingRule(keyword string) (*model.ForwardRule, error) {
	var rule model.ForwardRule

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
func (p *EmailParser) GetAllRules() ([]model.ForwardRule, error) {
	var rules []model.ForwardRule
	result := p.db.Find(&rules)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get rules: %w", result.Error)
	}
	return rules, nil
}

// GetEnabledRules returns all enabled forwarding rules
func (p *EmailParser) GetEnabledRules() ([]model.ForwardRule, error) {
	var rules []model.ForwardRule
	result := p.db.Where("enabled = ?", true).Find(&rules)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get enabled rules: %w", result.Error)
	}
	return rules, nil
}

// IsEmailProcessed checks if an email has already been processed
func (p *EmailParser) IsEmailProcessed(messageID string) (bool, error) {
	var processed model.ProcessedEmail
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
	processed := model.ProcessedEmail{
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
	log := model.ForwardLog{
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

// EmailForwarder handles forwarding emails via Gmail API
type EmailForwarder struct {
	service   *gmail.Service
	userEmail string
	config    *config.GmailConfig
}

// NewEmailForwarder creates a new email forwarder
func NewEmailForwarder(cfg *config.GmailConfig) (*EmailForwarder, error) {
	ctx := context.Background()

	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       []string{gmail.GmailSendScope},
		Endpoint:     google.Endpoint,
	}

	// Create token source from refresh token
	token := &oauth2.Token{
		RefreshToken: cfg.RefreshToken,
	}

	tokenSource := oauth2Config.TokenSource(ctx, token)

	// Create Gmail service
	service, err := gmail.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	return &EmailForwarder{
		service:   service,
		userEmail: cfg.UserEmail,
		config:    cfg,
	}, nil
}

// ForwardEmail forwards an email to the target address
func (f *EmailForwarder) ForwardEmail(ctx context.Context, originalEmail EmailMessage, targetEmail string) error {
	// Create the forwarded email
	forwardedEmail, err := f.createForwardedEmail(originalEmail, targetEmail)
	if err != nil {
		return fmt.Errorf("failed to create forwarded email: %w", err)
	}

	// Encode the email
	encodedEmail := base64.URLEncoding.EncodeToString([]byte(forwardedEmail))

	// Create the message
	message := &gmail.Message{
		Raw: encodedEmail,
	}

	// Send the email with retry logic
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		_, err := f.service.Users.Messages.Send(f.userEmail, message).Do()
		if err == nil {
			logrus.Infof("Successfully forwarded email %s to %s", originalEmail.ID, targetEmail)
			return nil
		}

		lastErr = err
		logrus.Warnf("Failed to forward email (attempt %d/%d): %v", attempt, 3, err)

		// Check if it's a rate limit error
		if strings.Contains(err.Error(), "quota") || strings.Contains(err.Error(), "rate") {
			// Wait with exponential backoff
			waitTime := time.Duration(attempt*attempt) * time.Second
			logrus.Infof("Rate limited, waiting %v before retry", waitTime)
			time.Sleep(waitTime)
		} else {
			// For other errors, don't retry
			break
		}
	}

	return fmt.Errorf("failed to forward email after 3 attempts: %w", lastErr)
}

// createForwardedEmail creates a forwarded email with proper headers
func (f *EmailForwarder) createForwardedEmail(original EmailMessage, targetEmail string) (string, error) {
	var emailBuilder strings.Builder

	// Add headers
	emailBuilder.WriteString(fmt.Sprintf("From: %s\r\n", f.userEmail))
	emailBuilder.WriteString(fmt.Sprintf("To: %s\r\n", targetEmail))
	emailBuilder.WriteString(fmt.Sprintf("Subject: Fwd: %s\r\n", original.Subject))
	emailBuilder.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	emailBuilder.WriteString("MIME-Version: 1.0\r\n")
	emailBuilder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	emailBuilder.WriteString("Content-Transfer-Encoding: 7bit\r\n")

	// Add original headers as references
	if original.From != "" {
		emailBuilder.WriteString(fmt.Sprintf("X-Original-From: %s\r\n", original.From))
	}
	if len(original.To) > 0 {
		emailBuilder.WriteString(fmt.Sprintf("X-Original-To: %s\r\n", strings.Join(original.To, ", ")))
	}
	if len(original.CC) > 0 {
		emailBuilder.WriteString(fmt.Sprintf("X-Original-Cc: %s\r\n", strings.Join(original.CC, ", ")))
	}
	emailBuilder.WriteString(fmt.Sprintf("X-Original-Message-ID: %s\r\n", original.ID))
	emailBuilder.WriteString(fmt.Sprintf("X-Forwarded-At: %s\r\n", time.Now().Format(time.RFC3339)))

	emailBuilder.WriteString("\r\n")

	// Add forwarded content
	emailBuilder.WriteString("---------- Forwarded message ----------\r\n")
	emailBuilder.WriteString(fmt.Sprintf("From: %s\r\n", original.From))
	if len(original.To) > 0 {
		emailBuilder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(original.To, ", ")))
	}
	if len(original.CC) > 0 {
		emailBuilder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(original.CC, ", ")))
	}
	emailBuilder.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	emailBuilder.WriteString(fmt.Sprintf("Subject: %s\r\n", original.Subject))
	emailBuilder.WriteString(fmt.Sprintf("Message-ID: %s\r\n", original.ID))
	emailBuilder.WriteString("\r\n")

	// Add original body
	if original.Body != "" {
		emailBuilder.WriteString(original.Body)
	} else if original.HTMLBody != "" {
		// Convert HTML to plain text (simple approach)
		plainText := f.htmlToPlainText(original.HTMLBody)
		emailBuilder.WriteString(plainText)
	} else {
		emailBuilder.WriteString("[No text content available]\r\n")
	}

	return emailBuilder.String(), nil
}

// htmlToPlainText converts HTML to plain text (simple implementation)
func (f *EmailForwarder) htmlToPlainText(html string) string {
	// Remove HTML tags (simple approach)
	// In a production environment, you might want to use a proper HTML parser
	text := html

	// Remove common HTML tags
	replacements := []struct {
		from string
		to   string
	}{
		{"<br>", "\r\n"},
		{"<br/>", "\r\n"},
		{"<br />", "\r\n"},
		{"<p>", "\r\n"},
		{"</p>", "\r\n"},
		{"<div>", "\r\n"},
		{"</div>", "\r\n"},
		{"&nbsp;", " "},
		{"&amp;", "&"},
		{"&lt;", "<"},
		{"&gt;", ">"},
		{"&quot;", "\""},
	}

	for _, replacement := range replacements {
		text = strings.ReplaceAll(text, replacement.from, replacement.to)
	}

	// Remove remaining HTML tags using regex
	re := regexp.MustCompile(`<[^>]*>`)
	text = re.ReplaceAllString(text, "")

	// Clean up whitespace
	text = strings.TrimSpace(text)

	// Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "\n", "\r\n")

	return text
}

// TestConnection tests the Gmail API connection
func (f *EmailForwarder) TestConnection(ctx context.Context) error {
	// Try to get user profile to test connection
	_, err := f.service.Users.GetProfile(f.userEmail).Do()
	if err != nil {
		return fmt.Errorf("failed to test Gmail API connection: %w", err)
	}
	return nil
}

// Close closes the forwarder (no-op for Gmail API)
func (f *EmailForwarder) Close() error {
	return nil
}
