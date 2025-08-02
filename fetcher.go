package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
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
)

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
func NewGmailAPIFetcher(config *GmailConfig) (*GmailAPIFetcher, error) {
	ctx := context.Background()

	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Scopes:       []string{gmail.GmailReadonlyScope},
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

	return &GmailAPIFetcher{
		service:   service,
		userEmail: config.UserEmail,
		lastCheck: time.Now().Add(-24 * time.Hour), // Start with emails from last 24 hours
	}, nil
}

// NewIMAPFetcher creates a new IMAP fetcher
func NewIMAPFetcher(config *GmailConfig) (*IMAPFetcher, error) {
	// Connect to IMAP server
	c, err := client.DialTLS(fmt.Sprintf("%s:%d", config.IMAPHost, config.IMAPPort), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %w", err)
	}

	// Login
	if err := c.Login(config.IMAPUser, config.IMAPPassword); err != nil {
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
