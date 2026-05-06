package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Transaction struct {
	ID         string          `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     string          `gorm:"type:uuid;index;not null" json:"user_id"`
	Type       string          `gorm:"type:varchar(20);not null" json:"type"`
	Currency   string          `gorm:"type:varchar(10);not null" json:"currency"`
	Network    string          `gorm:"type:varchar(20)" json:"network"`
	Amount     decimal.Decimal `gorm:"type:decimal(28,18);not null" json:"amount"`
	Fee        decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"fee"`
	Address    string          `gorm:"type:varchar(255)" json:"address"`
	TxHash     string          `gorm:"type:varchar(255)" json:"tx_hash"`
	Status     string          `gorm:"type:varchar(20);default:pending" json:"status"`
	ExternalID string          `gorm:"type:varchar(255)" json:"external_id"`

	// Review-layer fields. Populated only for type='withdraw'.
	ReviewStatus    string     `gorm:"type:varchar(32);index" json:"review_status,omitempty"`
	ReviewedBy      *string    `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	RejectReason    string     `gorm:"type:varchar(500)" json:"reject_reason,omitempty"`
	SubmitAttempts  int        `gorm:"not null;default:0" json:"submit_attempts"`
	LastSubmitError string     `gorm:"type:text" json:"last_submit_error,omitempty"`
	LastSubmitAt    *time.Time `json:"last_submit_at,omitempty"`

	ConfirmedAt *time.Time `json:"confirmed_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (t *Transaction) BeforeCreate(_ *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

const (
	TxTypeDeposit  = "deposit"
	TxTypeWithdraw = "withdraw"

	TxStatusPending    = "pending"
	TxStatusProcessing = "processing"
	TxStatusConfirmed  = "confirmed"
	TxStatusFailed     = "failed"
	TxStatusCancelled  = "cancelled"

	// Review states (only meaningful for type='withdraw').
	ReviewStatusPendingReview = "pending_review"
	ReviewStatusSubmitted     = "submitted"
	ReviewStatusSubmitFailed  = "submit_failed"
	ReviewStatusRejected      = "rejected"
	ReviewStatusCancelled     = "cancelled"
)
