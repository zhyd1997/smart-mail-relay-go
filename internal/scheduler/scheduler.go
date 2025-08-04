package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"smart-mail-relay-go/internal/config"
	"smart-mail-relay-go/internal/fetcher"
	"smart-mail-relay-go/internal/forwarder"
	"smart-mail-relay-go/internal/metrics"
	"smart-mail-relay-go/internal/models"
	"smart-mail-relay-go/internal/services/parser"
)

// Scheduler manages the periodic email processing
type Scheduler struct {
	cron      *cron.Cron
	entryID   cron.EntryID
	config    *config.SchedulerConfig
	fetcher   fetcher.EmailFetcher
	parser    *parser.EmailParser
	forwarder *forwarder.EmailForwarder
	metrics   *metrics.Metrics
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	isRunning bool
	mu        sync.RWMutex
}

// NewScheduler creates a new scheduler
func NewScheduler(cfg *config.SchedulerConfig, fetcher fetcher.EmailFetcher, parser *parser.EmailParser, forwarder *forwarder.EmailForwarder, metrics *metrics.Metrics) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		cron:      cron.New(cron.WithSeconds()),
		config:    cfg,
		fetcher:   fetcher,
		parser:    parser,
		forwarder: forwarder,
		metrics:   metrics,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("scheduler is already running")
	}

	// Schedule the job to run every N minutes
	schedule := fmt.Sprintf("0 */%d * * * *", s.config.IntervalMinutes)

	entryID, err := s.cron.AddFunc(schedule, s.processEmails)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.entryID = entryID
	s.cron.Start()
	s.isRunning = true

	logrus.Infof("Scheduler started with interval: %d minutes", s.config.IntervalMinutes)
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	// Cancel context to stop any running operations
	s.cancel()

	// Stop the cron scheduler
	ctx := s.cron.Stop()

	// Wait for all jobs to complete
	select {
	case <-ctx.Done():
		logrus.Info("Scheduler stopped gracefully")
	case <-time.After(30 * time.Second):
		logrus.Warn("Scheduler stop timeout, forcing shutdown")
	}

	s.isRunning = false
	return nil
}

// IsRunning returns whether the scheduler is running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// processEmails is the main processing function that runs periodically
func (s *Scheduler) processEmails() {
	s.wg.Add(1)
	defer s.wg.Done()

	logrus.Info("Starting email processing cycle")

	// Check if scheduler is still running
	s.mu.RLock()
	if !s.isRunning {
		s.mu.RUnlock()
		logrus.Info("Scheduler not running, skipping processing cycle")
		return
	}
	s.mu.RUnlock()

	startTime := time.Now()

	// Increment pull count metric
	s.metrics.PullCount.Inc()

	// Fetch new emails
	emails, err := s.fetcher.FetchNewEmails(s.ctx)
	if err != nil {
		logrus.Errorf("Failed to fetch emails: %v", err)
		s.metrics.ForwardFailures.Inc()
		return
	}

	logrus.Infof("Fetched %d new emails", len(emails))

	// Process each email
	for _, email := range emails {
		if err := s.processEmail(email); err != nil {
			logrus.Errorf("Failed to process email %s: %v", email.ID, err)
		}
	}

	duration := time.Since(startTime)
	logrus.Infof("Email processing cycle completed in %v", duration)
}

// processEmail processes a single email
func (s *Scheduler) processEmail(email models.EmailMessage) error {
	// Check if context is cancelled
	select {
	case <-s.ctx.Done():
		return fmt.Errorf("context cancelled")
	default:
	}

	// Check if email has already been processed
	processed, err := s.parser.IsEmailProcessed(email.ID)
	if err != nil {
		return fmt.Errorf("failed to check if email is processed: %w", err)
	}

	if processed {
		logrus.Debugf("Email %s already processed, skipping", email.ID)
		return nil
	}

	// Parse and match email
	rule, err := s.parser.ParseAndMatchEmail(email)
	if err != nil {
		// Log the error but don't mark as processed to allow retry
		s.parser.LogForwardAttempt(email.ID, nil, "error", err.Error())
		return fmt.Errorf("failed to parse and match email: %w", err)
	}

	// If no matching rule found, log and mark as processed
	if rule == nil {
		s.parser.LogForwardAttempt(email.ID, nil, "skipped", "No matching rule found")
		s.parser.MarkEmailAsProcessed(email.ID)
		return nil
	}

	// Increment match count metric
	s.metrics.MatchCount.Inc()

	// Forward the email
	err = s.forwarder.ForwardEmail(s.ctx, email, rule.TargetEmail)
	if err != nil {
		// Log the failure
		s.parser.LogForwardAttempt(email.ID, &rule.ID, "failure", err.Error())
		s.metrics.ForwardFailures.Inc()
		return fmt.Errorf("failed to forward email: %w", err)
	}

	// Mark email as processed
	if err := s.parser.MarkEmailAsProcessed(email.ID); err != nil {
		logrus.Errorf("Failed to mark email as processed: %v", err)
	}

	// Log the success
	s.parser.LogForwardAttempt(email.ID, &rule.ID, "success", "")
	s.metrics.ForwardSuccesses.Inc()

	logrus.Infof("Successfully processed email %s with rule %s", email.ID, rule.Keyword)
	return nil
}

// RunOnce runs the email processing once (for manual triggering)
func (s *Scheduler) RunOnce() error {
	logrus.Info("Running email processing once")
	s.processEmails()
	return nil
}

// GetNextRun returns the time of the next scheduled run
func (s *Scheduler) GetNextRun() time.Time {
	if !s.isRunning {
		return time.Time{}
	}

	entry := s.cron.Entry(s.entryID)
	return entry.Next
}

// GetLastRun returns the time of the last run
func (s *Scheduler) GetLastRun() time.Time {
	if !s.isRunning {
		return time.Time{}
	}

	entry := s.cron.Entry(s.entryID)
	return entry.Prev
}

// Wait waits for the scheduler to stop
func (s *Scheduler) Wait() {
	s.wg.Wait()
}
