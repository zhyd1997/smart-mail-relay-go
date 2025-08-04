package main

import (
	"context"
	"testing"
)

// dummyFetcher implements EmailFetcher but does nothing
type dummyFetcher struct{}

func (d *dummyFetcher) FetchNewEmails(ctx context.Context) ([]EmailMessage, error) { return nil, nil }
func (d *dummyFetcher) Close() error                                               { return nil }

func TestSchedulerRestart(t *testing.T) {
	config := &SchedulerConfig{IntervalMinutes: 60}
	sched := NewScheduler(config, &dummyFetcher{}, nil, nil, NewMetrics())

	if err := sched.Start(); err != nil {
		t.Fatalf("first start failed: %v", err)
	}
	if !sched.IsRunning() {
		t.Fatalf("scheduler should be running after Start")
	}
	if err := sched.Stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	if sched.IsRunning() {
		t.Fatalf("scheduler should not be running after Stop")
	}
	if err := sched.Start(); err != nil {
		t.Fatalf("second start failed: %v", err)
	}
	if !sched.IsRunning() {
		t.Fatalf("scheduler should be running after second Start")
	}
	// context should be active
	if sched.ctx == nil || sched.ctx.Err() != nil {
		t.Fatalf("scheduler context should be active after restart")
	}
	sched.Stop()
}
