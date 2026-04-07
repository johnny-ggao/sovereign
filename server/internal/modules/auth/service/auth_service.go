package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/sovereign-fund/sovereign/internal/modules/auth/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/auth/model"
	"github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	notifsvc "github.com/sovereign-fund/sovereign/internal/modules/notification/service"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"github.com/sovereign-fund/sovereign/pkg/crypto"
	jwtpkg "github.com/sovereign-fund/sovereign/pkg/jwt"
	"gorm.io/gorm"
)

type AuthService interface {
	SendRegisterOTP(ctx context.Context, req dto.SendRegisterOTPRequest) error
	Register(ctx context.Context, req dto.RegisterRequest, userAgent, clientIP string) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest, userAgent, clientIP string) (*dto.AuthResponse, error)
	Verify2FA(ctx context.Context, userID, code, userAgent, clientIP string) (*dto.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken, userAgent, clientIP string) (*dto.AuthResponse, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error
	GoogleLogin(ctx context.Context, idToken, userAgent, clientIP string) (*dto.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	GetProfile(ctx context.Context, userID string) (*dto.UserProfile, error)
}

type GoogleTokenVerifier interface {
	Verify(ctx context.Context, idToken string) (*GoogleClaims, error)
}

type GoogleClaims struct {
	Sub       string
	Email     string
	Name      string
	Picture   string
	EmailVerified bool
}

type authService struct {
	userRepo       repository.UserRepository
	tokenRepo      repository.TokenRepository
	jwtMgr         *jwtpkg.Manager
	otpSvc         OTPService
	googleVerifier GoogleTokenVerifier
	eventBus       events.Bus
	notifSvc       notifsvc.NotificationService
	logger         *slog.Logger
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	jwtMgr *jwtpkg.Manager,
	otpSvc OTPService,
	googleVerifier GoogleTokenVerifier,
	eventBus events.Bus,
	notifSvc notifsvc.NotificationService,
	logger *slog.Logger,
) AuthService {
	return &authService{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		jwtMgr:         jwtMgr,
		otpSvc:         otpSvc,
		googleVerifier: googleVerifier,
		eventBus:       eventBus,
		notifSvc:       notifSvc,
		logger:         logger,
	}
}

func (s *authService) SendRegisterOTP(ctx context.Context, req dto.SendRegisterOTPRequest) error {
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	if exists {
		return apperr.ErrAccountExists
	}

	otp, err := s.otpSvc.Generate(ctx, req.Email, "register")
	if err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	if err := s.notifSvc.SendOTP(ctx, req.Email, "", "register_otp", otp, "5 minutes"); err != nil {
		s.logger.Error("failed to send register OTP email", slog.String("email", req.Email), slog.String("error", err.Error()))
	}

	return nil
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest, userAgent, clientIP string) (*dto.AuthResponse, error) {
	valid, err := s.otpSvc.Verify(ctx, req.Email, "register", req.Code)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}
	if !valid {
		return nil, apperr.ErrInvalidOTP
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}
	if exists {
		return nil, apperr.ErrAccountExists
	}

	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, fmt.Errorf("hash password: %w", err))
	}

	lang := req.Language
	if lang == "" {
		lang = "ko"
	}

	user := &model.User{
		Email:        req.Email,
		PasswordHash: hash,
		FullName:     req.FullName,
		Phone:        req.Phone,
		Language:     lang,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, fmt.Errorf("create user: %w", err))
	}

	s.eventBus.Publish(ctx, events.Event{
		Type:    events.UserRegistered,
		Payload: map[string]string{"user_id": user.ID, "email": user.Email},
	})

	s.logger.Info("user registered", slog.String("user_id", user.ID), slog.String("email", user.Email))

	return s.issueTokens(ctx, user, userAgent, clientIP)
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest, userAgent, clientIP string) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrInvalidCredentials
		}
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	valid, err := crypto.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		return nil, apperr.ErrInvalidCredentials
	}

	if user.TwoFAEnabled {
		return &dto.AuthResponse{
			Requires2FA: true,
			User:        toUserProfile(user),
		}, nil
	}

	return s.issueTokens(ctx, user, userAgent, clientIP)
}

func (s *authService) Verify2FA(ctx context.Context, userID, code, userAgent, clientIP string) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	if !user.TwoFAEnabled || user.TwoFASecret == "" {
		return nil, apperr.ErrInvalid2FA
	}

	valid, err := crypto.VerifyTOTP(user.TwoFASecret, code)
	if err != nil || !valid {
		return nil, apperr.ErrInvalid2FA
	}

	return s.issueTokens(ctx, user, userAgent, clientIP)
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken, userAgent, clientIP string) (*dto.AuthResponse, error) {
	claims, err := s.jwtMgr.ValidateRefresh(refreshToken)
	if err != nil {
		return nil, apperr.ErrInvalidToken
	}

	storedToken, err := s.tokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrInvalidToken
		}
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	if storedToken.IsExpired() {
		s.tokenRepo.DeleteByToken(ctx, refreshToken)
		return nil, apperr.ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	if err := s.tokenRepo.DeleteByToken(ctx, refreshToken); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	return s.issueTokens(ctx, user, userAgent, clientIP)
}

