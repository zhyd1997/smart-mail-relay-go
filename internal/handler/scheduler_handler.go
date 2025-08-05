package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
