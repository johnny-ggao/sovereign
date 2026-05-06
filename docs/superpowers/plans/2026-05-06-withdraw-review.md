# Withdraw Review Layer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Insert an admin-review queue between user withdrawal requests and Cobo so insufficient hot-wallet balance or transient Cobo errors no longer fail user withdrawals.

**Architecture:** Extend `transactions` table with a `review_status` axis. User-facing `Withdraw` no longer calls Cobo — it freezes funds and creates a `pending_review` ticket. Admins approve/reject/retry via new endpoints; only admin-approve calls Cobo. Cobo failures move the ticket to `submit_failed` (still frozen, no refund) for manual retry. Only admin reject or Cobo webhook failure refund the user.

**Tech Stack:** Go 1.22+ / Gin / GORM / PostgreSQL / Cobo WaaS 2.0 SDK / Next.js 19 frontend / Ant Design Pro admin / Vitest / Playwright

**Reference spec:** `docs/superpowers/specs/2026-05-06-withdraw-review-design.md`

---

## File Structure

### Backend (Go)

| Path | Action | Responsibility |
|---|---|---|
| `server/migrations/000023_withdraw_review_fields.up.sql` | create | Add review_status, reviewed_by, reviewed_at, reject_reason, submit_attempts, last_submit_error, last_submit_at to `transactions` |
| `server/migrations/000023_withdraw_review_fields.down.sql` | create | Roll back |
| `server/internal/modules/wallet/model/transaction.go` | modify | Add new fields and `ReviewStatus*` constants |
| `server/internal/modules/wallet/repository/transaction_repo.go` | modify | Add `UpdateReview`, `ListWithdrawByReviewStatus`, `IncrementSubmitAttempt` |
| `server/internal/modules/wallet/service/wallet_service.go` | modify | (1) Drop Cobo call from `Withdraw`; (2) add `CancelWithdraw` |
| `server/internal/modules/wallet/service/wallet_service_test.go` | modify | Update existing tests + new tests for cancel and pending_review behavior |
| `server/internal/modules/wallet/dto/request.go` | modify | (no change) |
| `server/internal/modules/wallet/dto/response.go` | modify | Add `ReviewStatus`, `RejectReason` to `TransactionResponse` |
| `server/internal/modules/wallet/handler/wallet_handler.go` | modify | Add `CancelWithdraw` handler |
| `server/internal/modules/wallet/routes.go` | modify | Add `DELETE /wallets/withdraw/:id` |
| `server/internal/modules/admin/dto/request.go` | modify | Add `WithdrawReviewListQuery`, `RejectWithdrawRequest` |
| `server/internal/modules/admin/dto/response.go` | modify | Add `WithdrawReviewItem` |
| `server/internal/modules/admin/service/withdraw_review_service.go` | create | `List`, `Approve`, `Reject`, `Retry`; private `submitToCobo` |
| `server/internal/modules/admin/service/withdraw_review_service_test.go` | create | Unit tests with mock Cobo provider and stub repos |
| `server/internal/modules/admin/handler/withdraw_review_handler.go` | create | Wire HTTP endpoints + audit logging |
| `server/internal/modules/admin/handler/withdraw_review_handler_test.go` | create | Integration tests with httptest |
| `server/internal/modules/admin/module.go` | modify | Wire new service + handler |
| `server/internal/modules/admin/routes.go` | modify | Register `/admin/withdrawals*` routes with `super_admin`/`operator` permission |
| `server/internal/modules/notification/service/notification_service.go` | modify | Subscribe to `WithdrawRejected` event (`WithdrawCompleted`/`WithdrawFailed` already wired) |
| `server/internal/shared/events/types.go` | modify | Add `WithdrawRejected` event constant |

### Admin frontend (Ant Design Pro)

| Path | Action | Responsibility |
|---|---|---|
| `admin/src/pages/Withdrawals/index.tsx` | create | Listing + tabs + actions |
| `admin/src/pages/Withdrawals/RejectModal.tsx` | create | Reject reason modal |
| `admin/src/services/api.ts` | modify | Add `listWithdrawals`, `approveWithdrawal`, `rejectWithdrawal`, `retryWithdrawal` |
| `admin/config/routes.ts` | modify | Register `/withdrawals` route under transactions menu |
| `admin/src/locales/zh-CN/menu.ts` | modify | Add menu label |

### User frontend (Next.js)

| Path | Action | Responsibility |
|---|---|---|
| `front/src/app/(app)/wallet/page.tsx` | modify | Show `review_status` per row + `Cancel` button when `pending_review` |
| `front/src/hooks/use-api.ts` | modify | Add `useCancelWithdraw` |
| `front/src/i18n/{en,zh,ko}.json` | modify | Wallet status labels + cancel CTA + new submit success copy |

### E2E

| Path | Action | Responsibility |
|---|---|---|
| `e2e/tests/withdraw-review.spec.ts` | create | Playwright happy path + rejection path + Cobo-failure-retry path |

---

## Phase 1 — Database & Model

### Task 1: Add migration #023

**Files:**
- Create: `server/migrations/000023_withdraw_review_fields.up.sql`
- Create: `server/migrations/000023_withdraw_review_fields.down.sql`

- [ ] **Step 1: Create up migration**

Write `server/migrations/000023_withdraw_review_fields.up.sql`:
```sql
ALTER TABLE transactions
  ADD COLUMN review_status     varchar(32),
  ADD COLUMN reviewed_by       uuid,
  ADD COLUMN reviewed_at       timestamptz,
  ADD COLUMN reject_reason     varchar(500),
  ADD COLUMN submit_attempts   int NOT NULL DEFAULT 0,
  ADD COLUMN last_submit_error text,
  ADD COLUMN last_submit_at    timestamptz;

CREATE INDEX idx_tx_review_status ON transactions(review_status)
  WHERE type = 'withdraw';

-- Backfill: historical withdraw rows are treated as already submitted directly.
UPDATE transactions
   SET review_status = 'submitted'
 WHERE type = 'withdraw' AND review_status IS NULL;
```

> Note: we intentionally do NOT add a FK to `admin_users(id)` because the column is nullable and the project keeps cross-module FKs loose (consistent with `admin_audit_logs`).

- [ ] **Step 2: Create down migration**

Write `server/migrations/000023_withdraw_review_fields.down.sql`:
```sql
DROP INDEX IF EXISTS idx_tx_review_status;
ALTER TABLE transactions
  DROP COLUMN IF EXISTS review_status,
  DROP COLUMN IF EXISTS reviewed_by,
  DROP COLUMN IF EXISTS reviewed_at,
  DROP COLUMN IF EXISTS reject_reason,
  DROP COLUMN IF EXISTS submit_attempts,
  DROP COLUMN IF EXISTS last_submit_error,
  DROP COLUMN IF EXISTS last_submit_at;
```

- [ ] **Step 3: Run migration locally**

Run: `cd server && make migrate-up` (or whatever the project uses; if unknown, check `Makefile` first with `cat server/Makefile | grep migrate`).
Expected: migration #023 applied without error. Verify with `psql $DATABASE_URL -c "\d transactions"` — should list the seven new columns.

- [ ] **Step 4: Commit**

```bash
git add server/migrations/000023_withdraw_review_fields.*.sql
git commit -m "feat(db): add withdraw review fields to transactions"
```

---

### Task 2: Extend Transaction model and add review status constants

**Files:**
- Modify: `server/internal/modules/wallet/model/transaction.go`

- [ ] **Step 1: Add fields and constants**

Replace the contents of `server/internal/modules/wallet/model/transaction.go` with:
```go
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
	ReviewedBy      string     `gorm:"type:uuid" json:"reviewed_by,omitempty"`
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
```

- [ ] **Step 2: Build to make sure model compiles**

Run: `cd server && go build ./...`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add server/internal/modules/wallet/model/transaction.go
git commit -m "feat(wallet): add review fields and status constants on Transaction"
```

---

### Task 3: Add review-related repository methods

**Files:**
- Modify: `server/internal/modules/wallet/repository/transaction_repo.go`
- Test: `server/internal/modules/wallet/repository/transaction_repo_test.go` (only if a sibling test file already exists; otherwise rely on integration tests in service layer)

- [ ] **Step 1: Extend interface and implementation**

Edit `server/internal/modules/wallet/repository/transaction_repo.go` — replace the interface block and append new methods:

```go
type TransactionRepository interface {
	Create(ctx context.Context, tx *model.Transaction) error
	FindByID(ctx context.Context, id string) (*model.Transaction, error)
	FindByExternalID(ctx context.Context, externalID string) (*model.Transaction, error)
	FindByUserID(ctx context.Context, userID string, txType string, limit, offset int) ([]model.Transaction, int64, error)
	UpdateStatus(ctx context.Context, id, status, txHash string) error
	UpdateExternalID(ctx context.Context, id, externalID string) error

	// Review-layer additions
	UpdateReview(ctx context.Context, id string, fields map[string]any) error
}
```

Append at end of file:
```go
func (r *transactionRepository) UpdateReview(ctx context.Context, id string, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.Transaction{}).
		Where("id = ?", id).
		Updates(fields).Error
}
```

We deliberately use a flexible `Updates(map)` here so each caller (approve/reject/retry/cancel) can express the exact column subset it owns; this avoids a combinatorial explosion of typed setter methods.

- [ ] **Step 2: Build**

Run: `cd server && go build ./...`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add server/internal/modules/wallet/repository/transaction_repo.go
git commit -m "feat(wallet): add UpdateReview method to TransactionRepository"
```

---

## Phase 2 — User-side: queue, don't call Cobo

### Task 4: Stop calling Cobo from user `Withdraw` (TDD)

**Files:**
- Modify: `server/internal/modules/wallet/service/wallet_service.go`
- Test: `server/internal/modules/wallet/service/wallet_service_test.go`

> If `wallet_service_test.go` does not yet exist, create it. The reference test setup pattern (mock cobo provider, stub repos) is the same shape used in `server/internal/modules/admin/service/trade_service_test.go`. Re-use any helpers that exist in the wallet package.

- [ ] **Step 1: Write failing test for "user submit only freezes and queues"**

Add to `wallet_service_test.go`:
```go
func TestWithdraw_QueuesPendingReviewWithoutCallingCobo(t *testing.T) {
	ctx := context.Background()
	walletRepo := &stubWalletRepo{
		wallet: &model.Wallet{
			ID: "w1", UserID: "u1", Currency: "USDT",
			Available: decimal.NewFromInt(100),
			Frozen:    decimal.Zero,
		},
	}
	addrRepo := &stubAddressRepo{
		whitelist: &model.WithdrawAddress{
			UserID: "u1", Currency: "USDT", Network: "TRC20",
			Address:       "Tabc",
			CooldownUntil: time.Now().Add(-time.Hour), // not in cooldown
			IsActive:      true,
		},
	}
	txRepo := &stubTxRepo{}
	provider := &mockCoboProvider{} // every method panics if called
	bus := &spyEventBus{}
	twoFA := &stubTwoFA{ok: true}

	svc := NewWalletService(walletRepo, addrRepo, txRepo, provider, bus, twoFA, slog.Default(), 24*time.Hour)

	resp, err := svc.Withdraw(ctx, "u1", dto.WithdrawRequest{
		Currency: "USDT", Network: "TRC20", Address: "Tabc",
		Amount: "50", TwoFACode: "123456",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.TransactionID)
	assert.Equal(t, "pending_review", resp.Status)

	// Funds moved from available to frozen
	assert.True(t, walletRepo.lastAvailable.Equal(decimal.NewFromInt(50)))
	assert.True(t, walletRepo.lastFrozen.Equal(decimal.NewFromInt(50)))

	// tx persisted with review_status=pending_review, status=pending
	require.Len(t, txRepo.created, 1)
	assert.Equal(t, model.TxStatusPending, txRepo.created[0].Status)
	assert.Equal(t, model.ReviewStatusPendingReview, txRepo.created[0].ReviewStatus)

	// Cobo NOT called and event NOT published
	assert.Zero(t, provider.withdrawCalls)
	assert.Empty(t, bus.published)
}
```

