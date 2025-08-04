package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// StartScheduler starts the email scheduler
func (h *Handlers) StartScheduler(c *gin.Context) {
	if err := h.scheduler.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// StopScheduler stops the email scheduler
func (h *Handlers) StopScheduler(c *gin.Context) {
	if err := h.scheduler.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// RunOnce triggers the scheduler to run once
func (h *Handlers) RunOnce(c *gin.Context) {
	h.scheduler.RunOnce()
	c.Status(http.StatusOK)
}

// GetSchedulerStatus returns scheduler status
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
