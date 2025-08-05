package scheduler

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	service "smart-mail-relay-go/internal/service"
)

// processEmails is the main processing function that runs periodically
func (s *Scheduler) processEmails() {
	s.wg.Add(1)
	defer s.wg.Done()

	logrus.Info("Starting email processing cycle")

	s.mu.RLock()
	if !s.isRunning {
		s.mu.RUnlock()
		logrus.Info("Scheduler not running, skipping processing cycle")
		return
	}
	s.mu.RUnlock()

	startTime := time.Now()

	s.metrics.PullCount.Inc()

	emails, err := s.fetcher.FetchNewEmails(s.ctx)
	if err != nil {
		logrus.Errorf("Failed to fetch emails: %v", err)
		s.metrics.ForwardFailures.Inc()
		return
	}

	logrus.Infof("Fetched %d new emails", len(emails))

	for _, email := range emails {
		if err := s.processEmail(email); err != nil {
			logrus.Errorf("Failed to process email %s: %v", email.ID, err)
		}
	}

	duration := time.Since(startTime)
	logrus.Infof("Email processing cycle completed in %v", duration)
}

// processEmail processes a single email
func (s *Scheduler) processEmail(email service.EmailMessage) error {
	select {
	case <-s.ctx.Done():
		return fmt.Errorf("context cancelled")
	default:
	}

	processed, err := s.parser.IsEmailProcessed(email.ID)
	if err != nil {
		return fmt.Errorf("failed to check if email is processed: %w", err)
	}

	if processed {
		logrus.Debugf("Email %s already processed, skipping", email.ID)
		return nil
	}

	rule, err := s.parser.ParseAndMatchEmail(email)
	if err != nil {
		s.parser.LogForwardAttempt(email.ID, nil, "error", err.Error())
		return fmt.Errorf("failed to parse and match email: %w", err)
	}

	if rule == nil {
		s.parser.LogForwardAttempt(email.ID, nil, "skipped", "No matching rule found")
		s.parser.MarkEmailAsProcessed(email.ID)
		return nil
	}

	s.metrics.MatchCount.Inc()

	err = s.forwarder.ForwardEmail(s.ctx, email, rule.TargetEmail)
	if err != nil {
		s.parser.LogForwardAttempt(email.ID, &rule.ID, "failure", err.Error())
		s.metrics.ForwardFailures.Inc()
		return fmt.Errorf("failed to forward email: %w", err)
	}

	if err := s.parser.MarkEmailAsProcessed(email.ID); err != nil {
		logrus.Errorf("Failed to mark email as processed: %v", err)
	}

	s.parser.LogForwardAttempt(email.ID, &rule.ID, "success", "")
	s.metrics.ForwardSuccesses.Inc()

	logrus.Infof("Successfully processed email %s with rule %s", email.ID, rule.Keyword)
	return nil
}
