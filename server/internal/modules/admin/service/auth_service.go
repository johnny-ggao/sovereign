package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	adminmw "github.com/sovereign-fund/sovereign/internal/modules/admin/middleware"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/repository"
	"github.com/sovereign-fund/sovereign/pkg/crypto"
	"gorm.io/gorm"
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	ChangePassword(ctx context.Context, adminID string, req dto.ChangePasswordRequest) error
}

type authService struct {
	repo      repository.AdminRepository
	jwtSecret string
	jwtExpiry time.Duration
	logger    *slog.Logger
}

func NewAuthService(repo repository.AdminRepository, jwtSecret string, jwtExpiry time.Duration, logger *slog.Logger) AuthService {
	return &authService{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
		logger:    logger,
	}
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	admin, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("find admin: %w", err)
	}

	if !admin.IsActive {
		return nil, fmt.Errorf("account is disabled")
	}

	match, err := crypto.VerifyPassword(req.Password, admin.PasswordHash)
	if err != nil || !match {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, expiresAt, err := adminmw.GenerateAdminToken(s.jwtSecret, s.jwtExpiry, admin.ID, admin.Email, admin.Role)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	_ = s.repo.UpdateLastLogin(ctx, admin.ID)

	s.logger.Info("admin login", slog.String("admin_id", admin.ID), slog.String("email", admin.Email))

	return &dto.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		Admin:     toAdminResponse(admin),
	}, nil
}

func (s *authService) ChangePassword(ctx context.Context, adminID string, req dto.ChangePasswordRequest) error {
	admin, err := s.repo.FindByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("find admin: %w", err)
	}

	match, err := crypto.VerifyPassword(req.OldPassword, admin.PasswordHash)
	if err != nil || !match {
		return fmt.Errorf("invalid old password")
	}

	hash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	updated := *admin
	updated.PasswordHash = hash
	if err := s.repo.Update(ctx, &updated); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}

func toAdminResponse(a *model.AdminUser) dto.AdminResponse {
	return dto.AdminResponse{
		ID:        a.ID,
		Email:     a.Email,
		Name:      a.Name,
		Role:      a.Role,
		IsActive:  a.IsActive,
		LastLogin: a.LastLogin,
		CreatedAt: a.CreatedAt,
	}
}
