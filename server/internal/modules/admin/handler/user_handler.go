package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) List(c *gin.Context) {
	var query dto.UserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	items, total, err := h.svc.List(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "USER_LIST_FAILED", err.Error())
		return
	}

	response.Paginated(c, items, response.Meta{
		Total:   total,
		Page:    query.Page,
		PerPage: query.Limit,
	})
}

func (h *UserHandler) Detail(c *gin.Context) {
	userID := c.Param("id")

	detail, err := h.svc.Detail(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
		return
	}

	response.OK(c, detail)
}

func (h *UserHandler) Update(c *gin.Context) {
	userID := c.Param("id")

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.svc.Update(c.Request.Context(), userID, req); err != nil {
		response.Fail(c, http.StatusBadRequest, "USER_UPDATE_FAILED", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "user updated"})
}

func (h *UserHandler) Disable(c *gin.Context) {
	userID := c.Param("id")

	if err := h.svc.Disable(c.Request.Context(), userID); err != nil {
		response.Fail(c, http.StatusBadRequest, "USER_DISABLE_FAILED", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "user disabled"})
}

func (h *UserHandler) Enable(c *gin.Context) {
	userID := c.Param("id")

	if err := h.svc.Enable(c.Request.Context(), userID); err != nil {
		response.Fail(c, http.StatusBadRequest, "USER_ENABLE_FAILED", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "user enabled"})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	userID := c.Param("id")

	tempPassword, err := h.svc.ResetPassword(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "RESET_PASSWORD_FAILED", err.Error())
		return
	}

	response.OK(c, gin.H{"temporary_password": tempPassword})
}

func (h *UserHandler) Reset2FA(c *gin.Context) {
	userID := c.Param("id")
	if err := h.svc.Reset2FA(c.Request.Context(), userID); err != nil {
		response.Fail(c, http.StatusBadRequest, "RESET_2FA_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"message": "2fa reset successful"})
}

func (h *UserHandler) ListInvestments(c *gin.Context) {
	var query dto.InvestmentListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_QUERY", err.Error())
		return
	}

	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	items, total, err := h.svc.ListInvestments(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INVESTMENT_LIST_FAILED", err.Error())
		return
	}

	response.Paginated(c, items, response.Meta{
		Total:   total,
		Page:    query.Page,
		PerPage: query.Limit,
	})
}

func (h *UserHandler) AdjustBalance(c *gin.Context) {
	userID := c.Param("id")
	adminID := c.GetString("admin_id")

	var req dto.AdjustBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.svc.AdjustBalance(c.Request.Context(), userID, req, adminID); err != nil {
		response.Fail(c, http.StatusBadRequest, "ADJUST_BALANCE_FAILED", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "balance adjusted"})
}
