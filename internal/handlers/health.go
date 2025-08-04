package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"smart-mail-relay-go/internal/models"
)

// HealthCheck handles health check requests
func (h *Handlers) HealthCheck(c *gin.Context) {
	response := models.HealthResponse{
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