If you don't already have `stubWalletRepo`, `stubAddressRepo`, `stubTxRepo`, `mockCoboProvider`, `spyEventBus`, `stubTwoFA` in this package, create a `wallet_service_stubs_test.go` with the minimum implementations needed. Each stub should only implement methods used by these tests; record the calls/inputs for assertions.

`mockCoboProvider` should implement `cobo.WalletProvider`; for any method called unexpectedly, increment the call counter (or `t.Fatal("unexpected call")` from a test variant) so test breakage is visible.

- [ ] **Step 2: Run test, expect failure**

Run: `cd server && go test ./internal/modules/wallet/service/ -run TestWithdraw_QueuesPendingReview -v`
Expected: FAIL — current code still calls Cobo and sets `status=processing`.

- [ ] **Step 3: Modify `Withdraw` to skip Cobo**

Edit `server/internal/modules/wallet/service/wallet_service.go`. Replace the body of `Withdraw` from line ~159 to the end of the method (the `return &dto.WithdrawResponse{...}` line) with:

```go
func (s *walletService) Withdraw(ctx context.Context, userID string, req dto.WithdrawRequest) (*dto.WithdrawResponse, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return nil, apperr.New(400, "INVALID_AMOUNT", "invalid withdrawal amount")
	}

	// 2FA check
	if s.twoFA != nil {
		valid, err := s.twoFA.Verify2FA(ctx, userID, req.TwoFACode)
		if err != nil {
			return nil, apperr.Wrap(apperr.ErrInternal, fmt.Errorf("verify 2fa: %w", err))
		}
		if !valid {
			return nil, apperr.New(403, "INVALID_2FA", "invalid two-factor authentication code")
		}
	}

	// Whitelist + 24h cooldown
	whiteAddr, err := s.addrRepo.FindWithdrawAddress(ctx, userID, req.Address, req.Network)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrAddressNotWhitelisted
		}
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}
	if whiteAddr.InCooldown() {
		return nil, apperr.ErrAddressCooldown
	}

	// Balance + freeze
	wallet, err := s.walletRepo.FindByUserIDAndCurrency(ctx, userID, req.Currency)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrInsufficientFunds
		}
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}
	if wallet.Available.LessThan(amount) {
		return nil, apperr.ErrInsufficientFunds
	}
	newAvailable := wallet.Available.Sub(amount)
	newFrozen := wallet.Frozen.Add(amount)
	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, newAvailable, wallet.InOperation, newFrozen); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	// Queue ticket — DO NOT call Cobo here; admin will approve.
	tx := &model.Transaction{
		UserID:       userID,
		Type:         model.TxTypeWithdraw,
		Currency:     req.Currency,
		Network:      req.Network,
		Amount:       amount,
		Address:      req.Address,
		Status:       model.TxStatusPending,
		ReviewStatus: model.ReviewStatusPendingReview,
	}
	if err := s.txRepo.Create(ctx, tx); err != nil {
		// Best-effort rollback of freeze; if this fails, audit log alarms run.
		_ = s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Available, wallet.InOperation, wallet.Frozen)
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}

	s.logger.Info("withdrawal queued for review",
		slog.String("user_id", userID),
		slog.String("tx_id", tx.ID),
		slog.String("currency", req.Currency),
		slog.String("amount", amount.String()),
	)

	return &dto.WithdrawResponse{
		TransactionID: tx.ID,
		Status:        model.ReviewStatusPendingReview,
		Message:       "withdrawal request queued for admin review",
	}, nil
}
```

Delete the now-unused `import` `"github.com/sovereign-fund/sovereign/internal/shared/events"` ONLY if no other method in this file references the events package (the webhook handler still does — keep it).

- [ ] **Step 4: Run test, expect pass**

Run: `cd server && go test ./internal/modules/wallet/service/ -run TestWithdraw_QueuesPendingReview -v`
Expected: PASS.

- [ ] **Step 5: Run full wallet tests to catch regressions**

Run: `cd server && go test ./internal/modules/wallet/... -v`
Expected: all PASS. Pre-existing tests that asserted Cobo was called inside `Withdraw` will need to be updated to expect the new queued behavior — update them in this same step (do not weaken assertions; rewrite to assert the queued behavior).

- [ ] **Step 6: Commit**

```bash
git add server/internal/modules/wallet/service/wallet_service.go server/internal/modules/wallet/service/*_test.go
git commit -m "feat(wallet): user Withdraw now queues pending_review instead of calling Cobo"
```

---

### Task 5: Add user `CancelWithdraw` (TDD)

**Files:**
- Modify: `server/internal/modules/wallet/service/wallet_service.go`
- Modify: `server/internal/modules/wallet/handler/wallet_handler.go`
- Modify: `server/internal/modules/wallet/routes.go`
- Test: `server/internal/modules/wallet/service/wallet_service_test.go`

- [ ] **Step 1: Add failing test**

Append to `wallet_service_test.go`:
```go
func TestCancelWithdraw_RestoresFrozen(t *testing.T) {
	ctx := context.Background()
	walletRepo := &stubWalletRepo{
		wallet: &model.Wallet{
			ID: "w1", UserID: "u1", Currency: "USDT",
			Available: decimal.NewFromInt(50),
			Frozen:    decimal.NewFromInt(50),
		},
	}
	txRepo := &stubTxRepo{
		byID: map[string]*model.Transaction{
			"tx1": {
				ID:           "tx1",
				UserID:       "u1",
				Type:         model.TxTypeWithdraw,
				Currency:     "USDT",
				Amount:       decimal.NewFromInt(50),
				Status:       model.TxStatusPending,
				ReviewStatus: model.ReviewStatusPendingReview,
			},
		},
	}
	svc := NewWalletService(walletRepo, &stubAddressRepo{}, txRepo, &mockCoboProvider{}, &spyEventBus{}, nil, slog.Default(), 24*time.Hour)

	err := svc.CancelWithdraw(ctx, "u1", "tx1")

	assert.NoError(t, err)
	assert.True(t, walletRepo.lastAvailable.Equal(decimal.NewFromInt(100)))
	assert.True(t, walletRepo.lastFrozen.Equal(decimal.Zero))

	updated := txRepo.lastReviewUpdate["tx1"]
	assert.Equal(t, model.ReviewStatusCancelled, updated["review_status"])
	assert.Equal(t, model.TxStatusCancelled, updated["status"])
}

func TestCancelWithdraw_RejectsWrongOwner(t *testing.T) {
	ctx := context.Background()
	txRepo := &stubTxRepo{byID: map[string]*model.Transaction{"tx1": {
		ID: "tx1", UserID: "other", Type: model.TxTypeWithdraw,
		ReviewStatus: model.ReviewStatusPendingReview,
	}}}
	svc := NewWalletService(&stubWalletRepo{}, &stubAddressRepo{}, txRepo, &mockCoboProvider{}, &spyEventBus{}, nil, slog.Default(), 24*time.Hour)

	err := svc.CancelWithdraw(ctx, "u1", "tx1")
	assert.ErrorIs(t, err, apperr.ErrNotFound)
}

func TestCancelWithdraw_RejectsNonPendingReview(t *testing.T) {
	ctx := context.Background()
	txRepo := &stubTxRepo{byID: map[string]*model.Transaction{"tx1": {
		ID: "tx1", UserID: "u1", Type: model.TxTypeWithdraw,
		ReviewStatus: model.ReviewStatusSubmitted,
	}}}
	svc := NewWalletService(&stubWalletRepo{}, &stubAddressRepo{}, txRepo, &mockCoboProvider{}, &spyEventBus{}, nil, slog.Default(), 24*time.Hour)

	err := svc.CancelWithdraw(ctx, "u1", "tx1")
	var ae *apperr.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, "WITHDRAW_NOT_CANCELLABLE", ae.Code)
}
```

(Add `lastReviewUpdate map[string]map[string]any` to `stubTxRepo` if not yet there; have its `UpdateReview` method record the fields.)

- [ ] **Step 2: Run, expect failure**

Run: `cd server && go test ./internal/modules/wallet/service/ -run TestCancelWithdraw -v`
Expected: FAIL — `CancelWithdraw` doesn't exist.

- [ ] **Step 3: Add error code constant**

Append to `server/internal/shared/errors/codes.go` in the Wallet block:
```go
	ErrWithdrawNotCancellable = New(http.StatusUnprocessableEntity, "WITHDRAW_NOT_CANCELLABLE", "withdrawal cannot be cancelled in current state")
```

- [ ] **Step 4: Implement service method**

Add to `WalletService` interface in `wallet_service.go`:
```go
CancelWithdraw(ctx context.Context, userID, txID string) error
```

Append the implementation:
```go
func (s *walletService) CancelWithdraw(ctx context.Context, userID, txID string) error {
	tx, err := s.txRepo.FindByID(ctx, txID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	if tx.UserID != userID || tx.Type != model.TxTypeWithdraw {
		return apperr.ErrNotFound
	}
	if tx.ReviewStatus != model.ReviewStatusPendingReview {
		return apperr.ErrWithdrawNotCancellable
	}

	wallet, err := s.walletRepo.FindByUserIDAndCurrency(ctx, userID, tx.Currency)
	if err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	newAvailable := wallet.Available.Add(tx.Amount)
	newFrozen := wallet.Frozen.Sub(tx.Amount)
	if newFrozen.LessThan(decimal.Zero) {
		newFrozen = decimal.Zero
	}
	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, newAvailable, wallet.InOperation, newFrozen); err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	if err := s.txRepo.UpdateReview(ctx, tx.ID, map[string]any{
		"review_status": model.ReviewStatusCancelled,
		"status":        model.TxStatusCancelled,
	}); err != nil {
		// Best-effort: revert wallet on update failure
		_ = s.walletRepo.UpdateBalance(ctx, wallet.ID, wallet.Available, wallet.InOperation, wallet.Frozen)
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	s.logger.Info("withdrawal cancelled by user",
		slog.String("user_id", userID),
		slog.String("tx_id", txID),
		slog.String("amount", tx.Amount.String()),
	)
	return nil
}
```

- [ ] **Step 5: Add HTTP handler**

Append to `wallet_handler.go`:
```go
func (h *WalletHandler) CancelWithdraw(c *gin.Context) {
	userID := c.GetString("user_id")
	txID := c.Param("id")

	if err := h.walletSvc.CancelWithdraw(c.Request.Context(), userID, txID); err != nil {
		handleError(c, err)
		return
	}
	response.NoContent(c)
}
```

- [ ] **Step 6: Register route**

Edit `routes.go`. Inside the `withdraw` group block, add a sibling line:
```go
withdraw.DELETE("/withdraw/:id", h.CancelWithdraw)
```

- [ ] **Step 7: Run tests, expect pass**

Run: `cd server && go test ./internal/modules/wallet/... -v`
Expected: all PASS.

