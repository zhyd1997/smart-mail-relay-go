package router

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"smart-mail-relay-go/internal/handlers"
)

// Setup configures routes and middleware
func Setup(h *handlers.Handlers) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggerMiddleware())

	r.GET("/healthz", h.HealthCheck)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := r.Group("/api/v1")
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

	return r
}

func loggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}
