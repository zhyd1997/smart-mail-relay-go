package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"smart-mail-relay-go/internal/model"
)

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

	var logs []model.ForwardLog
	var total int64

	if err := h.db.Model(&model.ForwardLog{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to count logs",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	if err := h.db.Preload("Rule").Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch logs",
			Code:    http.StatusInternalServerError,
		})
		return
	}

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

	var log model.ForwardLog
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
