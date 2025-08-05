package scheduler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	schedulerSvc "smart-mail-relay-go/internal/service/scheduler"
)

// Start starts the email processing scheduler
func Start(s *schedulerSvc.Scheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := s.Start(); err != nil {
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
}

// Stop stops the email processing scheduler
func Stop(s *schedulerSvc.Scheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := s.Stop(); err != nil {
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
}
