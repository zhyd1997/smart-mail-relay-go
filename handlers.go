package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	db        *gorm.DB
	parser    *EmailParser
	scheduler *Scheduler
	metrics   *Metrics
}

// NewHandlers creates new HTTP handlers
func NewHandlers(db *gorm.DB, parser *EmailParser, scheduler *Scheduler, metrics *Metrics) *Handlers {
	return &Handlers{
		db:        db,
		parser:    parser,
		scheduler: scheduler,
		metrics:   metrics,
	}
}

// SetupRoutes sets up all HTTP routes
func (h *Handlers) SetupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/healthz", h.HealthCheck)

	// Prometheus metrics
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes
	api := router.Group("/api/v1")
	{
		// Forward rules
		api.GET("/rules", h.GetRules)
		api.POST("/rules", h.CreateRule)
		api.GET("/rules/:id", h.GetRule)
		api.PUT("/rules/:id", h.UpdateRule)
		api.DELETE("/rules/:id", h.DeleteRule)
		api.PATCH("/rules/:id/enable", h.EnableRule)
		api.PATCH("/rules/:id/disable", h.DisableRule)

		// Forward logs
		api.GET("/logs", h.GetLogs)
		api.GET("/logs/:id", h.GetLog)

		// Scheduler control
		api.POST("/scheduler/start", h.StartScheduler)
		api.POST("/scheduler/stop", h.StopScheduler)
		api.POST("/scheduler/run-once", h.RunOnce)
		api.GET("/scheduler/status", h.GetSchedulerStatus)
	}
}

// HealthCheck handles health check requests
func (h *Handlers) HealthCheck(c *gin.Context) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Database:  "ok",
		Gmail:     "ok",
		Metrics:   make(map[string]string),
	}

	// Check database connection
	if err := h.db.Raw("SELECT 1").Error; err != nil {
		response.Status = "error"
		response.Database = "error"
		logrus.Errorf("Database health check failed: %v", err)
	}

	// Check scheduler status
	if h.scheduler.IsRunning() {
		response.Metrics["scheduler"] = "running"
		response.Metrics["next_run"] = h.scheduler.GetNextRun().Format(time.RFC3339)
		response.Metrics["last_run"] = h.scheduler.GetLastRun().Format(time.RFC3339)
	} else {
		response.Metrics["scheduler"] = "stopped"
	}

	// Add Prometheus metrics (counters don't have a Get method, so we'll show 0 for now)
	response.Metrics["pull_count"] = "0"
	response.Metrics["match_count"] = "0"
	response.Metrics["forward_successes"] = "0"
	response.Metrics["forward_failures"] = "0"

	statusCode := http.StatusOK
	if response.Status == "error" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// GetRules returns all forwarding rules
