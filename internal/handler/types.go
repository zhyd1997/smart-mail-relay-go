package handler

import "time"

// ForwardRuleRequest represents the request structure for creating/updating forward rules
type ForwardRuleRequest struct {
	Keyword     string `json:"keyword" binding:"required"`
	TargetEmail string `json:"target_email" binding:"required,email"`
	Enabled     *bool  `json:"enabled"`
}

// ForwardRuleResponse represents the response structure for forward rules
type ForwardRuleResponse struct {
	ID          uint      `json:"id"`
	Keyword     string    `json:"keyword"`
	TargetEmail string    `json:"target_email"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ForwardLogResponse represents the response structure for forward logs
type ForwardLogResponse struct {
	ID        uint                 `json:"id"`
	MessageID string               `json:"message_id"`
	RuleID    *uint                `json:"rule_id"`
	Status    string               `json:"status"`
	ErrorMsg  string               `json:"error_msg"`
	CreatedAt time.Time            `json:"created_at"`
	Rule      *ForwardRuleResponse `json:"rule,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Database  string            `json:"database"`
	Gmail     string            `json:"gmail"`
	Metrics   map[string]string `json:"metrics,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
