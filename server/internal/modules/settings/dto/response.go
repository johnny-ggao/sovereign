package dto

type ProfileResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	FullName     string `json:"full_name"`
	Phone        string `json:"phone"`
	Language     string `json:"language"`
	KYCStatus    string `json:"kyc_status"`
	TwoFAEnabled bool   `json:"two_fa_enabled"`
	CreatedAt    string `json:"created_at"`
}

type Setup2FAResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
}

type SecurityOverview struct {
	TwoFAEnabled bool            `json:"two_fa_enabled"`
	Devices      []DeviceResponse `json:"devices"`
}

type DeviceResponse struct {
	ID        string `json:"id"`
	UserAgent string `json:"user_agent"`
	IP        string `json:"ip"`
	Location  string `json:"location"`
	LastLogin string `json:"last_login"`
}

type NotificationPrefResponse struct {
	EmailTrade       bool    `json:"email_trade"`
	EmailDeposit     bool    `json:"email_deposit"`
	EmailWithdraw    bool    `json:"email_withdraw"`
	EmailSettlement  bool    `json:"email_settlement"`
	PushPremiumAlert bool    `json:"push_premium_alert"`
	PushTrade        bool    `json:"push_trade"`
	PushDeposit      bool    `json:"push_deposit"`
	PushWithdraw     bool    `json:"push_withdraw"`
	PremiumThreshold float64 `json:"premium_threshold"`
}

type KYCStatusResponse struct {
	Status    string `json:"status"`
	SubmitURL string `json:"submit_url,omitempty"`
	Message   string `json:"message,omitempty"`
}
