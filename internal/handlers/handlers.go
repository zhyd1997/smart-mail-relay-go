package handlers

import (
	"gorm.io/gorm"

	"smart-mail-relay-go/internal/metrics"
	"smart-mail-relay-go/internal/scheduler"
	"smart-mail-relay-go/internal/services/parser"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	db        *gorm.DB
	parser    *parser.EmailParser
	scheduler *scheduler.Scheduler
	metrics   *metrics.Metrics
}

// NewHandlers creates new HTTP handlers
func NewHandlers(db *gorm.DB, p *parser.EmailParser, s *scheduler.Scheduler, m *metrics.Metrics) *Handlers {
	return &Handlers{db: db, parser: p, scheduler: s, metrics: m}
}
