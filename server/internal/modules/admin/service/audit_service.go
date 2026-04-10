package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	"gorm.io/gorm"
)

type AuditListQuery struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
	Action   string `form:"action"`
	AdminID  string `form:"admin_id"`
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
}

type AuditService interface {
	Log(ctx context.Context, adminID, adminEmail, action, targetType, targetID, detail, ipAddress string) error
	List(ctx context.Context, query AuditListQuery) ([]model.AuditLog, int64, error)
}

type auditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) AuditService {
	return &auditService{db: db}
}

func (s *auditService) Log(ctx context.Context, adminID, adminEmail, action, targetType, targetID, detail, ipAddress string) error {
	entry := &model.AuditLog{
		AdminID:    adminID,
		AdminEmail: adminEmail,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detail,
		IPAddress:  ipAddress,
	}

	if err := s.db.WithContext(ctx).Create(entry).Error; err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}

	return nil
}

func (s *auditService) List(ctx context.Context, query AuditListQuery) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	db := s.db.WithContext(ctx).Model(&model.AuditLog{})

	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.AdminID != "" {
		db = db.Where("admin_id = ?", query.AdminID)
	}
	if query.DateFrom != "" {
		db = db.Where("created_at >= ?", strings.TrimSpace(query.DateFrom))
	}
	if query.DateTo != "" {
		db = db.Where("created_at <= ?", strings.TrimSpace(query.DateTo)+" 23:59:59")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}

	offset := (query.Page - 1) * query.Limit
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.Limit).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("find audit logs: %w", err)
	}

	return logs, total, nil
}
