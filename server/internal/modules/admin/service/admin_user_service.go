package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/repository"
	"github.com/sovereign-fund/sovereign/pkg/crypto"
)

type AdminUserService interface {
	List(ctx context.Context) ([]dto.AdminResponse, error)
	Create(ctx context.Context, req dto.CreateAdminRequest) (*dto.AdminResponse, error)
	Update(ctx context.Context, id string, req dto.UpdateAdminRequest) (*dto.AdminResponse, error)
	Delete(ctx context.Context, id, currentAdminID string) error
}

type adminUserService struct {
	repo   repository.AdminRepository
	logger *slog.Logger
}

func NewAdminUserService(repo repository.AdminRepository, logger *slog.Logger) AdminUserService {
	return &adminUserService{repo: repo, logger: logger}
}

func (s *adminUserService) List(ctx context.Context) ([]dto.AdminResponse, error) {
	admins, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("find admins: %w", err)
	}

	result := make([]dto.AdminResponse, len(admins))
	for i, a := range admins {
		result[i] = toAdminResponse(&a)
	}
	return result, nil
}

func (s *adminUserService) Create(ctx context.Context, req dto.CreateAdminRequest) (*dto.AdminResponse, error) {
	if !model.IsValidRole(req.Role) {
		return nil, fmt.Errorf("invalid role: %s", req.Role)
	}

	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	admin := &model.AdminUser{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Role:         req.Role,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, admin); err != nil {
		return nil, fmt.Errorf("create admin: %w", err)
	}

	s.logger.Info("admin created", slog.String("admin_id", admin.ID), slog.String("email", admin.Email), slog.String("role", admin.Role))
	resp := toAdminResponse(admin)
	return &resp, nil
}

func (s *adminUserService) Update(ctx context.Context, id string, req dto.UpdateAdminRequest) (*dto.AdminResponse, error) {
	admin, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find admin: %w", err)
	}

	updated := *admin
	if req.Name != "" {
		updated.Name = req.Name
	}
	if req.Role != "" {
		if !model.IsValidRole(req.Role) {
			return nil, fmt.Errorf("invalid role: %s", req.Role)
		}
		updated.Role = req.Role
	}
	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, &updated); err != nil {
		return nil, fmt.Errorf("update admin: %w", err)
	}

	resp := toAdminResponse(&updated)
	return &resp, nil
}

func (s *adminUserService) Delete(ctx context.Context, id, currentAdminID string) error {
	if id == currentAdminID {
		return fmt.Errorf("cannot delete yourself")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete admin: %w", err)
	}

	s.logger.Info("admin deleted", slog.String("admin_id", id))
	return nil
}
