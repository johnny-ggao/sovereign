package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/wallet/service"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
	"github.com/sovereign-fund/sovereign/pkg/cobo"
	"github.com/sovereign-fund/sovereign/pkg/pagination"
)

type WalletHandler struct {
	walletSvc service.WalletService
	provider  cobo.WalletProvider
}

func NewWalletHandler(walletSvc service.WalletService, provider cobo.WalletProvider) *WalletHandler {
	return &WalletHandler{walletSvc: walletSvc, provider: provider}
}

func (h *WalletHandler) GetWallets(c *gin.Context) {
	userID := c.GetString("user_id")

	overview, err := h.walletSvc.GetWallets(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, overview)
}

func (h *WalletHandler) GetDepositAddress(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.GetDepositAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	addr, err := h.walletSvc.GetDepositAddress(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, addr)
}

func (h *WalletHandler) Withdraw(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.walletSvc.Withdraw(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Created(c, resp)
}

func (h *WalletHandler) AddWhitelistAddress(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.AddWhitelistAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.walletSvc.AddWhitelistAddress(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Created(c, resp)
}

func (h *WalletHandler) GetWhitelistAddresses(c *gin.Context) {
	userID := c.GetString("user_id")

	addrs, err := h.walletSvc.GetWhitelistAddresses(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, addrs)
}

func (h *WalletHandler) RemoveWhitelistAddress(c *gin.Context) {
	userID := c.GetString("user_id")
	addressID := c.Param("id")

	if err := h.walletSvc.RemoveWhitelistAddress(c.Request.Context(), userID, addressID); err != nil {
		handleError(c, err)
		return
	}

	response.NoContent(c)
}

func (h *WalletHandler) ClaimEarnings(c *gin.Context) {
	userID := c.GetString("user_id")
	if err := h.walletSvc.ClaimEarnings(c.Request.Context(), userID); err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "earnings claimed"})
}

func (h *WalletHandler) GetTransactions(c *gin.Context) {
	userID := c.GetString("user_id")
	txType := c.Query("type")
	p := pagination.Parse(c)

	txs, total, err := h.walletSvc.GetTransactions(c.Request.Context(), userID, txType, p.Page, p.PerPage)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Paginated(c, txs, response.Meta{
		Total:   total,
		Page:    p.Page,
		PerPage: p.PerPage,
	})
}

func (h *WalletHandler) GetTransaction(c *gin.Context) {
	userID := c.GetString("user_id")
	txID := c.Param("id")

	tx, err := h.walletSvc.GetTransaction(c.Request.Context(), userID, txID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, tx)
}

func (h *WalletHandler) HandleWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "failed to read body")
		return
	}

	signature := c.GetHeader("Biz-Resp-Signature")
	valid, err := h.provider.VerifyWebhook(signature, body)
	if err != nil || !valid {
		response.Fail(c, http.StatusUnauthorized, "INVALID_SIGNATURE", "invalid webhook signature")
		return
	}

	// 记录原始 webhook payload 用于调试
	slog.Info("webhook received", slog.String("body", string(body)))

	// 解析 Cobo WaaS 2.0 webhook 格式
	payload, err := cobo.ParseWebhookPayload(body)
	if err != nil {
		slog.Error("parse webhook failed", slog.String("error", err.Error()), slog.String("body", string(body)))
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.walletSvc.HandleWebhook(c.Request.Context(), *payload); err != nil {
		response.Fail(c, http.StatusInternalServerError, "WEBHOOK_ERROR", "failed to process webhook")
		return
	}

	response.OK(c, gin.H{"received": true})
}

func handleError(c *gin.Context, err error) {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
