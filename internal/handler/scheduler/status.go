package scheduler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	schedulerSvc "smart-mail-relay-go/internal/service/scheduler"
)

// Status returns the current scheduler status
func Status(s *schedulerSvc.Scheduler) gin.HandlerFunc {
	return func(c *gin.Context) {
		state := "stopped"
		if s.IsRunning() {
			state = "running"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   state,
			"next_run": s.GetNextRun(),
			"last_run": s.GetLastRun(),
		})
	}
}
