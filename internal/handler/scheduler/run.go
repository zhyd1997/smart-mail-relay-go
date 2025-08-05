package scheduler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	schedulerSvc "smart-mail-relay-go/internal/service/scheduler"
)

// RunOnce runs the email processing once
func RunOnce(s *schedulerSvc.Scheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := s.RunOnce(); err != nil {
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
}