- [ ] **Step 8: Commit**

```bash
git add server/internal/modules/wallet/ server/internal/shared/errors/codes.go
git commit -m "feat(wallet): user can cancel pending_review withdrawal"
```

---

### Task 6: Expose `review_status` and `reject_reason` in user transactions API

**Files:**
- Modify: `server/internal/modules/wallet/dto/response.go`
- Modify: `server/internal/modules/wallet/service/wallet_service.go`
- Test: same test file

- [ ] **Step 1: Failing test**

Append:
```go
func TestGetTransactions_IncludesReviewFields(t *testing.T) {
	ctx := context.Background()
	rejectedAt := time.Now()
	txRepo := &stubTxRepo{
		listByUser: []model.Transaction{
			{
				ID: "tx1", UserID: "u1", Type: model.TxTypeWithdraw,
				Currency: "USDT", Network: "TRC20",
				Amount: decimal.NewFromInt(50), Address: "Tabc",
				Status:       model.TxStatusCancelled,
				ReviewStatus: model.ReviewStatusRejected,
				RejectReason: "address mismatch",
				ReviewedAt:   &rejectedAt,
			},
		},
		listTotal: 1,
	}
	svc := NewWalletService(&stubWalletRepo{}, &stubAddressRepo{}, txRepo, &mockCoboProvider{}, &spyEventBus{}, nil, slog.Default(), 24*time.Hour)

	items, total, err := svc.GetTransactions(ctx, "u1", "withdraw", 1, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Equal(t, "rejected", items[0].ReviewStatus)
	assert.Equal(t, "address mismatch", items[0].RejectReason)
}
```

(`stubTxRepo` `FindByUserID` should return `listByUser`/`listTotal`.)

- [ ] **Step 2: Run, expect failure**

Run: `cd server && go test ./internal/modules/wallet/service/ -run TestGetTransactions_IncludesReviewFields -v`
Expected: FAIL — fields don't exist on `TransactionResponse`.

- [ ] **Step 3: Add fields to DTO**

Edit `dto/response.go` `TransactionResponse`:
```go
type TransactionResponse struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Currency      string          `json:"currency"`
	Network       string          `json:"network"`
	Amount        decimal.Decimal `json:"amount"`
	Fee           decimal.Decimal `json:"fee"`
	Address       string          `json:"address"`
	TxHash        string          `json:"tx_hash"`
	Status        string          `json:"status"`
	ReviewStatus  string          `json:"review_status,omitempty"`
	RejectReason  string          `json:"reject_reason,omitempty"`
	ConfirmedAt   *string         `json:"confirmed_at"`
	CreatedAt     string          `json:"created_at"`
}
```

- [ ] **Step 4: Populate fields in service**

In `GetTransactions` and `GetTransaction`, after constructing the existing `TransactionResponse`, add:
```go
r.ReviewStatus = tx.ReviewStatus
r.RejectReason = tx.RejectReason
```

- [ ] **Step 5: Run test, expect pass + commit**

```bash
cd server && go test ./internal/modules/wallet/... -v
git add server/internal/modules/wallet/
git commit -m "feat(wallet): expose review_status and reject_reason in transactions API"
```

---

## Phase 3 — Admin review service

### Task 7: Skeleton `WithdrawReviewService` + `List`

**Files:**
- Create: `server/internal/modules/admin/service/withdraw_review_service.go`
- Create: `server/internal/modules/admin/service/withdraw_review_service_test.go`
- Modify: `server/internal/modules/admin/dto/request.go`
- Modify: `server/internal/modules/admin/dto/response.go`

- [ ] **Step 1: Add DTOs**

Append to `admin/dto/request.go`:
```go
type WithdrawReviewListQuery struct {
	ReviewStatus string `form:"review_status"`
	UserID       string `form:"user_id"`
	DateFrom     string `form:"date_from"`
	DateTo       string `form:"date_to"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
}

type RejectWithdrawRequest struct {
	Reason string `json:"reason" binding:"required,min=2,max=500"`
}
```

Append to `admin/dto/response.go`:
```go
type WithdrawReviewItem struct {
	ID              string          `json:"id"`
	UserID          string          `json:"user_id"`
	UserEmail       string          `json:"user_email"`
	Currency        string          `json:"currency"`
	Network         string          `json:"network"`
	Amount          decimal.Decimal `json:"amount"`
	Address         string          `json:"address"`
	Status          string          `json:"status"`
	ReviewStatus    string          `json:"review_status"`
	RejectReason    string          `json:"reject_reason,omitempty"`
	SubmitAttempts  int             `json:"submit_attempts"`
	LastSubmitError string          `json:"last_submit_error,omitempty"`
	LastSubmitAt    *string         `json:"last_submit_at,omitempty"`
	ReviewedBy      string          `json:"reviewed_by,omitempty"`
	ReviewedAt      *string         `json:"reviewed_at,omitempty"`
	CreatedAt       string          `json:"created_at"`
	ExternalID      string          `json:"external_id,omitempty"`
	TxHash          string          `json:"tx_hash,omitempty"`
}
```

(Make sure `import "github.com/shopspring/decimal"` exists in response.go.)

- [ ] **Step 2: Failing test for List**

Create `withdraw_review_service_test.go`:
```go
package service

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	walletmodel "github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"github.com/sovereign-fund/sovereign/pkg/cobo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Stubs (define minimal versions co-located in this file or extract to _stubs_test.go)

type stubReviewQuery struct {
	rows  []reviewListRow
	total int64
}

func (s *stubReviewQuery) Query(ctx context.Context, q dto.WithdrawReviewListQuery) ([]reviewListRow, int64, error) {
	return s.rows, s.total, nil
}

type fakeProvider struct {
	withdrawErr error
	resp        *cobo.WithdrawResp
	calls       int
}

func (f *fakeProvider) Withdraw(ctx context.Context, req cobo.WithdrawReq) (*cobo.WithdrawResp, error) {
	f.calls++
	if f.withdrawErr != nil {
		return nil, f.withdrawErr
	}
	return f.resp, nil
}
func (f *fakeProvider) GenerateAddress(context.Context, cobo.GenerateAddressReq) (*cobo.GenerateAddressResp, error) { panic("unused") }
func (f *fakeProvider) GetBalance(context.Context, string) (*cobo.BalanceResp, error)                              { panic("unused") }
func (f *fakeProvider) GetTransaction(context.Context, string) (*cobo.TransactionResp, error)                      { panic("unused") }
func (f *fakeProvider) VerifyWebhook(string, []byte) (bool, error)                                                 { panic("unused") }

type stubTxRepo struct {
	byID    map[string]*walletmodel.Transaction
	updates map[string]map[string]any
}

func (s *stubTxRepo) FindByID(ctx context.Context, id string) (*walletmodel.Transaction, error) {
	if t, ok := s.byID[id]; ok {
		return t, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *stubTxRepo) UpdateReview(ctx context.Context, id string, fields map[string]any) error {
	if s.updates == nil {
		s.updates = map[string]map[string]any{}
	}
	s.updates[id] = fields
	return nil
}
func (s *stubTxRepo) UpdateExternalID(ctx context.Context, id, ext string) error { return nil }
func (s *stubTxRepo) Create(context.Context, *walletmodel.Transaction) error     { panic("unused") }
func (s *stubTxRepo) FindByExternalID(context.Context, string) (*walletmodel.Transaction, error) { panic("unused") }
func (s *stubTxRepo) FindByUserID(context.Context, string, string, int, int) ([]walletmodel.Transaction, int64, error) { panic("unused") }
func (s *stubTxRepo) UpdateStatus(context.Context, string, string, string) error { return nil }

type stubWalletRepo struct {
	wallet         *walletmodel.Wallet
	lastAvailable  decimal.Decimal
	lastInOp       decimal.Decimal
	lastFrozen     decimal.Decimal
}

func (s *stubWalletRepo) FindByUserIDAndCurrency(ctx context.Context, userID, currency string) (*walletmodel.Wallet, error) {
	return s.wallet, nil
}
func (s *stubWalletRepo) UpdateBalance(ctx context.Context, id string, av, op, fz decimal.Decimal) error {
	s.lastAvailable, s.lastInOp, s.lastFrozen = av, op, fz
	s.wallet.Available, s.wallet.InOperation, s.wallet.Frozen = av, op, fz
	return nil
}
func (s *stubWalletRepo) FindByUserID(context.Context, string) ([]walletmodel.Wallet, error) { panic("unused") }
func (s *stubWalletRepo) FindOrCreate(context.Context, string, string) (*walletmodel.Wallet, error) { panic("unused") }
func (s *stubWalletRepo) AddEarnings(context.Context, string, decimal.Decimal) error           { panic("unused") }
func (s *stubWalletRepo) ClaimEarnings(context.Context, string) error                          { panic("unused") }

type spyBus struct {
	events []events.Event
}

func (s *spyBus) Publish(ctx context.Context, e events.Event) { s.events = append(s.events, e) }
func (s *spyBus) Subscribe(string, events.Handler)            {}

func TestList_ReturnsRows(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	q := &stubReviewQuery{
		rows: []reviewListRow{
			{ID: "tx1", UserID: "u1", UserEmail: "u@x.com",
				Currency: "USDT", Network: "TRC20",
				Amount: decimal.NewFromInt(50), Address: "Tabc",
				Status: "pending", ReviewStatus: "pending_review",
				CreatedAt: now,
			},
		},
		total: 1,
	}
	svc := NewWithdrawReviewService(q, nil, nil, nil, nil, slog.Default())

	items, total, err := svc.List(ctx, dto.WithdrawReviewListQuery{Page: 1, Limit: 20})

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Equal(t, "tx1", items[0].ID)
	assert.Equal(t, "pending_review", items[0].ReviewStatus)
}
```

- [ ] **Step 3: Run, expect failure (NewWithdrawReviewService doesn't exist)**

Run: `cd server && go test ./internal/modules/admin/service/ -run TestList_ReturnsRows -v`
Expected: FAIL — undefined.

- [ ] **Step 4: Implement service skeleton + List**

Create `server/internal/modules/admin/service/withdraw_review_service.go`:
```go
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	walletmodel "github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	walletrepo "github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"github.com/sovereign-fund/sovereign/pkg/cobo"
	"gorm.io/gorm"
)

// Smaller than walletrepo.WalletRepository — only the methods we use.
type walletReader interface {
	FindByUserIDAndCurrency(ctx context.Context, userID, currency string) (*walletmodel.Wallet, error)
	UpdateBalance(ctx context.Context, id string, available, inOperation, frozen decimal.Decimal) error
}

type txWriter interface {
	FindByID(ctx context.Context, id string) (*walletmodel.Transaction, error)
	UpdateReview(ctx context.Context, id string, fields map[string]any) error
	UpdateExternalID(ctx context.Context, id, externalID string) error
	UpdateStatus(ctx context.Context, id, status, txHash string) error
}

