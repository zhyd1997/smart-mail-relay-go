package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/sirupsen/logrus"
)

// EmailForwarder handles forwarding emails via Gmail API
type EmailForwarder struct {
	service   *gmail.Service
	userEmail string
	config    *GmailConfig
}

// NewEmailForwarder creates a new email forwarder
func NewEmailForwarder(config *GmailConfig) (*EmailForwarder, error) {
	ctx := context.Background()

	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Scopes:       []string{gmail.GmailSendScope},
		Endpoint:     google.Endpoint,
	}

	// Create token source from refresh token
	token := &oauth2.Token{
		RefreshToken: config.RefreshToken,
	}

	tokenSource := oauth2Config.TokenSource(ctx, token)

	// Create Gmail service
	service, err := gmail.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	return &EmailForwarder{
		service:   service,
		userEmail: config.UserEmail,
		config:    config,
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

	boundary := fmt.Sprintf("smart-relay-%d", time.Now().UnixNano())

	// Add headers
	emailBuilder.WriteString(fmt.Sprintf("From: %s\r\n", f.userEmail))
	emailBuilder.WriteString(fmt.Sprintf("To: %s\r\n", targetEmail))
	emailBuilder.WriteString(fmt.Sprintf("Subject: Fwd: %s\r\n", original.Subject))
	emailBuilder.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	emailBuilder.WriteString("MIME-Version: 1.0\r\n")

	if len(original.Attachments) > 0 {
		emailBuilder.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	} else {
		emailBuilder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		emailBuilder.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	}

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

	if len(original.Attachments) > 0 {
		// Start multipart body
		emailBuilder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		emailBuilder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		emailBuilder.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	}

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
		plainText := f.htmlToPlainText(original.HTMLBody)
		emailBuilder.WriteString(plainText)
	} else {
		emailBuilder.WriteString("[No text content available]\r\n")
	}

	if len(original.Attachments) > 0 {
		for _, att := range original.Attachments {
			emailBuilder.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
			emailBuilder.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", att.MIMEType, att.Filename))
			emailBuilder.WriteString("Content-Transfer-Encoding: base64\r\n")
			emailBuilder.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", att.Filename))
			encoded := base64.StdEncoding.EncodeToString(att.Data)
			emailBuilder.WriteString(encoded)
		}
		emailBuilder.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))
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
