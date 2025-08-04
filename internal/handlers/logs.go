package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"smart-mail-relay-go/internal/models"
)

// GetLogs returns all forward logs
func (h *Handlers) GetLogs(c *gin.Context) {
	var logs []models.ForwardLog
	if err := h.db.Order("created_at desc").Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch logs",
			Code:    http.StatusInternalServerError,
		})
		return
	}
	c.JSON(http.StatusOK, logs)
}

// GetLog returns a single log by ID
func (h *Handlers) GetLog(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_id", Message: "Invalid log ID", Code: http.StatusBadRequest})
		return
	}
	var log models.ForwardLog
	if err := h.db.First(&log, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "Log not found", Code: http.StatusNotFound})
		return
	}
	c.JSON(http.StatusOK, log)
}