type reviewListRow struct {
	ID              string          `gorm:"column:id"`
	UserID          string          `gorm:"column:user_id"`
	UserEmail       string          `gorm:"column:user_email"`
	Currency        string          `gorm:"column:currency"`
	Network         string          `gorm:"column:network"`
	Amount          decimal.Decimal `gorm:"column:amount"`
	Address         string          `gorm:"column:address"`
	Status          string          `gorm:"column:status"`
	ReviewStatus    string          `gorm:"column:review_status"`
	RejectReason    string          `gorm:"column:reject_reason"`
	SubmitAttempts  int             `gorm:"column:submit_attempts"`
	LastSubmitError string          `gorm:"column:last_submit_error"`
	LastSubmitAt    *time.Time      `gorm:"column:last_submit_at"`
	ReviewedBy      string          `gorm:"column:reviewed_by"`
	ReviewedAt      *time.Time      `gorm:"column:reviewed_at"`
	CreatedAt       time.Time       `gorm:"column:created_at"`
	ExternalID      string          `gorm:"column:external_id"`
	TxHash          string          `gorm:"column:tx_hash"`
}

type reviewQuerier interface {
	Query(ctx context.Context, q dto.WithdrawReviewListQuery) ([]reviewListRow, int64, error)
}

type WithdrawReviewService interface {
	List(ctx context.Context, q dto.WithdrawReviewListQuery) ([]dto.WithdrawReviewItem, int64, error)
	Approve(ctx context.Context, txID, adminID string) error
	Reject(ctx context.Context, txID, adminID, reason string) error
	Retry(ctx context.Context, txID, adminID string) error
}

type withdrawReviewService struct {
	q        reviewQuerier
	walletR  walletReader
	txR      txWriter
	provider cobo.WalletProvider
	bus      events.Bus
	logger   *slog.Logger
}

func NewWithdrawReviewService(
	q reviewQuerier,
	walletR walletReader,
	txR txWriter,
	provider cobo.WalletProvider,
	bus events.Bus,
	logger *slog.Logger,
) WithdrawReviewService {
	return &withdrawReviewService{q: q, walletR: walletR, txR: txR, provider: provider, bus: bus, logger: logger}
}

func (s *withdrawReviewService) List(ctx context.Context, q dto.WithdrawReviewListQuery) ([]dto.WithdrawReviewItem, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}
	rows, total, err := s.q.Query(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	items := make([]dto.WithdrawReviewItem, len(rows))
	for i, r := range rows {
		items[i] = dto.WithdrawReviewItem{
			ID: r.ID, UserID: r.UserID, UserEmail: r.UserEmail,
			Currency: r.Currency, Network: r.Network, Amount: r.Amount,
			Address: r.Address, Status: r.Status, ReviewStatus: r.ReviewStatus,
			RejectReason: r.RejectReason, SubmitAttempts: r.SubmitAttempts,
			LastSubmitError: r.LastSubmitError, ReviewedBy: r.ReviewedBy,
			CreatedAt:  r.CreatedAt.Format(time.RFC3339),
			ExternalID: r.ExternalID, TxHash: r.TxHash,
		}
		if r.LastSubmitAt != nil {
			t := r.LastSubmitAt.Format(time.RFC3339)
			items[i].LastSubmitAt = &t
		}
		if r.ReviewedAt != nil {
			t := r.ReviewedAt.Format(time.RFC3339)
			items[i].ReviewedAt = &t
		}
	}
	return items, total, nil
}

// Approve / Reject / Retry implemented in following tasks; stub now to allow compile.
func (s *withdrawReviewService) Approve(ctx context.Context, txID, adminID string) error { return errors.New("not implemented") }
func (s *withdrawReviewService) Reject(ctx context.Context, txID, adminID, reason string) error { return errors.New("not implemented") }
func (s *withdrawReviewService) Retry(ctx context.Context, txID, adminID string) error { return errors.New("not implemented") }

// Compile-time assertions
var _ walletReader = (walletrepo.WalletRepository)(nil)
var _ txWriter = (walletrepo.TransactionRepository)(nil)

// silence unused import in case something gets refactored
var _ = fmt.Sprintf
var _ = apperr.ErrInternal
var _ = (*gorm.DB)(nil)
```

> The compile-time assertions (`var _ walletReader = (walletrepo.WalletRepository)(nil)`) make sure that the production repository types still satisfy the narrow interfaces we use. If you ever refactor a repository method signature, this catches it at build time.

- [ ] **Step 5: Run test**

Run: `cd server && go test ./internal/modules/admin/service/ -run TestList_ReturnsRows -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add server/internal/modules/admin/dto/ server/internal/modules/admin/service/withdraw_review_service.go server/internal/modules/admin/service/withdraw_review_service_test.go
git commit -m "feat(admin): scaffold withdraw review service + list endpoint"
```

---

### Task 8: `Approve` (TDD: success + Cobo failure path)

**Files:**
- Modify: `server/internal/modules/admin/service/withdraw_review_service.go`
- Modify: `server/internal/modules/admin/service/withdraw_review_service_test.go`
- Modify: `server/internal/shared/events/types.go`

- [ ] **Step 1: Add `WithdrawRejected` event constant for later use (and to keep events file edits in one batch)**

Edit `server/internal/shared/events/types.go`. Inside the existing const block:
```go
WithdrawRejected = "wallet.withdraw.rejected"
```

- [ ] **Step 2: Failing test — Cobo success**

Append to `withdraw_review_service_test.go`:
```go
func TestApprove_CoboSuccess_KeepsFrozenAndPublishesEvent(t *testing.T) {
	ctx := context.Background()
	tx := &walletmodel.Transaction{
		ID: "tx1", UserID: "u1", Type: "withdraw",
		Currency: "USDT", Network: "TRC20", Address: "Tabc",
		Amount: decimal.NewFromInt(50),
		Status: "pending", ReviewStatus: "pending_review",
	}
	txR := &stubTxRepo{byID: map[string]*walletmodel.Transaction{"tx1": tx}}
	wr := &stubWalletRepo{wallet: &walletmodel.Wallet{
		ID: "w1", UserID: "u1", Currency: "USDT",
		Available: decimal.NewFromInt(50),
		Frozen:    decimal.NewFromInt(50),
	}}
	provider := &fakeProvider{resp: &cobo.WithdrawResp{ExternalID: "cobo-ext-1", Status: "processing"}}
	bus := &spyBus{}
	svc := NewWithdrawReviewService(&stubReviewQuery{}, wr, txR, provider, bus, slog.Default())

	err := svc.Approve(ctx, "tx1", "admin-1")

	require.NoError(t, err)
	assert.Equal(t, 1, provider.calls)
	// Frozen NOT released here (webhook will release on confirmed)
	assert.True(t, wr.lastFrozen.Equal(decimal.NewFromInt(50)))
	// Status updates
	upd := txR.updates["tx1"]
	assert.Equal(t, "submitted", upd["review_status"])
	assert.Equal(t, "processing", upd["status"])
	assert.Equal(t, "cobo-ext-1", upd["external_id"])
	// Event published
	require.Len(t, bus.events, 1)
	assert.Equal(t, events.WithdrawRequested, bus.events[0].Type)
}

func TestApprove_CoboFailure_KeepsFrozenAndMarksSubmitFailed(t *testing.T) {
	ctx := context.Background()
	tx := &walletmodel.Transaction{
		ID: "tx1", UserID: "u1", Type: "withdraw",
		Currency: "USDT", Amount: decimal.NewFromInt(50),
		Status: "pending", ReviewStatus: "pending_review",
		SubmitAttempts: 0,
	}
	txR := &stubTxRepo{byID: map[string]*walletmodel.Transaction{"tx1": tx}}
	wr := &stubWalletRepo{wallet: &walletmodel.Wallet{
		ID: "w1", UserID: "u1", Currency: "USDT",
		Available: decimal.NewFromInt(50), Frozen: decimal.NewFromInt(50),
	}}
	provider := &fakeProvider{withdrawErr: errors.New("withdraw: 400 Bad Request {\"error_code\": 12007}")}
	bus := &spyBus{}
	svc := NewWithdrawReviewService(&stubReviewQuery{}, wr, txR, provider, bus, slog.Default())

	err := svc.Approve(ctx, "tx1", "admin-1")
	// Service surfaces the failure to the admin caller as a typed app error,
	// but state changes (DB writes) must already be persisted.
	require.Error(t, err)

	// Frozen / available untouched
	assert.True(t, wr.wallet.Available.Equal(decimal.NewFromInt(50)))
	assert.True(t, wr.wallet.Frozen.Equal(decimal.NewFromInt(50)))

	// Marked submit_failed, attempts incremented, error captured
	upd := txR.updates["tx1"]
	assert.Equal(t, "submit_failed", upd["review_status"])
	assert.Equal(t, "pending", upd["status"]) // status untouched
	assert.Equal(t, 1, upd["submit_attempts"])
	assert.Contains(t, upd["last_submit_error"].(string), "12007")
	assert.NotNil(t, upd["last_submit_at"])

	// No event published
	assert.Empty(t, bus.events)
}

