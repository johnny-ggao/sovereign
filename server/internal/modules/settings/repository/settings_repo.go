package repository

import (
	"context"

	"github.com/sovereign-fund/sovereign/internal/modules/settings/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SettingsRepository interface {
	FindNotificationPref(ctx context.Context, userID string) (*model.NotificationPref, error)
	UpsertNotificationPref(ctx context.Context, pref *model.NotificationPref) error

	FindLoginDevices(ctx context.Context, userID string) ([]model.LoginDevice, error)
	UpsertLoginDevice(ctx context.Context, device *model.LoginDevice) error
	DeleteLoginDevice(ctx context.Context, id, userID string) error
}

type settingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) SettingsRepository {
	return &settingsRepository{db: db}
}

func (r *settingsRepository) FindNotificationPref(ctx context.Context, userID string) (*model.NotificationPref, error) {
	var pref model.NotificationPref
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&pref).Error
	if err != nil {
		return nil, err
	}
	return &pref, nil
}

func (r *settingsRepository) UpsertNotificationPref(ctx context.Context, pref *model.NotificationPref) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"email_trade", "email_deposit", "email_withdraw", "email_settlement",
				"push_premium_alert", "push_trade", "push_deposit", "push_withdraw",
				"premium_threshold", "updated_at",
			}),
		}).
		Create(pref).Error
}

func (r *settingsRepository) FindLoginDevices(ctx context.Context, userID string) ([]model.LoginDevice, error) {
	var devices []model.LoginDevice
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("last_login DESC").
		Limit(20).
		Find(&devices).Error
	return devices, err
}

func (r *settingsRepository) UpsertLoginDevice(ctx context.Context, device *model.LoginDevice) error {
	return r.db.WithContext(ctx).Create(device).Error
}

func (r *settingsRepository) DeleteLoginDevice(ctx context.Context, id, userID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.LoginDevice{}).Error
}
