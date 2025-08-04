package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"smart-mail-relay-go/internal/models"
)

// GetRules returns all forwarding rules
func (h *Handlers) GetRules(c *gin.Context) {
	rules, err := h.parser.GetAllRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch rules",
			Code:    http.StatusInternalServerError,
		})
		return
	}
	var responses []models.ForwardRuleResponse
	for _, rule := range rules {
		responses = append(responses, models.ForwardRuleResponse{
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
	var req models.ForwardRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
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

	rule := models.ForwardRule{
		Keyword:     req.Keyword,
		TargetEmail: req.TargetEmail,
		Enabled:     enabled,
	}
	if err := h.db.Create(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to create rule",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, models.ForwardRuleResponse{
		ID:          rule.ID,
		Keyword:     rule.Keyword,
		TargetEmail: rule.TargetEmail,
		Enabled:     rule.Enabled,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	})
}

// GetRule returns a single rule by ID
func (h *Handlers) GetRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_id", Message: "Invalid rule ID", Code: http.StatusBadRequest})
		return
	}
	var rule models.ForwardRule
	if err := h.db.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "Rule not found", Code: http.StatusNotFound})
		return
	}
	c.JSON(http.StatusOK, models.ForwardRuleResponse{
		ID:          rule.ID,
		Keyword:     rule.Keyword,
		TargetEmail: rule.TargetEmail,
		Enabled:     rule.Enabled,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	})
}

// UpdateRule updates an existing rule
func (h *Handlers) UpdateRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_id", Message: "Invalid rule ID", Code: http.StatusBadRequest})
		return
	}
	var rule models.ForwardRule
	if err := h.db.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "Rule not found", Code: http.StatusNotFound})
		return
	}
	var req models.ForwardRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Invalid request body", Code: http.StatusBadRequest})
		return
	}
	if req.Keyword != "" {
		rule.Keyword = req.Keyword
	}
	if req.TargetEmail != "" {
		rule.TargetEmail = req.TargetEmail
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if err := h.db.Save(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "database_error", Message: "Failed to update rule", Code: http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, models.ForwardRuleResponse{
		ID:          rule.ID,
		Keyword:     rule.Keyword,
		TargetEmail: rule.TargetEmail,
		Enabled:     rule.Enabled,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	})
}

// DeleteRule deletes a rule by ID
func (h *Handlers) DeleteRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_id", Message: "Invalid rule ID", Code: http.StatusBadRequest})
		return
	}
	if err := h.db.Delete(&models.ForwardRule{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "database_error", Message: "Failed to delete rule", Code: http.StatusInternalServerError})
		return
	}
	c.Status(http.StatusNoContent)
}

// EnableRule enables a rule by ID
func (h *Handlers) EnableRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_id", Message: "Invalid rule ID", Code: http.StatusBadRequest})
		return
	}
	if err := h.db.Model(&models.ForwardRule{}).Where("id = ?", id).Update("enabled", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "database_error", Message: "Failed to enable rule", Code: http.StatusInternalServerError})
		return
	}
	c.Status(http.StatusNoContent)
}

// DisableRule disables a rule by ID
func (h *Handlers) DisableRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_id", Message: "Invalid rule ID", Code: http.StatusBadRequest})
		return
	}
	if err := h.db.Model(&models.ForwardRule{}).Where("id = ?", id).Update("enabled", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "database_error", Message: "Failed to disable rule", Code: http.StatusInternalServerError})
		return
	}
	c.Status(http.StatusNoContent)
}