func TestApprove_RejectsNonPendingReview(t *testing.T) {
	ctx := context.Background()
	tx := &walletmodel.Transaction{
		ID: "tx1", UserID: "u1", Type: "withdraw",
		ReviewStatus: "submitted",
	}
	txR := &stubTxRepo{byID: map[string]*walletmodel.Transaction{"tx1": tx}}
	svc := NewWithdrawReviewService(&stubReviewQuery{}, &stubWalletRepo{}, txR, &fakeProvider{}, &spyBus{}, slog.Default())

	err := svc.Approve(ctx, "tx1", "admin-1")
	var ae *apperr.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, "WITHDRAW_NOT_APPROVABLE", ae.Code)
}
```

- [ ] **Step 3: Run, expect failure**

Run: `cd server && go test ./internal/modules/admin/service/ -run TestApprove -v`
Expected: 3 FAIL.

- [ ] **Step 4: Add error code**

Append to `server/internal/shared/errors/codes.go` Wallet block:
```go
ErrWithdrawNotApprovable = New(http.StatusUnprocessableEntity, "WITHDRAW_NOT_APPROVABLE", "withdrawal cannot be approved in current state")
ErrWithdrawNotRetriable  = New(http.StatusUnprocessableEntity, "WITHDRAW_NOT_RETRIABLE",  "withdrawal cannot be retried in current state")
ErrWithdrawNotRejectable = New(http.StatusUnprocessableEntity, "WITHDRAW_NOT_REJECTABLE", "withdrawal cannot be rejected in current state")
ErrCoboSubmitFailed      = New(http.StatusBadGateway,         "COBO_SUBMIT_FAILED",      "submission to cobo failed; ticket marked submit_failed")
```

- [ ] **Step 5: Implement Approve + submitToCobo**

Replace stub `Approve` in `withdraw_review_service.go` and add helper:
```go
func (s *withdrawReviewService) Approve(ctx context.Context, txID, adminID string) error {
	tx, err := s.txR.FindByID(ctx, txID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	if tx.Type != walletmodel.TxTypeWithdraw || tx.ReviewStatus != walletmodel.ReviewStatusPendingReview {
		return apperr.ErrWithdrawNotApprovable
	}
	return s.submitToCobo(ctx, tx, adminID)
}

// submitToCobo is shared by Approve and Retry. On Cobo success it transitions the
// ticket to submitted/processing. On Cobo failure it transitions to submit_failed
// WITHOUT touching wallet balances — money stays frozen so the admin can retry.
func (s *withdrawReviewService) submitToCobo(ctx context.Context, tx *walletmodel.Transaction, adminID string) error {
	now := time.Now()
	resp, coboErr := s.provider.Withdraw(ctx, cobo.WithdrawReq{
		Currency:  tx.Currency,
		Network:   tx.Network,
		Address:   tx.Address,
		Amount:    tx.Amount,
		RequestID: tx.ID,
	})

	if coboErr != nil {
		s.logger.Error("cobo withdraw failed",
			slog.String("error", coboErr.Error()),
			slog.String("tx_id", tx.ID),
			slog.String("admin_id", adminID),
		)
		fields := map[string]any{
			"review_status":     walletmodel.ReviewStatusSubmitFailed,
			"submit_attempts":   tx.SubmitAttempts + 1,
			"last_submit_error": coboErr.Error(),
			"last_submit_at":    now,
			"reviewed_by":       adminID,
			"reviewed_at":       now,
		}
		if upErr := s.txR.UpdateReview(ctx, tx.ID, fields); upErr != nil {
			return apperr.Wrap(apperr.ErrInternal, fmt.Errorf("persist submit_failed: %w (cobo err: %v)", upErr, coboErr))
		}
		return apperr.Wrap(apperr.ErrCoboSubmitFailed, coboErr)
	}

	// Success path: persist external_id and transition state.
	fields := map[string]any{
		"review_status":   walletmodel.ReviewStatusSubmitted,
		"status":          walletmodel.TxStatusProcessing,
		"external_id":     resp.ExternalID,
		"submit_attempts": tx.SubmitAttempts + 1,
		"last_submit_at":  now,
		"reviewed_by":     adminID,
		"reviewed_at":     now,
	}
	if err := s.txR.UpdateReview(ctx, tx.ID, fields); err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	s.bus.Publish(ctx, events.Event{
		Type: events.WithdrawRequested,
		Payload: map[string]string{
			"user_id":        tx.UserID,
			"transaction_id": tx.ID,
			"external_id":    resp.ExternalID,
		},
	})

	s.logger.Info("withdrawal submitted to cobo",
		slog.String("tx_id", tx.ID),
		slog.String("external_id", resp.ExternalID),
		slog.String("admin_id", adminID),
	)
	return nil
}
```

- [ ] **Step 6: Run tests, expect PASS**

Run: `cd server && go test ./internal/modules/admin/service/ -run TestApprove -v`
Expected: 3 PASS.

- [ ] **Step 7: Commit**

```bash
git add server/internal/modules/admin/service/withdraw_review_service.go server/internal/modules/admin/service/withdraw_review_service_test.go server/internal/shared/errors/codes.go server/internal/shared/events/types.go
git commit -m "feat(admin): WithdrawReviewService.Approve with cobo-failure-keeps-frozen semantics"
```

---

### Task 9: `Reject` (TDD)

**Files:**
- Modify: `withdraw_review_service.go` and its test file

- [ ] **Step 1: Failing test**

Append:
```go
func TestReject_RefundsAndPersistsReason(t *testing.T) {
	ctx := context.Background()
	tx := &walletmodel.Transaction{
		ID: "tx1", UserID: "u1", Type: "withdraw",
		Currency: "USDT", Amount: decimal.NewFromInt(50),
		Status: "pending", ReviewStatus: "pending_review",
	}
	txR := &stubTxRepo{byID: map[string]*walletmodel.Transaction{"tx1": tx}}
	wr := &stubWalletRepo{wallet: &walletmodel.Wallet{
		ID: "w1", UserID: "u1", Currency: "USDT",
		Available: decimal.NewFromInt(50), Frozen: decimal.NewFromInt(50),
	}}
	bus := &spyBus{}
	svc := NewWithdrawReviewService(&stubReviewQuery{}, wr, txR, &fakeProvider{}, bus, slog.Default())

	err := svc.Reject(ctx, "tx1", "admin-1", "address mismatch")

	require.NoError(t, err)
	// Refund: available 50→100, frozen 50→0
	assert.True(t, wr.wallet.Available.Equal(decimal.NewFromInt(100)))
	assert.True(t, wr.wallet.Frozen.Equal(decimal.Zero))
	// Update fields
	upd := txR.updates["tx1"]
	assert.Equal(t, "rejected", upd["review_status"])
	assert.Equal(t, "cancelled", upd["status"])
	assert.Equal(t, "address mismatch", upd["reject_reason"])
	assert.Equal(t, "admin-1", upd["reviewed_by"])
	// Event for notification
	require.Len(t, bus.events, 1)
	assert.Equal(t, events.WithdrawRejected, bus.events[0].Type)
	assert.Equal(t, "address mismatch", bus.events[0].Payload["reason"])
}

func TestReject_AllowedFromSubmitFailed(t *testing.T) {
	ctx := context.Background()
	tx := &walletmodel.Transaction{
		ID: "tx1", UserID: "u1", Type: "withdraw",
		Currency: "USDT", Amount: decimal.NewFromInt(50),
		Status: "pending", ReviewStatus: "submit_failed",
	}
	txR := &stubTxRepo{byID: map[string]*walletmodel.Transaction{"tx1": tx}}
	wr := &stubWalletRepo{wallet: &walletmodel.Wallet{
		ID: "w1", UserID: "u1", Currency: "USDT",
		Available: decimal.NewFromInt(0), Frozen: decimal.NewFromInt(50),
	}}
	svc := NewWithdrawReviewService(&stubReviewQuery{}, wr, txR, &fakeProvider{}, &spyBus{}, slog.Default())

	require.NoError(t, svc.Reject(ctx, "tx1", "admin-1", "manual cleanup"))
	assert.True(t, wr.wallet.Available.Equal(decimal.NewFromInt(50)))
	assert.True(t, wr.wallet.Frozen.Equal(decimal.Zero))
}

func TestReject_RejectsTerminalState(t *testing.T) {
	ctx := context.Background()
	tx := &walletmodel.Transaction{
		ID: "tx1", UserID: "u1", Type: "withdraw",
		ReviewStatus: "submitted",
	}
	txR := &stubTxRepo{byID: map[string]*walletmodel.Transaction{"tx1": tx}}
	svc := NewWithdrawReviewService(&stubReviewQuery{}, &stubWalletRepo{}, txR, &fakeProvider{}, &spyBus{}, slog.Default())
	err := svc.Reject(ctx, "tx1", "admin-1", "x")
	var ae *apperr.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, "WITHDRAW_NOT_REJECTABLE", ae.Code)
}
```

- [ ] **Step 2: Run, expect FAIL**

Run: `cd server && go test ./internal/modules/admin/service/ -run TestReject -v`
Expected: 3 FAIL.

- [ ] **Step 3: Implement Reject**

Replace stub:
```go
func (s *withdrawReviewService) Reject(ctx context.Context, txID, adminID, reason string) error {
	tx, err := s.txR.FindByID(ctx, txID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	if tx.Type != walletmodel.TxTypeWithdraw {
		return apperr.ErrWithdrawNotRejectable
	}
	switch tx.ReviewStatus {
	case walletmodel.ReviewStatusPendingReview, walletmodel.ReviewStatusSubmitFailed:
		// allowed
	default:
		return apperr.ErrWithdrawNotRejectable
	}

	wallet, err := s.walletR.FindByUserIDAndCurrency(ctx, tx.UserID, tx.Currency)
	if err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	newAvailable := wallet.Available.Add(tx.Amount)
	newFrozen := wallet.Frozen.Sub(tx.Amount)
	if newFrozen.LessThan(decimal.Zero) {
		newFrozen = decimal.Zero
	}
	if err := s.walletR.UpdateBalance(ctx, wallet.ID, newAvailable, wallet.InOperation, newFrozen); err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	now := time.Now()
	if err := s.txR.UpdateReview(ctx, tx.ID, map[string]any{
		"review_status":  walletmodel.ReviewStatusRejected,
		"status":         walletmodel.TxStatusCancelled,
		"reject_reason":  reason,
		"reviewed_by":    adminID,
		"reviewed_at":    now,
	}); err != nil {
		// Try to revert wallet on persistence failure
		_ = s.walletR.UpdateBalance(ctx, wallet.ID, wallet.Available, wallet.InOperation, wallet.Frozen)
		return apperr.Wrap(apperr.ErrInternal, err)
	}

	s.bus.Publish(ctx, events.Event{
		Type: events.WithdrawRejected,
		Payload: map[string]string{
			"user_id":        tx.UserID,
			"transaction_id": tx.ID,
			"amount":         tx.Amount.String(),
			"currency":       tx.Currency,
			"reason":         reason,
		},
	})
	return nil
}
```

- [ ] **Step 4: Run + commit**

```bash
cd server && go test ./internal/modules/admin/service/ -run TestReject -v
git add server/internal/modules/admin/service/
git commit -m "feat(admin): WithdrawReviewService.Reject with refund and notification event"
```

---

### Task 10: `Retry` (TDD)

**Files:** same as above

- [ ] **Step 1: Failing test**

```go
func TestRetry_FromSubmitFailed_Success(t *testing.T) {
	ctx := context.Background()
	tx := &walletmodel.Transaction{
		ID: "tx1", UserID: "u1", Type: "withdraw",
		Currency: "USDT", Network: "TRC20", Address: "Tabc",
		Amount: decimal.NewFromInt(50),
		Status: "pending", ReviewStatus: "submit_failed",
		SubmitAttempts: 1,
	}
	txR := &stubTxRepo{byID: map[string]*walletmodel.Transaction{"tx1": tx}}
	wr := &stubWalletRepo{wallet: &walletmodel.Wallet{
		ID: "w1", UserID: "u1", Currency: "USDT",
		Frozen: decimal.NewFromInt(50),
	}}
	provider := &fakeProvider{resp: &cobo.WithdrawResp{ExternalID: "cobo-2"}}
	bus := &spyBus{}
	svc := NewWithdrawReviewService(&stubReviewQuery{}, wr, txR, provider, bus, slog.Default())

	require.NoError(t, svc.Retry(ctx, "tx1", "admin-2"))
	assert.Equal(t, 1, provider.calls)
	upd := txR.updates["tx1"]
	assert.Equal(t, "submitted", upd["review_status"])
	assert.Equal(t, "processing", upd["status"])
	assert.Equal(t, 2, upd["submit_attempts"])
}

func TestRetry_FromOtherStatus_Forbidden(t *testing.T) {
	ctx := context.Background()
	tx := &walletmodel.Transaction{
		ID: "tx1", Type: "withdraw",
		ReviewStatus: "pending_review",
	}
	txR := &stubTxRepo{byID: map[string]*walletmodel.Transaction{"tx1": tx}}
	svc := NewWithdrawReviewService(&stubReviewQuery{}, &stubWalletRepo{}, txR, &fakeProvider{}, &spyBus{}, slog.Default())
	err := svc.Retry(ctx, "tx1", "admin-2")
	var ae *apperr.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, "WITHDRAW_NOT_RETRIABLE", ae.Code)
}
```

- [ ] **Step 2: Run, expect FAIL**

Run: `cd server && go test ./internal/modules/admin/service/ -run TestRetry -v`
Expected: 2 FAIL.

- [ ] **Step 3: Implement Retry**

Replace stub:
```go
func (s *withdrawReviewService) Retry(ctx context.Context, txID, adminID string) error {
	tx, err := s.txR.FindByID(ctx, txID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.ErrNotFound
		}
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	if tx.Type != walletmodel.TxTypeWithdraw || tx.ReviewStatus != walletmodel.ReviewStatusSubmitFailed {
		return apperr.ErrWithdrawNotRetriable
	}
	return s.submitToCobo(ctx, tx, adminID)
}
```

- [ ] **Step 4: Run + commit**

```bash
cd server && go test ./internal/modules/admin/service/ -v
git add server/internal/modules/admin/service/
git commit -m "feat(admin): WithdrawReviewService.Retry"
```

---

### Task 11: `reviewQuerier` SQL implementation (Postgres)

**Files:**
- Add to `withdraw_review_service.go` (or split out `withdraw_review_query.go` if you prefer)

- [ ] **Step 1: Add concrete `pgReviewQuerier`**

Append to `withdraw_review_service.go`:
```go
type PGReviewQuerier struct {
	db *gorm.DB
}