func (s *authService) GoogleLogin(ctx context.Context, idToken, userAgent, clientIP string) (*dto.AuthResponse, error) {
	claims, err := s.googleVerifier.Verify(ctx, idToken)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrUnauthorized, fmt.Errorf("verify google token: %w", err))
	}

	if !claims.EmailVerified {
		return nil, apperr.New(400, "GOOGLE_EMAIL_NOT_VERIFIED", "Google email is not verified")
	}

	// Try to find existing user by Google ID
	user, err := s.userRepo.FindByGoogleID(ctx, claims.Sub)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	if user != nil {
		// Existing Google user — login
		s.logger.Info("google login", slog.String("user_id", user.ID), slog.String("email", user.Email))
		return s.issueTokens(ctx, user, userAgent, clientIP)
	}

	// Try to find by email (link existing account)
	user, err = s.userRepo.FindByEmail(ctx, claims.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	if user != nil {
		// Link Google to existing account
		updated := *user
		updated.GoogleID = claims.Sub
		if updated.AvatarURL == "" && claims.Picture != "" {
			updated.AvatarURL = claims.Picture
		}
		if err := s.userRepo.Update(ctx, &updated); err != nil {
			return nil, apperr.Wrap(apperr.ErrInternal, err)
		}
		s.logger.Info("google linked to existing account", slog.String("user_id", user.ID))
		return s.issueTokens(ctx, &updated, userAgent, clientIP)
	}

	// New user — auto register
	newUser := &model.User{
		Email:     claims.Email,
		GoogleID:  claims.Sub,
		FullName:  claims.Name,
		AvatarURL: claims.Picture,
		Language:  "ko",
	}
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, fmt.Errorf("create google user: %w", err))
	}

	s.eventBus.Publish(ctx, events.Event{
		Type:    events.UserRegistered,
		Payload: map[string]string{"user_id": newUser.ID, "email": newUser.Email, "provider": "google"},
	})

	s.logger.Info("google user registered", slog.String("user_id", newUser.ID), slog.String("email", newUser.Email))
	return s.issueTokens(ctx, newUser, userAgent, clientIP)
}

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	otp, err := s.otpSvc.Generate(ctx, email, "reset_password")
	if err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	lang := user.Language
	if lang == "" {
		lang = "en"
	}
	if err := s.notifSvc.SendOTP(ctx, email, lang, "password_reset", otp, "5 minutes"); err != nil {
		s.logger.Error("failed to send password reset email", slog.String("email", email), slog.String("error", err.Error()))
	}

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error {
	// 验证 OTP
	valid, err := s.otpSvc.Verify(ctx, req.Email, "reset_password", req.Code)
	if err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	if !valid {
		return apperr.New(400, "INVALID_CODE", "invalid or expired verification code")
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	hash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	updated := *user
	updated.PasswordHash = hash
	if err := s.userRepo.Update(ctx, &updated); err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	s.logger.Info("password reset", slog.String("email", req.Email))
	return nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	if err := s.tokenRepo.DeleteByToken(ctx, refreshToken); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	return nil
}

func (s *authService) GetProfile(ctx context.Context, userID string) (*dto.UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}
	return toUserProfile(user), nil
}

func (s *authService) issueTokens(ctx context.Context, user *model.User, userAgent, clientIP string) (*dto.AuthResponse, error) {
	tokenPair, err := s.jwtMgr.Generate(user.ID, user.Email)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, fmt.Errorf("generate tokens: %w", err))
	}

	rt := &model.RefreshToken{
		UserID:    user.ID,
		Token:     tokenPair.RefreshToken,
		UserAgent: userAgent,
		ClientIP:  clientIP,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.tokenRepo.Create(ctx, rt); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, fmt.Errorf("store refresh token: %w", err))
	}

	s.eventBus.Publish(ctx, events.Event{
		Type:    events.UserLoggedIn,
		Payload: map[string]string{"user_id": user.ID},
	})

	return &dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		User:         toUserProfile(user),
	}, nil
}

func toUserProfile(user *model.User) *dto.UserProfile {
	return &dto.UserProfile{
		ID:           user.ID,
		Email:        user.Email,
		FullName:     user.FullName,
		AvatarURL:    user.AvatarURL,
		Phone:        user.Phone,
		Language:     user.Language,
		KYCStatus:    user.KYCStatus,
		TwoFAEnabled: user.TwoFAEnabled,
	}
}
