package dto

type AuthResponse struct {
	AccessToken  string       `json:"access_token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	ExpiresAt    int64        `json:"expires_at,omitempty"`
	Requires2FA  bool         `json:"requires_2fa,omitempty"`
	User         *UserProfile `json:"user,omitempty"`
}

type UserProfile struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	FullName     string `json:"full_name"`
	AvatarURL    string `json:"avatar_url"`
	Phone        string `json:"phone"`
	Language     string `json:"language"`
	KYCStatus    string `json:"kyc_status"`
	TwoFAEnabled bool   `json:"two_fa_enabled"`
}

type Setup2FAResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