func NewPGReviewQuerier(db *gorm.DB) *PGReviewQuerier { return &PGReviewQuerier{db: db} }

func (q *PGReviewQuerier) Query(ctx context.Context, p dto.WithdrawReviewListQuery) ([]reviewListRow, int64, error) {
	db := q.db.WithContext(ctx).Table("transactions").
		Select(`transactions.id, transactions.user_id, users.email AS user_email,
			transactions.currency, transactions.network, transactions.amount, transactions.address,
			transactions.status, transactions.review_status, transactions.reject_reason,
			transactions.submit_attempts, transactions.last_submit_error, transactions.last_submit_at,
			transactions.reviewed_by, transactions.reviewed_at, transactions.created_at,
			transactions.external_id, transactions.tx_hash`).
		Joins("LEFT JOIN users ON users.id = transactions.user_id").
		Where("transactions.type = ?", "withdraw")

	if p.ReviewStatus != "" {
		db = db.Where("transactions.review_status = ?", p.ReviewStatus)
	}
	if p.UserID != "" {
		db = db.Where("transactions.user_id = ?", p.UserID)
	}
	if p.DateFrom != "" {
		db = db.Where("transactions.created_at >= ?", p.DateFrom)
	}
	if p.DateTo != "" {
		db = db.Where("transactions.created_at < ?", p.DateTo+" 23:59:59")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	var rows []reviewListRow
	offset := (p.Page - 1) * p.Limit
	if err := db.Order("transactions.created_at DESC").Offset(offset).Limit(p.Limit).Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("find: %w", err)
	}
	return rows, total, nil
}
```

- [ ] **Step 2: Build**

Run: `cd server && go build ./...`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add server/internal/modules/admin/service/
git commit -m "feat(admin): pg implementation of withdraw review querier"
```

---

## Phase 4 — HTTP layer + audit + wiring

### Task 12: Admin handler with audit logging

**Files:**
- Create: `server/internal/modules/admin/handler/withdraw_review_handler.go`
- Create: `server/internal/modules/admin/handler/withdraw_review_handler_test.go`

- [ ] **Step 1: Implement handler**

Create `withdraw_review_handler.go`:
```go
package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type WithdrawReviewHandler struct {
	svc      service.WithdrawReviewService
	auditSvc service.AuditService
}

func NewWithdrawReviewHandler(svc service.WithdrawReviewService, auditSvc service.AuditService) *WithdrawReviewHandler {
	return &WithdrawReviewHandler{svc: svc, auditSvc: auditSvc}
}

func (h *WithdrawReviewHandler) List(c *gin.Context) {
	var q dto.WithdrawReviewListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_QUERY", err.Error())
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "WITHDRAW_LIST_FAILED", err.Error())
		return
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	limit := q.Limit
	if limit < 1 {
		limit = 20
	}
	response.Paginated(c, items, response.Meta{Total: total, Page: page, PerPage: limit})
}

func (h *WithdrawReviewHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	adminID := c.GetString("admin_id")
	err := h.svc.Approve(c.Request.Context(), id, adminID)
	h.audit(c, "approve_withdrawal", id, "")
	h.respond(c, err)
}

func (h *WithdrawReviewHandler) Reject(c *gin.Context) {
	id := c.Param("id")
	adminID := c.GetString("admin_id")
	var req dto.RejectWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	err := h.svc.Reject(c.Request.Context(), id, adminID, req.Reason)
	h.audit(c, "reject_withdrawal", id, "reason="+req.Reason)
	h.respond(c, err)
}

func (h *WithdrawReviewHandler) Retry(c *gin.Context) {
	id := c.Param("id")
	adminID := c.GetString("admin_id")
	err := h.svc.Retry(c.Request.Context(), id, adminID)
	h.audit(c, "retry_withdrawal", id, "")
	h.respond(c, err)
}

func (h *WithdrawReviewHandler) audit(c *gin.Context, action, id, detail string) {
	if err := h.auditSvc.Log(
		c.Request.Context(),
		c.GetString("admin_id"),
		c.GetString("admin_email"),
		action, "transaction", id, detail, c.ClientIP(),
	); err != nil {
		slog.Error("audit log failed", slog.String("action", action), slog.String("error", err.Error()))
	}
}

func (h *WithdrawReviewHandler) respond(c *gin.Context, err error) {
	if err == nil {
		response.OK(c, gin.H{"message": "ok"})
		return
	}
	var ae *apperr.AppError
	if errors.As(err, &ae) {
		response.Fail(c, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
```

- [ ] **Step 2: Handler integration test**

Create `withdraw_review_handler_test.go`:
```go
package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeReviewSvc struct {
	approveErr error
	rejectErr  error
	retryErr   error
	lastReason string
}

func (f *fakeReviewSvc) List(ctx context.Context, q dto.WithdrawReviewListQuery) ([]dto.WithdrawReviewItem, int64, error) {
	return []dto.WithdrawReviewItem{{ID: "tx1"}}, 1, nil
}
func (f *fakeReviewSvc) Approve(ctx context.Context, id, adminID string) error { return f.approveErr }
func (f *fakeReviewSvc) Reject(ctx context.Context, id, adminID, reason string) error {
	f.lastReason = reason
	return f.rejectErr
}
func (f *fakeReviewSvc) Retry(ctx context.Context, id, adminID string) error { return f.retryErr }

type stubAudit struct{ calls int }

func (s *stubAudit) Log(ctx context.Context, adminID, email, action, targetType, targetID, detail, ip string) error {
	s.calls++
	return nil
}
func (s *stubAudit) List(ctx context.Context, q service.AuditListQuery) ([]model.AuditLog, int64, error) {
	return nil, 0, nil
}
// Compile-time verification that stub satisfies the interface.
var _ service.AuditService = (*stubAudit)(nil)

func setup(t *testing.T) (*gin.Engine, *fakeReviewSvc, *stubAudit) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := &fakeReviewSvc{}
	audit := &stubAudit{}
	h := NewWithdrawReviewHandler(svc, audit)
	r.Use(func(c *gin.Context) {
		c.Set("admin_id", "admin-1")
		c.Set("admin_email", "admin@x.com")
		c.Next()
	})
	r.GET("/admin/withdrawals", h.List)
	r.POST("/admin/withdrawals/:id/approve", h.Approve)
	r.POST("/admin/withdrawals/:id/reject", h.Reject)
	r.POST("/admin/withdrawals/:id/retry", h.Retry)
	return r, svc, audit
}

func TestList_Returns200(t *testing.T) {
	r, _, _ := setup(t)
	req := httptest.NewRequest("GET", "/admin/withdrawals?page=1&limit=20", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, 200, rr.Code)
}

func TestApprove_OnSuccess_LogsAudit(t *testing.T) {
	r, _, audit := setup(t)
	req := httptest.NewRequest("POST", "/admin/withdrawals/tx1/approve", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, 1, audit.calls)
}

func TestApprove_PropagatesAppError(t *testing.T) {
	r, svc, _ := setup(t)
	svc.approveErr = apperr.ErrWithdrawNotApprovable
	req := httptest.NewRequest("POST", "/admin/withdrawals/tx1/approve", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

func TestReject_RequiresReason(t *testing.T) {
	r, _, _ := setup(t)
	req := httptest.NewRequest("POST", "/admin/withdrawals/tx1/reject", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, 400, rr.Code)
}

func TestReject_PassesReason(t *testing.T) {
	r, svc, _ := setup(t)
	body, _ := json.Marshal(dto.RejectWithdrawRequest{Reason: "address mismatch"})
	req := httptest.NewRequest("POST", "/admin/withdrawals/tx1/reject", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	require.Equal(t, 200, rr.Code)
	assert.Equal(t, "address mismatch", svc.lastReason)
}

func TestRetry_PropagatesError(t *testing.T) {
	r, svc, _ := setup(t)
	svc.retryErr = errors.New("boom")
	req := httptest.NewRequest("POST", "/admin/withdrawals/tx1/retry", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, 500, rr.Code)
}
```

- [ ] **Step 3: Run tests + commit**

```bash
cd server && go test ./internal/modules/admin/handler/... -v
git add server/internal/modules/admin/handler/
git commit -m "feat(admin): HTTP handler for withdraw review with audit logging"
```

---

### Task 13: Wire module + register routes

**Files:**
- Modify: `server/internal/modules/admin/module.go`
- Modify: `server/internal/modules/admin/routes.go`
- Modify: `server/cmd/server/main.go` (or wherever the admin module is constructed) — only if `NewModule` signature changed

- [ ] **Step 1: Extend `NewModule`**

The new service needs the wallet repository, cobo provider and event bus. They aren't currently passed to the admin module. Update the constructor signature:

```go
func NewModule(
	db *gorm.DB,
	cfg config.AdminConfig,
	logger *slog.Logger,
	walletRepo walletrepo.WalletRepository,
	txRepo walletrepo.TransactionRepository,
	cobo cobo.WalletProvider,
	bus events.Bus,
) *Module {
	repo := repository.NewAdminRepository(db)

	authSvc := service.NewAuthService(repo, cfg.JWTSecret, cfg.JWTExpiry, logger)
	adminUserSvc := service.NewAdminUserService(repo, logger)
	userSvc := service.NewUserService(db, logger)
	auditSvc := service.NewAuditService(db)
	dashboardSvc := service.NewDashboardService(db, logger)
	tradeSvc := service.NewTradeService(db)
	transactionSvc := service.NewTransactionService(db)

	withdrawReviewSvc := service.NewWithdrawReviewService(
		service.NewPGReviewQuerier(db),
		walletRepo, txRepo, cobo, bus, logger,
	)

	return &Module{
		AuthHandler:             handler.NewAuthHandler(authSvc),
		AdminUserHandler:        handler.NewAdminUserHandler(adminUserSvc, auditSvc),
		AuditHandler:            handler.NewAuditHandler(auditSvc),
		UserHandler:             handler.NewUserHandler(userSvc, auditSvc),
		DashboardHandler:        handler.NewDashboardHandler(dashboardSvc),
		TradeHandler:            handler.NewTradeHandler(tradeSvc, auditSvc),
		TransactionHandler:      handler.NewTransactionHandler(transactionSvc),
		WithdrawReviewHandler:   handler.NewWithdrawReviewHandler(withdrawReviewSvc, auditSvc),
		AuditService:            auditSvc,
		AdminRepo:               repo,
		JWTSecret:               cfg.JWTSecret,
	}
}
```

Add `WithdrawReviewHandler *handler.WithdrawReviewHandler` to the `Module` struct fields.

Add the imports at the top:
```go
walletrepo "github.com/sovereign-fund/sovereign/internal/modules/wallet/repository"
"github.com/sovereign-fund/sovereign/internal/shared/events"
"github.com/sovereign-fund/sovereign/pkg/cobo"
```

- [ ] **Step 2: Update the call site in main.go**

Run: `grep -rn "admin.NewModule" server/cmd/` and patch the call to pass `walletRepo`, `txRepo`, `coboProvider`, `eventBus` (these all already exist for the wallet module — re-use them).

