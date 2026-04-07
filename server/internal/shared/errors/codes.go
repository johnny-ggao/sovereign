package apperr

import "net/http"

// Authentication
var (
	ErrUnauthorized       = New(http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "authentication required")
	ErrInvalidCredentials = New(http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "invalid email or password")
	ErrInvalid2FA         = New(http.StatusUnauthorized, "AUTH_INVALID_2FA", "invalid 2FA code")
	ErrInvalidOTP         = New(http.StatusUnauthorized, "AUTH_INVALID_OTP", "invalid or expired OTP")
	ErrInvalidToken       = New(http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "invalid or expired token")
	ErrForbidden          = New(http.StatusForbidden, "AUTH_FORBIDDEN", "insufficient permissions")
	ErrAccountExists      = New(http.StatusConflict, "AUTH_ACCOUNT_EXISTS", "account already exists")
	ErrAccountNotFound    = New(http.StatusNotFound, "AUTH_ACCOUNT_NOT_FOUND", "account not found")
)

// Wallet
var (
	ErrInsufficientFunds = New(http.StatusUnprocessableEntity, "WALLET_INSUFFICIENT_FUNDS", "insufficient balance")
	ErrAddressCooldown   = New(http.StatusTooManyRequests, "WALLET_ADDRESS_COOLDOWN", "new address requires 24h cooldown")
	ErrAddressNotWhitelisted = New(http.StatusForbidden, "WALLET_ADDRESS_NOT_WHITELISTED", "address not in whitelist")
	ErrWithdrawLimitExceeded = New(http.StatusUnprocessableEntity, "WALLET_WITHDRAW_LIMIT", "withdrawal limit exceeded")
)

// Investment
var (
	ErrInvestmentNotFound = New(http.StatusNotFound, "INVESTMENT_NOT_FOUND", "investment not found")
	ErrMinInvestment      = New(http.StatusUnprocessableEntity, "INVESTMENT_MIN_AMOUNT", "below minimum investment amount")
	ErrRedeemPending      = New(http.StatusConflict, "INVESTMENT_REDEEM_PENDING", "redemption already pending")
)

// General
var (
	ErrNotFound       = New(http.StatusNotFound, "RESOURCE_NOT_FOUND", "resource not found")
	ErrBadRequest     = New(http.StatusBadRequest, "BAD_REQUEST", "invalid request")
	ErrInternal       = New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	ErrRateLimited    = New(http.StatusTooManyRequests, "RATE_LIMITED", "too many requests")
	ErrValidation     = New(http.StatusBadRequest, "VALIDATION_ERROR", "validation failed")
)