func (h *Handlers) GetRules(c *gin.Context) {
	rules, err := h.parser.GetAllRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch rules",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Convert to response format
	var responses []ForwardRuleResponse
	for _, rule := range rules {
		responses = append(responses, ForwardRuleResponse{
			ID:          rule.ID,
			Keyword:     rule.Keyword,
			TargetEmail: rule.TargetEmail,
			Enabled:     rule.Enabled,
			CreatedAt:   rule.CreatedAt,
			UpdatedAt:   rule.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, responses)
}

// CreateRule creates a new forwarding rule
func (h *Handlers) CreateRule(c *gin.Context) {
	var req ForwardRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Set default enabled value
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	rule := ForwardRule{
		Keyword:     req.Keyword,
		TargetEmail: req.TargetEmail,
		Enabled:     enabled,
	}

	if err := h.db.Create(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to create rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	response := ForwardRuleResponse{
		ID:          rule.ID,
		Keyword:     rule.Keyword,
		TargetEmail: rule.TargetEmail,
		Enabled:     rule.Enabled,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetRule returns a specific forwarding rule
func (h *Handlers) GetRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid rule ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var rule ForwardRule
	if err := h.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Rule not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	response := ForwardRuleResponse{
		ID:          rule.ID,
		Keyword:     rule.Keyword,
		TargetEmail: rule.TargetEmail,
		Enabled:     rule.Enabled,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateRule updates a forwarding rule
func (h *Handlers) UpdateRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid rule ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req ForwardRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var rule ForwardRule
	if err := h.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Rule not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Update fields
	rule.Keyword = req.Keyword
	rule.TargetEmail = req.TargetEmail
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}

	if err := h.db.Save(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to update rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	response := ForwardRuleResponse{
		ID:          rule.ID,
		Keyword:     rule.Keyword,
		TargetEmail: rule.TargetEmail,
		Enabled:     rule.Enabled,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteRule deletes a forwarding rule
func (h *Handlers) DeleteRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid rule ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.db.Delete(&ForwardRule{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to delete rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// EnableRule enables a forwarding rule
func (h *Handlers) EnableRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid rule ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.db.Model(&ForwardRule{}).Where("id = ?", id).Update("enabled", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to enable rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// DisableRule disables a forwarding rule
func (h *Handlers) DisableRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid rule ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.db.Model(&ForwardRule{}).Where("id = ?", id).Update("enabled", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to disable rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetLogs returns forward logs with pagination
func (h *Handlers) GetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	var logs []ForwardLog
	var total int64

	// Get total count
	if err := h.db.Model(&ForwardLog{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to count logs",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Get logs with pagination
	if err := h.db.Preload("Rule").Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch logs",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Convert to response format
	var responses []ForwardLogResponse
	for _, log := range logs {
		response := ForwardLogResponse{
			ID:        log.ID,
			MessageID: log.MessageID,
			RuleID:    log.RuleID,
			Status:    log.Status,
			ErrorMsg:  log.ErrorMsg,
			CreatedAt: log.CreatedAt,
		}

		if log.Rule != nil {
			response.Rule = &ForwardRuleResponse{
				ID:          log.Rule.ID,
				Keyword:     log.Rule.Keyword,
				TargetEmail: log.Rule.TargetEmail,
				Enabled:     log.Rule.Enabled,
				CreatedAt:   log.Rule.CreatedAt,
				UpdatedAt:   log.Rule.UpdatedAt,
			}
		}

		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": responses,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetLog returns a specific forward log
func (h *Handlers) GetLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid log ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var log ForwardLog
	if err := h.db.Preload("Rule").First(&log, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Log not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch log",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	response := ForwardLogResponse{
		ID:        log.ID,
		MessageID: log.MessageID,
		RuleID:    log.RuleID,
		Status:    log.Status,
		ErrorMsg:  log.ErrorMsg,
		CreatedAt: log.CreatedAt,
	}

	if log.Rule != nil {
		response.Rule = &ForwardRuleResponse{
			ID:          log.Rule.ID,
			Keyword:     log.Rule.Keyword,
			TargetEmail: log.Rule.TargetEmail,
			Enabled:     log.Rule.Enabled,
			CreatedAt:   log.Rule.CreatedAt,
			UpdatedAt:   log.Rule.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

// StartScheduler starts the email processing scheduler
func (h *Handlers) StartScheduler(c *gin.Context) {
	if err := h.scheduler.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "scheduler_error",
			Message: "Failed to start scheduler",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Scheduler started successfully",
		"status":  "running",
	})
}

// StopScheduler stops the email processing scheduler
func (h *Handlers) StopScheduler(c *gin.Context) {
	if err := h.scheduler.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "scheduler_error",
			Message: "Failed to stop scheduler",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Scheduler stopped successfully",
		"status":  "stopped",
	})
}

// RunOnce runs the email processing once
func (h *Handlers) RunOnce(c *gin.Context) {
	if err := h.scheduler.RunOnce(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "scheduler_error",
			Message: "Failed to run email processing",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email processing completed successfully",
	})
}

// GetSchedulerStatus returns the current scheduler status
func (h *Handlers) GetSchedulerStatus(c *gin.Context) {
	status := "stopped"
	if h.scheduler.IsRunning() {
		status = "running"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   status,
		"next_run": h.scheduler.GetNextRun(),
		"last_run": h.scheduler.GetLastRun(),
	})
}