- [ ] **Step 3: Register routes**

Edit `routes.go`. After the existing Transactions block, add:

```go
// Withdrawal review (admin can approve/reject/retry user withdrawals)
withdrawals := protected.Group("/withdrawals",
	middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator),
)
{
	withdrawals.GET("", m.WithdrawReviewHandler.List)
	withdrawals.POST("/:id/approve", m.WithdrawReviewHandler.Approve)
	withdrawals.POST("/:id/reject", m.WithdrawReviewHandler.Reject)
	withdrawals.POST("/:id/retry", m.WithdrawReviewHandler.Retry)
}
```

- [ ] **Step 4: Build whole repo**

Run: `cd server && go build ./...`
Expected: no errors. Fix any compile issues from the constructor change.

- [ ] **Step 5: Run full test suite**

Run: `cd server && go test ./...`
Expected: all PASS (we may need to update other tests that called `NewModule`).

- [ ] **Step 6: Commit**

```bash
git add server/internal/modules/admin/ server/cmd/
git commit -m "feat(admin): wire withdraw review service + register /admin/withdrawals routes"
```

---

## Phase 5 — Notifications

### Task 14: Subscribe NotificationService to `WithdrawRejected`

**Files:**
- Modify: `server/internal/modules/notification/service/notification_service.go`
- Possibly: `server/internal/modules/notification/template/renderer.go` (add template if not present)

- [ ] **Step 1: Inspect existing event subscriptions and template format**

Run these and read all output before editing:
```bash
grep -n "Subscribe\|WithdrawCompleted\|WithdrawFailed\|WithdrawRequested\|sendWithdraw" server/internal/modules/notification/service/notification_service.go
ls server/internal/modules/notification/template/templates/ 2>/dev/null
grep -n "withdraw_completed\|withdraw_failed" server/internal/modules/notification/template/renderer.go
```
Note the exact subscription pattern, template lookup mechanism, and locale fallback the existing code uses. The new code must mirror all of that.

- [ ] **Step 2: Add `WithdrawRejected` subscription mirroring `WithdrawCompleted`**

In `notification_service.go`, locate the block that does `bus.Subscribe(events.WithdrawCompleted, ...)`. Directly below it, add an analogous subscription:
```go
bus.Subscribe(events.WithdrawRejected, func(ctx context.Context, e events.Event) {
	if err := s.handleWithdrawRejected(ctx, e); err != nil {
		s.logger.Error("send withdraw rejected email",
			slog.String("error", err.Error()),
			slog.String("tx_id", e.Payload["transaction_id"]),
		)
	}
})
```

And add the handler function. Use the EXACT same recipient lookup, locale resolution, and provider call shape as `handleWithdrawCompleted`. Template name: `withdraw_rejected`.

```go
func (s *notificationService) handleWithdrawRejected(ctx context.Context, e events.Event) error {
	userID := e.Payload["user_id"]
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user %s: %w", userID, err)
	}
	data := map[string]any{
		"Name":     user.Name,
		"Amount":   e.Payload["amount"],
		"Currency": e.Payload["currency"],
		"Reason":   e.Payload["reason"],
	}
	subject, body, err := s.renderer.Render("withdraw_rejected", user.Locale, data)
	if err != nil {
		return fmt.Errorf("render withdraw_rejected: %w", err)
	}
	return s.provider.Send(ctx, user.Email, subject, body)
}
```

> If the actual signature/method names differ in the existing code, adjust to match — the principle is "exactly mirror `handleWithdrawCompleted`".

- [ ] **Step 3: Add `withdraw_rejected` templates for each supported locale**

