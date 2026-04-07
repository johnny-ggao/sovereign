package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	authModel "github.com/sovereign-fund/sovereign/internal/modules/auth/model"
	authRepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/settings/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/settings/model"
	"github.com/sovereign-fund/sovereign/internal/modules/settings/repository"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/pkg/crypto"
	"gorm.io/gorm"
)

type SettingsService interface {
	GetProfile(ctx context.Context, userID string) (*dto.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error)
	ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) error

	Setup2FA(ctx context.Context, userID string) (*dto.Setup2FAResponse, error)
	Verify2FASetup(ctx context.Context, userID string, req dto.Verify2FASetupRequest) error
	Disable2FA(ctx context.Context, userID string, req dto.Disable2FARequest) error
	GetSecurityOverview(ctx context.Context, userID string) (*dto.SecurityOverview, error)
	RevokeDevice(ctx context.Context, userID, deviceID string) error

	GetNotificationPref(ctx context.Context, userID string) (*dto.NotificationPrefResponse, error)
	UpdateNotificationPref(ctx context.Context, userID string, req dto.UpdateNotificationRequest) (*dto.NotificationPrefResponse, error)

	UpdateLanguage(ctx context.Context, userID string, req dto.UpdateLanguageRequest) error

}

type settingsService struct {
	userRepo     authRepo.UserRepository
	settingsRepo repository.SettingsRepository
	logger       *slog.Logger
}

func NewSettingsService(
	ur authRepo.UserRepository,
	sr repository.SettingsRepository,
	logger *slog.Logger,
) SettingsService {
	return &settingsService{
		userRepo:     ur,
		settingsRepo: sr,
		logger:       logger,
	}
}

func (s *settingsService) GetProfile(ctx context.Context, userID string) (*dto.ProfileResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}
	return toProfileResponse(user), nil
}

func (s *settingsService) UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	updated := *user
	if req.FullName != "" {
		updated.FullName = req.FullName
	}
	if req.Phone != "" {
		updated.Phone = req.Phone
	}

	if err := s.userRepo.Update(ctx, &updated); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	return toProfileResponse(&updated), nil
}

func (s *settingsService) ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	valid, err := crypto.VerifyPassword(req.CurrentPassword, user.PasswordHash)
	if err != nil || !valid {
		return apperr.ErrInvalidCredentials
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

	s.logger.Info("password changed", slog.String("user_id", userID))
	return nil
}

func (s *settingsService) Setup2FA(ctx context.Context, userID string) (*dto.Setup2FAResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	if user.TwoFAEnabled {
		return nil, apperr.New(409, "2FA_ALREADY_ENABLED", "2FA is already enabled")
	}

	secret, qrURL, err := crypto.GenerateTOTPSecret("Sovereign", user.Email)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	updated := *user
	updated.TwoFASecret = secret
	if err := s.userRepo.Update(ctx, &updated); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	return &dto.Setup2FAResponse{
		Secret:    secret,
		QRCodeURL: qrURL,
	}, nil
}

func (s *settingsService) Verify2FASetup(ctx context.Context, userID string, req dto.Verify2FASetupRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	if user.TwoFASecret == "" {
		return apperr.New(400, "2FA_NOT_SETUP", "call setup 2FA first")
	}

	valid, err := crypto.VerifyTOTP(user.TwoFASecret, req.Code)
	if err != nil || !valid {
		return apperr.ErrInvalid2FA
	}

	updated := *user
	updated.TwoFAEnabled = true
	if err := s.userRepo.Update(ctx, &updated); err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	s.logger.Info("2FA enabled", slog.String("user_id", userID))
	return nil
}

func (s *settingsService) Disable2FA(ctx context.Context, userID string, req dto.Disable2FARequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	valid, err := crypto.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		return apperr.ErrInvalidCredentials
	}

	validCode, err := crypto.VerifyTOTP(user.TwoFASecret, req.Code)
	if err != nil || !validCode {
		return apperr.ErrInvalid2FA
	}

	updated := *user
	updated.TwoFAEnabled = false
	updated.TwoFASecret = ""
	if err := s.userRepo.Update(ctx, &updated); err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	s.logger.Info("2FA disabled", slog.String("user_id", userID))
	return nil
}

