package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"smart-mail-relay-go/internal/model"
)

// GetRules returns all forwarding rules
func (h *Handlers) GetRules(c *gin.Context) {
	rules, err := h.parser.GetAllRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch rules",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	var responses []ForwardRuleResponse
	for _, rule := range rules {
		responses = append(responses, ForwardRuleResponse{
			ID:          rule.ID,
			Keyword:     rule.Keyword,
			TargetEmail: rule.TargetEmail,
			Enabled:     rule.Enabled,
			CreatedAt:   rule.CreatedAt,
			UpdatedAt:   rule.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, responses)
}

// CreateRule creates a new forwarding rule
func (h *Handlers) CreateRule(c *gin.Context) {
	var req ForwardRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	rule := model.ForwardRule{
		Keyword:     req.Keyword,
		TargetEmail: req.TargetEmail,
		Enabled:     enabled,
	}

	if err := h.db.Create(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to create rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	response := ForwardRuleResponse{
		ID:          rule.ID,
		Keyword:     rule.Keyword,
		TargetEmail: rule.TargetEmail,
		Enabled:     rule.Enabled,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetRule returns a specific forwarding rule
func (h *Handlers) GetRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid rule ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var rule model.ForwardRule
	if err := h.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Rule not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	response := ForwardRuleResponse{
		ID:          rule.ID,
		Keyword:     rule.Keyword,
		TargetEmail: rule.TargetEmail,
		Enabled:     rule.Enabled,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateRule updates a forwarding rule
func (h *Handlers) UpdateRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid rule ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req ForwardRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request body",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var rule model.ForwardRule
	if err := h.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Rule not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	rule.Keyword = req.Keyword
	rule.TargetEmail = req.TargetEmail
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}

	if err := h.db.Save(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to update rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	response := ForwardRuleResponse{
		ID:          rule.ID,
		Keyword:     rule.Keyword,
		TargetEmail: rule.TargetEmail,
		Enabled:     rule.Enabled,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteRule deletes a forwarding rule
func (h *Handlers) DeleteRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid rule ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.db.Delete(&model.ForwardRule{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to delete rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// EnableRule enables a forwarding rule
func (h *Handlers) EnableRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid rule ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.db.Model(&model.ForwardRule{}).Where("id = ?", id).Update("enabled", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to enable rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// DisableRule disables a forwarding rule
func (h *Handlers) DisableRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid rule ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.db.Model(&model.ForwardRule{}).Where("id = ?", id).Update("enabled", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to disable rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.Status(http.StatusNoContent)
}
