package dto

type UpdateProfileRequest struct {
	FullName string `json:"full_name" binding:"omitempty,max=255"`
	Phone    string `json:"phone" binding:"omitempty,max=50"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=128"`
}

type Setup2FARequest struct{}

type Verify2FASetupRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

type Disable2FARequest struct {
	Code     string `json:"code" binding:"required,len=6"`
	Password string `json:"password" binding:"required"`
}

type UpdateNotificationRequest struct {
	EmailTrade       *bool    `json:"email_trade"`
	EmailDeposit     *bool    `json:"email_deposit"`
	EmailWithdraw    *bool    `json:"email_withdraw"`
	EmailSettlement  *bool    `json:"email_settlement"`
	PushPremiumAlert *bool    `json:"push_premium_alert"`
	PushTrade        *bool    `json:"push_trade"`
	PushDeposit      *bool    `json:"push_deposit"`
	PushWithdraw     *bool    `json:"push_withdraw"`
	PremiumThreshold *float64 `json:"premium_threshold" binding:"omitempty,min=0,max=100"`
}

type UpdateLanguageRequest struct {
	Language string `json:"language" binding:"required,oneof=ko en zh"`
}