func (s *settingsService) GetSecurityOverview(ctx context.Context, userID string) (*dto.SecurityOverview, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	devices, err := s.settingsRepo.FindLoginDevices(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	devResp := make([]dto.DeviceResponse, 0, len(devices))
	for _, d := range devices {
		devResp = append(devResp, dto.DeviceResponse{
			ID:        d.ID,
			UserAgent: d.UserAgent,
			IP:        d.IP,
			Location:  d.Location,
			LastLogin: d.LastLogin.Format(time.RFC3339),
		})
	}

	return &dto.SecurityOverview{
		TwoFAEnabled: user.TwoFAEnabled,
		Devices:      devResp,
	}, nil
}

func (s *settingsService) RevokeDevice(ctx context.Context, userID, deviceID string) error {
	if err := s.settingsRepo.DeleteLoginDevice(ctx, deviceID, userID); err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	return nil
}

func (s *settingsService) GetNotificationPref(ctx context.Context, userID string) (*dto.NotificationPrefResponse, error) {
	pref, err := s.settingsRepo.FindNotificationPref(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return defaultNotificationPref(), nil
		}
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}
	return toNotifPrefResponse(pref), nil
}

func (s *settingsService) UpdateNotificationPref(ctx context.Context, userID string, req dto.UpdateNotificationRequest) (*dto.NotificationPrefResponse, error) {
	pref, err := s.settingsRepo.FindNotificationPref(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	if pref == nil {
		pref = &model.NotificationPref{UserID: userID}
	}

	updated := *pref
	if req.EmailTrade != nil {
		updated.EmailTrade = *req.EmailTrade
	}
	if req.EmailDeposit != nil {
		updated.EmailDeposit = *req.EmailDeposit
	}
	if req.EmailWithdraw != nil {
		updated.EmailWithdraw = *req.EmailWithdraw
	}
	if req.EmailSettlement != nil {
		updated.EmailSettlement = *req.EmailSettlement
	}
	if req.PushPremiumAlert != nil {
		updated.PushPremiumAlert = *req.PushPremiumAlert
	}
	if req.PushTrade != nil {
		updated.PushTrade = *req.PushTrade
	}
	if req.PushDeposit != nil {
		updated.PushDeposit = *req.PushDeposit
	}
	if req.PushWithdraw != nil {
		updated.PushWithdraw = *req.PushWithdraw
	}
	if req.PremiumThreshold != nil {
		updated.PremiumThreshold = *req.PremiumThreshold
	}

	if err := s.settingsRepo.UpsertNotificationPref(ctx, &updated); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	return toNotifPrefResponse(&updated), nil
}

func (s *settingsService) UpdateLanguage(ctx context.Context, userID string, req dto.UpdateLanguageRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	updated := *user
	updated.Language = req.Language
	if err := s.userRepo.Update(ctx, &updated); err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	return nil
}

func toProfileResponse(u *authModel.User) *dto.ProfileResponse {
	return &dto.ProfileResponse{
		ID:           u.ID,
		Email:        u.Email,
		FullName:     u.FullName,
		Phone:        u.Phone,
		Language:     u.Language,
		KYCStatus:    u.KYCStatus,
		TwoFAEnabled: u.TwoFAEnabled,
		CreatedAt:    u.CreatedAt.Format(time.RFC3339),
	}
}

func toNotifPrefResponse(p *model.NotificationPref) *dto.NotificationPrefResponse {
	return &dto.NotificationPrefResponse{
		EmailTrade:       p.EmailTrade,
		EmailDeposit:     p.EmailDeposit,
		EmailWithdraw:    p.EmailWithdraw,
		EmailSettlement:  p.EmailSettlement,
		PushPremiumAlert: p.PushPremiumAlert,
		PushTrade:        p.PushTrade,
		PushDeposit:      p.PushDeposit,
		PushWithdraw:     p.PushWithdraw,
		PremiumThreshold: p.PremiumThreshold,
	}
}

func defaultNotificationPref() *dto.NotificationPrefResponse {
	return &dto.NotificationPrefResponse{
		EmailTrade:       true,
		EmailDeposit:     true,
		EmailWithdraw:    true,
		EmailSettlement:  true,
		PushPremiumAlert: false,
		PushTrade:        true,
		PushDeposit:      true,
		PushWithdraw:     true,
		PremiumThreshold: 3.0,
	}
}
