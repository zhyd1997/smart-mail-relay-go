package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	metricsPkg "smart-mail-relay-go/internal/metrics"
	"smart-mail-relay-go/internal/service"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	db        *gorm.DB
	parser    *service.EmailParser
	scheduler *service.Scheduler
	metrics   *metricsPkg.Metrics
}

// NewHandlers creates new HTTP handlers
func NewHandlers(db *gorm.DB, parser *service.EmailParser, scheduler *service.Scheduler, metrics *metricsPkg.Metrics) *Handlers {
	return &Handlers{
		db:        db,
		parser:    parser,
		scheduler: scheduler,
		metrics:   metrics,
	}
}

// SetupRoutes sets up all HTTP routes
func (h *Handlers) SetupRoutes(router *gin.Engine) {
	router.GET("/healthz", h.HealthCheck)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := router.Group("/api/v1")
	{
		api.GET("/rules", h.GetRules)
		api.POST("/rules", h.CreateRule)
		api.GET("/rules/:id", h.GetRule)
		api.PUT("/rules/:id", h.UpdateRule)
		api.DELETE("/rules/:id", h.DeleteRule)
		api.PATCH("/rules/:id/enable", h.EnableRule)
		api.PATCH("/rules/:id/disable", h.DisableRule)

		api.GET("/logs", h.GetLogs)
		api.GET("/logs/:id", h.GetLog)

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

	if err := h.db.Raw("SELECT 1").Error; err != nil {
		response.Status = "error"
		response.Database = "error"
		logrus.Errorf("Database health check failed: %v", err)
	}

	if h.scheduler.IsRunning() {
		response.Metrics["scheduler"] = "running"
		response.Metrics["next_run"] = h.scheduler.GetNextRun().Format(time.RFC3339)
		response.Metrics["last_run"] = h.scheduler.GetLastRun().Format(time.RFC3339)
	} else {
		response.Metrics["scheduler"] = "stopped"
	}

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
