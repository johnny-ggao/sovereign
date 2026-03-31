package dto

type SendRegisterOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required,len=6"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	FullName string `json:"full_name" binding:"required,min=1,max=255"`
	Phone    string `json:"phone" binding:"omitempty,max=50"`
	Language string `json:"language" binding:"omitempty,oneof=ko en zh"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type Verify2FARequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=128"`
}

type VerifyOTPRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Code    string `json:"code" binding:"required,len=6"`
	Purpose string `json:"purpose" binding:"required,oneof=register reset_password"`
}

type GoogleLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type Setup2FARequest struct{}