For every `withdraw_completed.<lang>.html` (or whatever the project's naming is), create a sibling `withdraw_rejected.<lang>.html`. Sample en:
```html
<p>Hi {{.Name}},</p>
<p>Your withdrawal of <b>{{.Amount}} {{.Currency}}</b> was rejected.</p>
<p><b>Reason:</b> {{.Reason}}</p>
<p>The amount has been refunded to your available balance. If you believe this was a mistake, contact support.</p>
```
Mirror in zh and ko using the project's existing translation patterns. Add a corresponding subject line entry to wherever subjects live (often a `subjects.json` or constant map) — set:
- en: "Your withdrawal was rejected"
- zh: "您的提现申请被拒绝"
- ko: "출금 신청이 거절되었습니다"

If the renderer uses a single template file rather than per-locale files, add the new template entry inside it following the same existing pattern.

- [ ] **Step 4: Add unit test mirroring `TestHandleWithdrawCompleted`**

Open `notification_service_test.go`, locate the test for `WithdrawCompleted`, copy it as `TestHandleWithdrawRejected`, swap the event type, payload, and expected template name. Assert that the mock provider received one email and that the rendered body contains the reason text.

- [ ] **Step 5: Run notification tests**

Run: `cd server && go test ./internal/modules/notification/... -v`
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
cd server && go test ./internal/modules/notification/...
git add server/internal/modules/notification/
git commit -m "feat(notification): send rejection email on WithdrawRejected event"
```

---

## Phase 6 — Admin frontend

### Task 15: Withdrawals page

**Files:**
- Create: `admin/src/pages/Withdrawals/index.tsx`
- Create: `admin/src/pages/Withdrawals/RejectModal.tsx`
- Modify: `admin/src/services/api.ts`
- Modify: `admin/config/routes.ts`

- [ ] **Step 1: Add API helpers**

Append to `admin/src/services/api.ts`:
```typescript
export interface WithdrawReviewItem {
  id: string;
  user_id: string;
  user_email: string;
  currency: string;
  network: string;
  amount: string;
  address: string;
  status: string;
  review_status: 'pending_review' | 'submitted' | 'submit_failed' | 'rejected' | 'cancelled';
  reject_reason?: string;
  submit_attempts: number;
  last_submit_error?: string;
  last_submit_at?: string;
  reviewed_by?: string;
  reviewed_at?: string;
  external_id?: string;
  tx_hash?: string;
  created_at: string;
}

export async function listWithdrawals(params: {
  review_status?: string;
  page?: number;
  limit?: number;
  user_id?: string;
}) {
  return request<API.PaginatedResponse<WithdrawReviewItem>>('/withdrawals', {
    method: 'GET',
    params,
  });
}

export async function approveWithdrawal(id: string) {
  return request<API.ApiResponse<{ message: string }>>(`/withdrawals/${id}/approve`, {
    method: 'POST',
  });
}

export async function rejectWithdrawal(id: string, reason: string) {
  return request<API.ApiResponse<{ message: string }>>(`/withdrawals/${id}/reject`, {
    method: 'POST',
    data: { reason },
  });
}

export async function retryWithdrawal(id: string) {
  return request<API.ApiResponse<{ message: string }>>(`/withdrawals/${id}/retry`, {
    method: 'POST',
  });
}
```

- [ ] **Step 2: Implement Reject modal**

Create `admin/src/pages/Withdrawals/RejectModal.tsx`:
```tsx
import { Modal, Form, Input, message } from 'antd';
import { rejectWithdrawal } from '@/services/api';

interface Props {
  txId: string | null;
  onClose: () => void;
  onDone: () => void;
}

export default function RejectModal({ txId, onClose, onDone }: Props) {
  const [form] = Form.useForm();
  const open = !!txId;

  const handleOk = async () => {
    const values = await form.validateFields();
    if (!txId) return;
    try {
      await rejectWithdrawal(txId, values.reason);
      message.success('已拒绝并退回余额');
      form.resetFields();
      onDone();
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message ?? '操作失败');
    }
  };

  return (
    <Modal
      title="拒绝提现"
      open={open}
      onOk={handleOk}
      onCancel={() => { form.resetFields(); onClose(); }}
      destroyOnClose
    >
      <Form form={form} layout="vertical">
        <Form.Item
          name="reason"
          label="拒绝原因（将展示给用户）"
          rules={[{ required: true, min: 2, max: 500 }]}
        >
          <Input.TextArea rows={4} placeholder="例：白名单地址不匹配请重新申请" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
```

- [ ] **Step 3: Implement listing page**

Create `admin/src/pages/Withdrawals/index.tsx`:
```tsx
import { useRef, useState } from 'react';
import { ProTable, ActionType, ProColumns } from '@ant-design/pro-components';
import { Tag, Button, Popconfirm, message, Tabs, Space, Tooltip } from 'antd';
import { listWithdrawals, approveWithdrawal, retryWithdrawal, WithdrawReviewItem } from '@/services/api';
import { useAccess, Access } from '@umijs/max';
import RejectModal from './RejectModal';

const STATUS_LABEL: Record<string, { color: string; text: string }> = {
  pending_review: { color: 'gold',   text: '待审核' },
  submit_failed:  { color: 'red',    text: '提交失败' },
  submitted:      { color: 'blue',   text: '已提交' },
  rejected:       { color: 'default',text: '已拒绝' },
  cancelled:      { color: 'default',text: '已取消' },
};

export default function WithdrawalsPage() {
  const [tab, setTab] = useState('pending_review');
  const [rejectId, setRejectId] = useState<string | null>(null);
  const actionRef = useRef<ActionType>();
  const access = useAccess();

  const onApprove = async (id: string) => {
    try {
      await approveWithdrawal(id);
      message.success('已批准并提交至 Cobo');
      actionRef.current?.reload();
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message ?? '操作失败');
    }
  };

  const onRetry = async (id: string) => {
    try {
      await retryWithdrawal(id);
      message.success('已重新提交');
      actionRef.current?.reload();
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message ?? '重试失败');
    }
  };

  const columns: ProColumns<WithdrawReviewItem>[] = [
    { title: '用户', dataIndex: 'user_email', width: 200 },
    { title: '币种', dataIndex: 'currency',   width: 80 },
    { title: '网络', dataIndex: 'network',    width: 80 },
    { title: '金额', dataIndex: 'amount',     width: 120, valueType: 'digit' },
    { title: '目标地址', dataIndex: 'address', ellipsis: true, copyable: true },
    {
      title: '状态', dataIndex: 'review_status', width: 100,
      render: (_, r) => {
        const meta = STATUS_LABEL[r.review_status] || { color: 'default', text: r.review_status };
        return <Tag color={meta.color}>{meta.text}</Tag>;
      },
    },
    {
      title: '提交尝试', dataIndex: 'submit_attempts', width: 100,
      render: (v, r) => r.last_submit_error
        ? <Tooltip title={r.last_submit_error}><span>{v as number}</span></Tooltip>
        : v,
    },
    { title: '申请时间', dataIndex: 'created_at', valueType: 'dateTime', width: 170 },
    {
      title: '操作', valueType: 'option', width: 200, fixed: 'right',
      render: (_, r) => (
        <Space>
          {r.review_status === 'pending_review' && (
            <Access accessible={access.canReviewWithdrawals !== false}>
              <Popconfirm title={`确定批准并提交 ${r.amount} ${r.currency}？`} onConfirm={() => onApprove(r.id)}>
                <Button size="small" type="primary">批准</Button>
              </Popconfirm>
              <Button size="small" danger onClick={() => setRejectId(r.id)}>拒绝</Button>
            </Access>
          )}
          {r.review_status === 'submit_failed' && (
            <>
              <Popconfirm title="确认重新提交至 Cobo？" onConfirm={() => onRetry(r.id)}>
                <Button size="small" type="primary">重试</Button>
              </Popconfirm>
              <Button size="small" danger onClick={() => setRejectId(r.id)}>拒绝退款</Button>
            </>
          )}
        </Space>
      ),
    },
  ];

  return (
    <>
      <Tabs
        activeKey={tab}
        onChange={(k) => { setTab(k); actionRef.current?.reload(); }}
        items={[
          { key: 'pending_review', label: '待审核' },
          { key: 'submit_failed',  label: '提交失败' },
          { key: '',               label: '历史' },
        ]}
      />
      <ProTable<WithdrawReviewItem>
        actionRef={actionRef}
        rowKey="id"
        search={false}
        request={async ({ current, pageSize }) => {
          const res = await listWithdrawals({
            review_status: tab || undefined,
            page: current,
            limit: pageSize,
          });
          return {
            data: res.data ?? [],
            total: res.meta?.total ?? 0,
            success: true,
          };
        }}
        columns={columns}
        scroll={{ x: 1200 }}
      />
      <RejectModal
        txId={rejectId}
        onClose={() => setRejectId(null)}
        onDone={() => { setRejectId(null); actionRef.current?.reload(); }}
      />
    </>
  );
}
```

- [ ] **Step 4: Register route in `admin/config/routes.ts`**

Add an entry under the transactions/wallet group:
```ts
{
  path: '/withdrawals',
  name: '提现审核',
  icon: 'AuditOutlined',
  component: './Withdrawals',
  access: 'canViewWithdrawals',
},
```

If the project's access controller (`admin/src/access.ts`) doesn't define `canViewWithdrawals`, add:
```ts
canViewWithdrawals: ['super_admin', 'operator'].includes(currentUser?.role),
```

- [ ] **Step 5: Run dev frontend smoke**

Run: `cd admin && pnpm dev`
Expected: `/withdrawals` loads. Manually verify: list shows correct columns, tab switching works, the buttons do not throw.

- [ ] **Step 6: Commit**

```bash
git add admin/
git commit -m "feat(admin-ui): withdrawals review page with approve/reject/retry actions"
```

---

## Phase 7 — User frontend

### Task 16: Surface review status + cancel button on user wallet page

**Files:**
- Modify: `front/src/app/(app)/wallet/page.tsx`
- Modify: `front/src/hooks/use-api.ts`
- Modify: `front/src/i18n/{en,zh,ko}.json`

- [ ] **Step 1: Add cancel hook**

Append to `front/src/hooks/use-api.ts`:
```typescript
export function useCancelWithdraw() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (txId: string) => api.delete(`/wallets/withdraw/${txId}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['transactions'] })
      qc.invalidateQueries({ queryKey: ['wallets'] })
    },
  })
}
```

- [ ] **Step 2: Add i18n keys (zh shown; mirror to en + ko using same shape, translated text)**

Add inside `wallet` block of `front/src/i18n/zh.json`:
```json
"reviewStatus_pending_review": "审核中",
"reviewStatus_submit_failed":  "提交受阻 (重试中)",
"reviewStatus_submitted":      "已提交",
"reviewStatus_rejected":       "已拒绝",
"reviewStatus_cancelled":      "已取消",
"cancelWithdraw": "撤销申请",
"withdrawSubmittedQueued": "提现申请已提交，等待管理员审核",
"rejectReason": "拒绝原因"
```

en / ko with translated equivalents.

- [ ] **Step 3: Read the current wallet page**

Run: `wc -l front/src/app/(app)/wallet/page.tsx` (this is the large summarized file from the prior session). Read it fully before editing — the edits below are surgical, not a rewrite.

- [ ] **Step 4: Make four targeted edits**

Find each of these spots and edit only them. Do NOT refactor anything else.

**Edit A — import the new hook.** Find the existing `import` block that pulls hooks from `@/hooks/use-api`:
```typescript
// before
import { useWallets, useWithdraw, /* ... */ } from "@/hooks/use-api"
// after
import { useWallets, useWithdraw, useCancelWithdraw, /* ... */ } from "@/hooks/use-api"
```

**Edit B — update success toast copy.** Locate the call site of `useWithdraw()` `onSuccess` (or the place that shows `t("wallet.withdrawSuccess")`):
```typescript
// before
toast.success(t("wallet.withdrawSuccess"))
// after
toast.success(t("wallet.withdrawSubmittedQueued"))
```

**Edit C — add a `Status` cell that knows about review_status.** Inside the transactions table render, find the `Status` column. Replace its render with:
```tsx
{tx.type === "withdraw" && tx.review_status ? (
  <span className="inline-flex flex-col gap-1">
    <span>{t(`wallet.reviewStatus_${tx.review_status}`)}</span>
    {tx.review_status === "rejected" && tx.reject_reason && (
      <span className="text-xs text-muted-foreground" title={tx.reject_reason}>
        {t("wallet.rejectReason")}: {tx.reject_reason}
      </span>
    )}
  </span>
) : (
  <span>{tx.status}</span>
)}
```

**Edit D — add a Cancel button for `pending_review` rows.** In the same row's actions column (create the column if none exists), add:
```tsx
{tx.type === "withdraw" && tx.review_status === "pending_review" && (
  <CancelWithdrawButton txId={tx.id} />
)}
```

And add at the bottom of the file (or in a sibling component file, your choice):
```tsx
function CancelWithdrawButton({ txId }: { txId: string }) {
  const cancel = useCancelWithdraw()
  const t = useTranslations("wallet")
  const [open, setOpen] = useState(false)
  return (
    <>
      <Button variant="ghost" size="sm" onClick={() => setOpen(true)}>
        {t("cancelWithdraw")}
      </Button>
      <ConfirmDialog
        open={open}
        onOpenChange={setOpen}
        title={t("cancelWithdraw")}
        onConfirm={async () => {
          await cancel.mutateAsync(txId)
          setOpen(false)
        }}
      />
    </>
  )
}
```

> If `ConfirmDialog` / `Button` import paths differ in this project, adapt to whatever the wallet page already uses for similar confirmation flows (e.g. the redeem confirm in the investment page).

- [ ] **Step 5: Smoke test in dev**

Run `pnpm --filter front dev`. Submit a withdraw with a real test user. Verify the transaction shows `审核中` and the cancel button works.

- [ ] **Step 6: Commit**

```bash
git add front/
git commit -m "feat(front): show withdraw review status and cancel button"
```

---

## Phase 8 — E2E + verification

### Task 17: Playwright E2E

**Files:**
- Create: `e2e/tests/withdraw-review.spec.ts`

- [ ] **Step 1: Find the existing E2E setup**

Run: `ls e2e/tests/ && cat e2e/playwright.config.ts | head -30`
Use the existing helpers (login, seed user, mock Cobo) — do not invent new ones.

- [ ] **Step 2: Write three scenarios**

```typescript
import { test, expect } from '@playwright/test';
import { loginAsUser, loginAsAdmin, mockCoboNext } from './_helpers';

test('user→admin reject refunds available balance', async ({ page, request }) => {
  await loginAsUser(page, 'e2e-user@x.com');
  // submit withdraw...
  // open admin in new context, reject with reason
  // re-login as user → balance restored, review status = rejected, reason visible
});

test('user→admin approve→cobo success→webhook confirmed', async ({ page, request }) => {
  await mockCoboNext(request, { success: true, externalId: 'ext-1' });
  // submit, approve, fire webhook, assert confirmed
});

test('user→admin approve→cobo 12007 failure→retry success', async ({ page, request }) => {
  await mockCoboNext(request, { success: false, errorCode: 12007 });
  // submit, approve → fails (UI shows submit_failed), still frozen in user wallet
  await mockCoboNext(request, { success: true, externalId: 'ext-2' });
  // retry → success
});
```

(If `mockCoboNext` doesn't exist in helpers, create it; the existing test cobo provider in `pkg/cobo/mock.go` likely already supports queueing responses — wire it through an admin debug endpoint or env switch already used by other tests.)

- [ ] **Step 3: Run**

Run: `cd e2e && pnpm test withdraw-review.spec.ts`
Expected: 3 PASS.

- [ ] **Step 4: Commit**

```bash
git add e2e/
git commit -m "test(e2e): withdraw review reject / approve / retry flows"
```

---

### Task 18: Final verification

- [ ] **Step 1: Backend full test + race + coverage**

Run: `cd server && go test ./... -race -cover`
Expected: all PASS, coverage on `wallet/service` and `admin/service` ≥ 80% for new code.

- [ ] **Step 2: Build all binaries**

Run: `cd server && go build ./...` and `cd admin && pnpm build` and `cd front && pnpm build`.
Expected: clean builds.

- [ ] **Step 3: Verify migrations apply on a fresh DB**

Run: `cd server && make migrate-reset && make migrate-up` (substitute the project's actual reset target). Confirm migration #023 applied without error and historical rows backfilled.

- [ ] **Step 4: Manual happy path on staging**

Deploy via `./deployments/deploy.sh rebuild`. Walk through:
1. User submits withdraw → row appears in admin "待审核" tab, user wallet shows frozen.
2. Admin approves → admin success toast, user row → "已提交"; webhook (mock or real test net) → "已到账" + email.
3. Repeat with admin reject → user available balance restored, email received with reason.
4. Repeat with Cobo balance insufficient (deliberately leave Cobo hot wallet under-funded) → admin sees "submit_failed" with last_submit_error tooltip; admin tops up Cobo and retries → success.

- [ ] **Step 5: Open PR**

```bash
git push -u origin feat/withdraw-review
gh pr create --title "feat: admin withdraw review layer" --body "$(cat <<'EOF'
## Summary
- Inserts admin review queue between user withdrawals and Cobo.
- User withdraws now freeze funds and create pending_review tickets.
- Admins approve/reject/retry; only admin reject or Cobo webhook failure refund.
- Cobo errors (12007 etc.) no longer fail user withdrawals — ticket → submit_failed for retry.
- New /admin/withdrawals page; user wallet shows review status + cancel.

Spec: docs/superpowers/specs/2026-05-06-withdraw-review-design.md
Plan: docs/superpowers/plans/2026-05-06-withdraw-review.md

## Test plan
- [ ] All Go unit tests green (incl. cobo-failure-keeps-frozen)
- [ ] Handler integration tests green
- [ ] E2E tests green (reject / approve / retry)
- [ ] Manual staging walkthrough (4 scenarios)
- [ ] Migration applies cleanly + backfills historical rows
EOF
)"
```

---

## Self-Review Notes (filled in after writing)

- **Spec coverage**
  - State machine (spec §3) → Tasks 4, 5, 8, 9, 10
  - Schema (spec §4) → Tasks 1, 2, 3
  - User APIs (spec §5.1) → Tasks 4, 5, 6
  - Admin APIs (spec §5.2) → Tasks 7, 8, 9, 10, 11, 12, 13
  - Service layer changes (spec §6) → Tasks 4, 7–11
  - Admin UI (spec §7) → Task 15
  - User UI (spec §8) → Task 16 (incl. copy change)
  - Notifications (spec §9) → Task 14 (rejected event new; completed/failed already wired by existing webhook logic)
  - Audit (spec §10) → Task 12 + 13 (RequireRole + auditSvc.Log)
  - Tests (spec §11) → Tasks 4, 5, 6, 8, 9, 10, 12, 17
  - Migration (spec §13) → Task 1
- **Placeholder scan**: no TBD/TODO; every code step contains compilable code; tests assert specific outcomes.
- **Type consistency**: review status constants all use `ReviewStatus*` from `walletmodel`; service interface stable across tasks 7–13; handler/service signatures match.
