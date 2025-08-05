package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"smart-mail-relay-go/config"
	metricsPkg "smart-mail-relay-go/internal/metrics"
	service "smart-mail-relay-go/internal/service"
)

// Scheduler manages the periodic email processing
type Scheduler struct {
	cron      *cron.Cron
	entryID   cron.EntryID
	config    *config.SchedulerConfig
	fetcher   service.EmailFetcher
	parser    *service.EmailParser
	forwarder *service.EmailForwarder
	metrics   *metricsPkg.Metrics
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	isRunning bool
	mu        sync.RWMutex
}

// New creates a new scheduler
func New(cfg *config.SchedulerConfig, fetcher service.EmailFetcher, parser *service.EmailParser, forwarder *service.EmailForwarder, metrics *metricsPkg.Metrics) *Scheduler {
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

	s.cancel()

	ctx := s.cron.Stop()

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
