# Trading Product Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a second investment product (trading strategy) alongside the existing arbitrage product. Users get tagged by admins; new investments route to the matching product pool. The feature is invisible on the user frontend (same "invest" button) but everything behind it is product-aware.

**Architecture:** Add `users.investment_type` (default `arbitrage`) + `product_type` columns on shared tables (`investments` / `user_trades` / `settlements`). Trading fund-level trades live in a NEW physically-separate `trading_trades` table. The existing `SettlementJob` loops over both product types daily; data sources switch by `product_type`. Wallet earnings are split into two columns under the hood (`earnings_arbitrage` + `earnings_trading`) but aggregated as a single number on the user frontend.

**Tech Stack:** Go 1.22+ / Gin / GORM / PostgreSQL (migrate v4 via Docker) / Next.js 19 frontend / Ant Design Pro admin / Vitest / Playwright (deferred — same caveat as the withdraw-review plan).

**Reference spec:** `docs/superpowers/specs/2026-05-06-trading-product-design.md`

---

## File Structure

### Backend (Go)

| Path | Action | Responsibility |
|---|---|---|
| `server/migrations/000024_trading_product.up.sql` | create | All schema changes (new tables + ALTERs + backfill + drop earnings) |
| `server/migrations/000024_trading_product.down.sql` | create | Reverse |
| `server/internal/modules/auth/model/user.go` | modify | Add `InvestmentType string` + constants |
| `server/internal/modules/investment/model/investment.go` | modify | Add `ProductType string` + constants |
| `server/internal/modules/investment/repository/investment_repo.go` | modify | Add `FindAllActiveBeforeDateByProduct` and product-aware list method |
| `server/internal/modules/investment/service/investment_service.go` | modify | Read `users.investment_type` on Create; accept product_type filter on List |
| `server/internal/modules/investment/dto/{request,response}.go` | modify | Surface product_type in responses; accept filter param |
| `server/internal/modules/tradelog/model/user_trade.go` | modify | Add `ProductType string` |
| `server/internal/modules/tradelog/repository/user_trade_repo.go` | modify | Filter by product_type |
| `server/internal/modules/settlement/model/settlement.go` | modify | Add `ProductType string` |
| `server/internal/modules/settlement/repository/settlement_repo.go` | modify | Filter by product_type |
| `server/internal/modules/settlement/service/settlement_service.go` | modify | Accept product_type filter on List |
| `server/internal/modules/wallet/model/wallet.go` | modify | Drop `Earnings`, add `EarningsArbitrage` + `EarningsTrading`; `TotalBalance()` sums them |
| `server/internal/modules/wallet/repository/wallet_repo.go` | modify | `AddEarnings` becomes product-aware; `ClaimEarnings` zeroes both |
| `server/internal/modules/wallet/service/wallet_service.go` | modify | Aggregated earnings in DTO; ClaimEarnings sums both |
| `server/internal/modules/trading_tradelog/model/trade.go` | create | `TradingTrade` struct (TableName=`trading_trades`) |
| `server/internal/modules/trading_tradelog/repository/trade_repo.go` | create | Same shape as `tradelog/repository/trade_repo.go` |
| `server/internal/modules/admin/model/user_product_change_log.go` | create | Audit row model (TableName=`user_product_change_logs`) |
| `server/internal/modules/admin/repository/user_product_change_log_repo.go` | create | Create + List by user |
| `server/internal/modules/admin/service/user_product_service.go` | create | Single + bulk tagging logic |
| `server/internal/modules/admin/service/trading_trade_service.go` | create | Mirror `trade_service.go` but writes to trading_trades; reuses `parseTradeRow` (export it or duplicate) |
| `server/internal/modules/admin/handler/user_product_handler.go` | create | Single + bulk tag endpoints + history |
| `server/internal/modules/admin/handler/trading_trade_handler.go` | create | List/import/template/stats/delete |
| `server/internal/modules/admin/dto/{request,response}.go` | modify | Add tagging DTOs + trading-trade DTOs (or reuse) + investment_type field on user responses |
| `server/internal/modules/admin/routes.go` | modify | Register all new routes (super_admin/operator) |
| `server/internal/modules/admin/module.go` | modify | Wire new services + handlers |
| `server/internal/worker/settlement_job.go` | modify | Refactor to per-product loop; accept TradingTradeRepository |
| `server/internal/worker/worker.go` | (no change) | |
| `server/cmd/worker/main.go` (or wherever worker is built) | modify | Construct + inject the new TradingTradeRepository |
| `server/internal/app/app.go` | modify | Construct trading repo; thread through wherever investment service / wallet service / settlement job are wired |

### Admin frontend

| Path | Action | Responsibility |
|---|---|---|
| `admin/src/pages/UserDetail/InvestmentTypeCard.tsx` | create | Card on user detail page with current type + change button + history |
| `admin/src/pages/UserDetail/index.tsx` | modify | Embed the card; pass userId |
| `admin/src/pages/Users/index.tsx` | modify | Add `投资类型` column + filter + bulk-tag button |
| `admin/src/pages/Users/BulkTagModal.tsx` | create | Bulk tag modal (target type + reason) |
| `admin/src/pages/TradingTrades/index.tsx` | create | List + stats |
| `admin/src/pages/TradingTrades/Import.tsx` | create | Excel import page |
| `admin/src/services/api.ts` | modify | New helpers: tagging + bulk + history + trading-trades CRUD/import/template |
| `admin/config/routes.ts` | modify | Register `/trading-trades` and `/trading-trades/import` |
| `admin/src/access.ts` | modify | (nothing new — reuse super_admin/operator) |

### User frontend

| Path | Action | Responsibility |
|---|---|---|
| `front/src/types/api.ts` | modify | Add `product_type` on Investment / Settlement; `investment_type` on User |
| `front/src/hooks/use-api.ts` | modify | `useInvestments(productType?)`, `useSettlements(productType?)`; user object exposes investment_type |
| `front/src/app/(app)/investments/page.tsx` | modify | Show tab when user has both products historically |
| `front/src/app/(app)/reports/page.tsx` (or settlements equivalent) | modify | Same tab pattern |
| `front/src/app/(app)/wallet/page.tsx` | (no change to UI) | Backend already returns aggregated earnings — verify shape |
| `front/src/i18n/{en,zh,ko}.json` | modify | Tab labels: "当前 / 历史" |

---

## Phase 1 — Database

### Task 1: Migration #024

**Files:**
- Create: `server/migrations/000024_trading_product.up.sql`
- Create: `server/migrations/000024_trading_product.down.sql`

- [ ] **Step 1: Write up migration**

```sql
-- 1. New: per-user product type marker
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS investment_type varchar(20) NOT NULL DEFAULT 'arbitrage';

-- 2. New: product_type on user-facing tables
ALTER TABLE investments
  ADD COLUMN IF NOT EXISTS product_type varchar(20) NOT NULL DEFAULT 'arbitrage';
CREATE INDEX IF NOT EXISTS idx_investments_product ON investments(product_type, status);

ALTER TABLE user_trades
  ADD COLUMN IF NOT EXISTS product_type varchar(20) NOT NULL DEFAULT 'arbitrage';
CREATE INDEX IF NOT EXISTS idx_user_trades_product ON user_trades(product_type);

ALTER TABLE settlements
  ADD COLUMN IF NOT EXISTS product_type varchar(20) NOT NULL DEFAULT 'arbitrage';
CREATE INDEX IF NOT EXISTS idx_settlements_product ON settlements(product_type, period);

-- 3. Wallet earnings split (backfill BEFORE dropping the old column)
ALTER TABLE wallets
  ADD COLUMN IF NOT EXISTS earnings_arbitrage decimal(28,18) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS earnings_trading   decimal(28,18) NOT NULL DEFAULT 0;
UPDATE wallets SET earnings_arbitrage = COALESCE(earnings, 0) WHERE earnings_arbitrage = 0;
ALTER TABLE wallets DROP COLUMN IF EXISTS earnings;

-- 4. Audit table for tagging changes
CREATE TABLE IF NOT EXISTS user_product_change_logs (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     uuid NOT NULL,
  from_type   varchar(20) NOT NULL,
  to_type     varchar(20) NOT NULL,
  admin_id    uuid NOT NULL,
  admin_email varchar(255) NOT NULL,
  reason      varchar(500) NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_upcl_user ON user_product_change_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_upcl_created ON user_product_change_logs(created_at);

-- 5. Trading fund trades (mirrors trades schema)
CREATE TABLE IF NOT EXISTS trading_trades (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  pair          varchar(20) NOT NULL,
  buy_exchange  varchar(20),
  sell_exchange varchar(20),
  buy_price     decimal(28,8),
  sell_price    decimal(28,8),
  amount        decimal(28,18) NOT NULL,
  premium_pct   decimal(8,4) DEFAULT 0,
  pnl           decimal(28,18) NOT NULL,
  fee           decimal(28,18) DEFAULT 0,
  source        varchar(20) NOT NULL DEFAULT 'import',
  executed_at   timestamptz NOT NULL,
  created_at    timestamptz NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_trading_trades_executed_at ON trading_trades(executed_at);
```

- [ ] **Step 2: Write down migration**

```sql
DROP TABLE IF EXISTS trading_trades;
DROP TABLE IF EXISTS user_product_change_logs;

ALTER TABLE wallets ADD COLUMN IF NOT EXISTS earnings decimal(28,18) NOT NULL DEFAULT 0;
UPDATE wallets SET earnings = earnings_arbitrage + earnings_trading;
ALTER TABLE wallets DROP COLUMN IF EXISTS earnings_arbitrage;
ALTER TABLE wallets DROP COLUMN IF EXISTS earnings_trading;

DROP INDEX IF EXISTS idx_settlements_product;
ALTER TABLE settlements DROP COLUMN IF EXISTS product_type;

DROP INDEX IF EXISTS idx_user_trades_product;
ALTER TABLE user_trades DROP COLUMN IF EXISTS product_type;

DROP INDEX IF EXISTS idx_investments_product;
ALTER TABLE investments DROP COLUMN IF EXISTS product_type;

ALTER TABLE users DROP COLUMN IF EXISTS investment_type;
```

- [ ] **Step 3: Commit (no apply — applied during deploy)**

```bash
git add server/migrations/000024_trading_product.up.sql server/migrations/000024_trading_product.down.sql
git commit -m "feat(db): trading product schema (migration 024)"
```

---

## Phase 2 — Models & repositories

### Task 2: User model gets `InvestmentType`

**Files:**
- Modify: `server/internal/modules/auth/model/user.go`

- [ ] **Step 1: Edit user struct**

Add field above `CreatedAt`:
```go
	InvestmentType string    `gorm:"type:varchar(20);not null;default:arbitrage" json:"investment_type"`
```

Append constants below the struct (or in a new `package model` const block):
```go
const (
	InvestmentTypeArbitrage = "arbitrage"
	InvestmentTypeTrading   = "trading"
)
```

- [ ] **Step 2: Build**

```bash
cd /Users/johnny/Work/soveregin/server && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add server/internal/modules/auth/model/user.go
git commit -m "feat(user): add investment_type field"
```

---

### Task 3: Investment model + repo product_type

**Files:**
- Modify: `server/internal/modules/investment/model/investment.go`
- Modify: `server/internal/modules/investment/repository/investment_repo.go`

- [ ] **Step 1: Add field + constants**

In `model/investment.go`, insert above `Status`:
```go
	ProductType string `gorm:"type:varchar(20);not null;default:arbitrage" json:"product_type"`
```

And append at the bottom of the const block (after the existing `InvestStatus*` constants):
```go
const (
	ProductTypeArbitrage = "arbitrage"
	ProductTypeTrading   = "trading"
)
```

- [ ] **Step 2: Add product-aware repo methods**

In `repository/investment_repo.go`, append to the `InvestmentRepository` interface:
```go
	FindAllActiveBeforeDateByProduct(ctx context.Context, before time.Time, productType string) ([]model.Investment, error)
	FindByUserIDAndProduct(ctx context.Context, userID, productType string) ([]model.Investment, error)
```

And append the implementations:
```go
func (r *investmentRepository) FindAllActiveBeforeDateByProduct(ctx context.Context, before time.Time, productType string) ([]model.Investment, error) {
	var invs []model.Investment
	err := r.db.WithContext(ctx).
		Where("status = ? AND start_date < ? AND product_type = ?", model.InvestStatusActive, before, productType).
		Find(&invs).Error
	return invs, err
}

func (r *investmentRepository) FindByUserIDAndProduct(ctx context.Context, userID, productType string) ([]model.Investment, error) {
	var invs []model.Investment
	q := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if productType != "" && productType != "all" {
		q = q.Where("product_type = ?", productType)
	}
	err := q.Order("created_at DESC").Find(&invs).Error
	return invs, err
}
```

- [ ] **Step 3: Build + commit**

```bash
cd /Users/johnny/Work/soveregin/server && go build ./...
git add server/internal/modules/investment/
git commit -m "feat(investment): add product_type field and product-aware queries"
```

---

### Task 4: UserTrade + Settlement models get product_type

**Files:**
- Modify: `server/internal/modules/tradelog/model/user_trade.go`
- Modify: `server/internal/modules/settlement/model/settlement.go`

- [ ] **Step 1: Edit UserTrade**

Insert into the struct (above `ExecutedAt`):
```go
	ProductType string `gorm:"type:varchar(20);not null;default:arbitrage" json:"product_type"`
```

- [ ] **Step 2: Edit Settlement**

Insert into the struct (above `SettledAt`):
```go
	ProductType string `gorm:"type:varchar(20);not null;default:arbitrage" json:"product_type"`
```

- [ ] **Step 3: Build + commit**

```bash
cd /Users/johnny/Work/soveregin/server && go build ./...
git add server/internal/modules/tradelog/model/user_trade.go server/internal/modules/settlement/model/settlement.go
git commit -m "feat(settlement): add product_type field on UserTrade and Settlement"
```

---

### Task 5: Wallet model + repo earnings split

**Files:**
- Modify: `server/internal/modules/wallet/model/wallet.go`
- Modify: `server/internal/modules/wallet/repository/wallet_repo.go`

- [ ] **Step 1: Replace Wallet struct**

Edit `model/wallet.go`:
```go
package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Wallet struct {
	ID                string          `gorm:"type:uuid;primaryKey" json:"id"`
	UserID            string          `gorm:"type:uuid;index;not null" json:"user_id"`
	Currency          string          `gorm:"type:varchar(10);not null" json:"currency"`
	Available         decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"available"`
	InOperation       decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"in_operation"`
	Frozen            decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"frozen"`
	EarningsArbitrage decimal.Decimal `gorm:"type:decimal(28,18);not null;default:0" json:"earnings_arbitrage"`
	EarningsTrading   decimal.Decimal `gorm:"type:decimal(28,18);not null;default:0" json:"earnings_trading"`
	CreatedAt         time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (w *Wallet) BeforeCreate(_ *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

func (w *Wallet) Earnings() decimal.Decimal {
	return w.EarningsArbitrage.Add(w.EarningsTrading)
}

func (w *Wallet) TotalBalance() decimal.Decimal {
	return w.Available.Add(w.InOperation).Add(w.Frozen).Add(w.Earnings())
}
```

- [ ] **Step 2: Replace AddEarnings + ClaimEarnings in repo**

Edit `repository/wallet_repo.go`. Replace the interface signatures:
```go
	AddEarnings(ctx context.Context, id string, amount decimal.Decimal, productType string) error
	ClaimEarnings(ctx context.Context, id string) error
```

Replace the implementations:
```go
func (r *walletRepository) AddEarnings(ctx context.Context, id string, amount decimal.Decimal, productType string) error {
	col := "earnings_arbitrage"
	if productType == "trading" {
		col = "earnings_trading"
	}
	return r.db.WithContext(ctx).
		Model(&model.Wallet{}).
		Where("id = ?", id).
		Update(col, gorm.Expr(col+" + ?", amount)).Error
}

func (r *walletRepository) ClaimEarnings(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&model.Wallet{}).
		Where("id = ? AND (earnings_arbitrage > 0 OR earnings_trading > 0)", id).
		Updates(map[string]any{
			"available":           gorm.Expr("available + earnings_arbitrage + earnings_trading"),
			"earnings_arbitrage":  0,
			"earnings_trading":    0,
		}).Error
}
```

- [ ] **Step 3: Find call sites**

```bash
grep -rn "AddEarnings\|\.Earnings\b" server/internal/ --include="*.go"
```

For each caller:
- `worker/settlement_job.go` — Task 7 will rewrite this with product_type
- `wallet/service/wallet_service.go` — `GetWallets` uses `w.Earnings` field. The model now exposes `Earnings()` method; update reads to call `w.Earnings()` instead. The DTO `WalletResponse.Earnings` should keep the aggregated value so the frontend doesn't see the split.

Patch `wallet_service.go`'s `GetWallets`:
```go
		resp = append(resp, dto.WalletResponse{
			ID:          w.ID,
			Currency:    w.Currency,
			Available:   w.Available,
			InOperation: w.InOperation,
			Frozen:      w.Frozen,
			Earnings:    w.Earnings(),       // aggregated
			Total:       w.TotalBalance(),
		})
```

Other reads of `wallet.Earnings` — replace with `wallet.Earnings()` (method call).

- [ ] **Step 4: Build (expect failures from other call sites)**

```bash
cd /Users/johnny/Work/soveregin/server && go build ./... 2>&1 | head -30
```

Fix any callers that pass the old 2-arg `AddEarnings` signature — update to 3 args. The settlement job and any test file may call this; minimum patch is to add a third arg `model.ProductTypeArbitrage` until Task 7 plumbs the real type.

- [ ] **Step 5: Tests + commit**

```bash
cd /Users/johnny/Work/soveregin/server && go test ./internal/modules/wallet/... -v
git add server/internal/modules/wallet/ server/internal/worker/
git commit -m "feat(wallet): split earnings into arbitrage + trading columns; expose aggregated Earnings()"
```

---

### Task 6: New TradingTrade model + repo

**Files:**
- Create: `server/internal/modules/trading_tradelog/model/trade.go`
- Create: `server/internal/modules/trading_tradelog/repository/trade_repo.go`

- [ ] **Step 1: Create model**

```go
package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TradingTrade struct {
	ID           string          `gorm:"type:uuid;primaryKey" json:"id"`
	Pair         string          `gorm:"type:varchar(20);not null" json:"pair"`
	BuyExchange  string          `gorm:"type:varchar(20)" json:"buy_exchange"`
	SellExchange string          `gorm:"type:varchar(20)" json:"sell_exchange"`
	BuyPrice     decimal.Decimal `gorm:"type:decimal(28,8)" json:"buy_price"`
	SellPrice    decimal.Decimal `gorm:"type:decimal(28,8)" json:"sell_price"`
	Amount       decimal.Decimal `gorm:"type:decimal(28,18);not null" json:"amount"`
	PremiumPct   decimal.Decimal `gorm:"type:decimal(8,4);default:0" json:"premium_pct"`
	PnL          decimal.Decimal `gorm:"column:pnl;type:decimal(28,18);not null" json:"pnl"`
	Fee          decimal.Decimal `gorm:"type:decimal(28,18);default:0" json:"fee"`
	Source       string          `gorm:"type:varchar(20);default:import" json:"source"`
	ExecutedAt   time.Time       `gorm:"not null" json:"executed_at"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
}

func (t *TradingTrade) BeforeCreate(_ *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

func (TradingTrade) TableName() string {
	return "trading_trades"
}
```

- [ ] **Step 2: Create repo (mirror tradelog pattern)**

```go
package repository

import (
	"context"
	"time"

	"github.com/sovereign-fund/sovereign/internal/modules/trading_tradelog/model"
	tradelogrepo "github.com/sovereign-fund/sovereign/internal/modules/tradelog/repository"
	"gorm.io/gorm"
)

type TradingTradeRepository interface {
	Create(ctx context.Context, t *model.TradingTrade) error
	FindByID(ctx context.Context, id string) (*model.TradingTrade, error)
	FindAll(ctx context.Context, filters tradelogrepo.TradeFilters, limit, offset int) ([]model.TradingTrade, int64, error)
	SummarizeByPeriod(ctx context.Context, from, to time.Time) (*tradelogrepo.TradeSummaryResult, error)
	SummarizeAll(ctx context.Context, filters tradelogrepo.TradeFilters) (*tradelogrepo.TradeSummaryResult, error)
	FindByPeriod(ctx context.Context, from, to time.Time) ([]model.TradingTrade, error)
	Delete(ctx context.Context, id string) error
	BatchCreate(ctx context.Context, trades []model.TradingTrade) error
}

type tradingTradeRepository struct {
	db *gorm.DB
}

func NewTradingTradeRepository(db *gorm.DB) TradingTradeRepository {
	return &tradingTradeRepository{db: db}
}

func (r *tradingTradeRepository) Create(ctx context.Context, t *model.TradingTrade) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *tradingTradeRepository) BatchCreate(ctx context.Context, trades []model.TradingTrade) error {
	if len(trades) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(trades, 100).Error
}

func (r *tradingTradeRepository) FindByID(ctx context.Context, id string) (*model.TradingTrade, error) {
	var t model.TradingTrade
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *tradingTradeRepository) FindAll(ctx context.Context, filters tradelogrepo.TradeFilters, limit, offset int) ([]model.TradingTrade, int64, error) {
	var trades []model.TradingTrade
	var total int64
	q := r.db.WithContext(ctx).Model(&model.TradingTrade{})
	if filters.Pair != "" {
		q = q.Where("pair = ?", filters.Pair)
	}
	if !filters.From.IsZero() {
		q = q.Where("executed_at >= ?", filters.From)
	}
	if !filters.To.IsZero() {
		q = q.Where("executed_at <= ?", filters.To)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("executed_at DESC").Limit(limit).Offset(offset).Find(&trades).Error
	return trades, total, err
}

func (r *tradingTradeRepository) SummarizeAll(ctx context.Context, filters tradelogrepo.TradeFilters) (*tradelogrepo.TradeSummaryResult, error) {
	q := r.db.WithContext(ctx).Model(&model.TradingTrade{})
	if !filters.From.IsZero() {
		q = q.Where("executed_at >= ?", filters.From)
	}
	if !filters.To.IsZero() {
		q = q.Where("executed_at <= ?", filters.To)
	}
	var result tradelogrepo.TradeSummaryResult
	err := q.Select(`
		COUNT(*) as total_trades,
		COALESCE(SUM(pnl), 0) as total_pnl,
		COALESCE(AVG(premium_pct), 0) as avg_premium,
		COUNT(CASE WHEN pnl > 0 THEN 1 END) as win_count
	`).Row().Scan(&result.TotalTrades, &result.TotalPnL, &result.AvgPremium, &result.WinCount)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *tradingTradeRepository) SummarizeByPeriod(ctx context.Context, from, to time.Time) (*tradelogrepo.TradeSummaryResult, error) {
	return r.SummarizeAll(ctx, tradelogrepo.TradeFilters{From: from, To: to})
}

func (r *tradingTradeRepository) FindByPeriod(ctx context.Context, from, to time.Time) ([]model.TradingTrade, error) {
	var trades []model.TradingTrade
	err := r.db.WithContext(ctx).
		Where("executed_at >= ? AND executed_at < ?", from, to).
		Order("executed_at ASC").
		Find(&trades).Error
	return trades, err
}

func (r *tradingTradeRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.TradingTrade{}).Error
}
```

- [ ] **Step 3: Build + commit**

```bash
cd /Users/johnny/Work/soveregin/server && go build ./...
git add server/internal/modules/trading_tradelog/
git commit -m "feat(trading): TradingTrade model and repository (mirrors tradelog)"
```

---

## Phase 3 — Investment service product-type wiring (TDD)

### Task 7: Investment.Create reads user.investment_type

**Files:**
- Modify: `server/internal/modules/investment/service/investment_service.go`
- Modify: `server/internal/modules/investment/service/investment_service_test.go`

- [ ] **Step 1: Inspect existing test setup**

```bash
cat /Users/johnny/Work/soveregin/server/internal/modules/investment/service/investment_service_test.go | head -80
```

The test file already has `stubInvestmentRepository`. Mirror its style (stdlib testing, no testify).

- [ ] **Step 2: Add user dependency to service**

In `investment_service.go`, the service struct currently doesn't have a user repo. Add it:

```go
type investmentService struct {
	invRepo    repository.InvestmentRepository
	walletRepo walletRepo.WalletRepository
	userRepo   userRepo.UserRepository  // NEW
	eventBus   events.Bus
	logger     *slog.Logger
}
```

Add the import: `userRepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"`.

Update `NewInvestmentService`:
```go
func NewInvestmentService(
	invRepo repository.InvestmentRepository,
	wr walletRepo.WalletRepository,
	ur userRepo.UserRepository,
	bus events.Bus,
	logger *slog.Logger,
) InvestmentService {
	return &investmentService{
		invRepo:    invRepo,
		walletRepo: wr,
		userRepo:   ur,
		eventBus:   bus,
		logger:     logger,
	}
}
```

- [ ] **Step 3: Failing test**

In `investment_service_test.go` add a stub for the user repo (or extend existing stubs). Then add:

```go
func TestCreate_AssignsProductTypeFromUser(t *testing.T) {
	ctx := context.Background()
	invRepo := &stubInvestmentRepository{}
	wRepo := &stubWalletRepo{wallet: &walletmodel.Wallet{ID: "w1", UserID: "u1", Currency: "USDT", Available: decimal.NewFromInt(500)}}
	uRepo := &stubUserRepo{users: map[string]*usermodel.User{
		"u1": {ID: "u1", InvestmentType: "trading"},
	}}
	bus := &spyBus{}
	svc := NewInvestmentService(invRepo, wRepo, uRepo, bus, slog.Default())

	resp, err := svc.Create(ctx, "u1", dto.CreateInvestmentRequest{Amount: "200", Currency: "USDT"})
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if resp == nil || resp.ProductType != "trading" {
		t.Errorf("response.ProductType = %v, want trading", resp)
	}
	if len(invRepo.created) != 1 || invRepo.created[0].ProductType != "trading" {
		t.Errorf("persisted product_type = %v, want trading", invRepo.created)
	}
}

func TestCreate_DefaultsToArbitrageWhenUserHasNoType(t *testing.T) {
	ctx := context.Background()
	invRepo := &stubInvestmentRepository{}
	wRepo := &stubWalletRepo{wallet: &walletmodel.Wallet{ID: "w1", UserID: "u1", Currency: "USDT", Available: decimal.NewFromInt(500)}}
	uRepo := &stubUserRepo{users: map[string]*usermodel.User{
		"u1": {ID: "u1", InvestmentType: ""}, // legacy data
	}}
	svc := NewInvestmentService(invRepo, wRepo, uRepo, &spyBus{}, slog.Default())

	resp, _ := svc.Create(ctx, "u1", dto.CreateInvestmentRequest{Amount: "200", Currency: "USDT"})
	if resp.ProductType != "arbitrage" {
		t.Errorf("response.ProductType = %s, want arbitrage", resp.ProductType)
	}
}
```

If `stubUserRepo` doesn't yet exist in the test file, define it:
```go
type stubUserRepo struct {
	users map[string]*usermodel.User
}
func (s *stubUserRepo) FindByID(ctx context.Context, id string) (*usermodel.User, error) {
	if u, ok := s.users[id]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepo) FindByEmail(context.Context, string) (*usermodel.User, error) { panic("unused") }
func (s *stubUserRepo) FindByGoogleID(context.Context, string) (*usermodel.User, error) { panic("unused") }
func (s *stubUserRepo) Create(context.Context, *usermodel.User) error                  { panic("unused") }
func (s *stubUserRepo) Update(context.Context, *usermodel.User) error                  { panic("unused") }
func (s *stubUserRepo) ExistsByEmail(context.Context, string) (bool, error)            { panic("unused") }
```

Add imports: `usermodel "github.com/sovereign-fund/sovereign/internal/modules/auth/model"`, `"gorm.io/gorm"`.

- [ ] **Step 4: Run, expect FAIL**

```bash
cd /Users/johnny/Work/soveregin/server && go test ./internal/modules/investment/service/ -run TestCreate_AssignsProductTypeFromUser -v
```

- [ ] **Step 5: Implement**

In `investmentService.Create`, after the balance checks but before creating the Investment, fetch the user:
```go
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, err)
	}
	productType := user.InvestmentType
	if productType == "" {
		productType = model.ProductTypeArbitrage
	}
```

Set on the Investment:
```go
	inv := &model.Investment{
		UserID:      userID,
		Amount:      amount,
		Currency:    currency,
		Status:      model.InvestStatusActive,
		ProductType: productType,
	}
```

- [ ] **Step 6: Surface in response DTO**

Edit `dto/response.go`. Add `ProductType string `json:"product_type"`` to `InvestmentResponse`. In service code where the response is built, set `ProductType: inv.ProductType`.

- [ ] **Step 7: Update other call sites**

`go build ./...` will fail because `NewInvestmentService` signature changed. Find caller:
```bash
grep -rn "NewInvestmentService" server/
```

In `server/internal/app/app.go` (or wherever it's wired), pass the user repo. The user repo is already constructed for the auth module — reuse it or create a fresh wrapper (`userRepo.NewUserRepository(db)`).

- [ ] **Step 8: Run + commit**

```bash
cd /Users/johnny/Work/soveregin/server && go test ./internal/modules/investment/... -v && go build ./...
git add server/internal/modules/investment/ server/internal/app/
git commit -m "feat(investment): assign product_type from users.investment_type on create"
```

---

### Task 8: Investment list + filtering by product

**Files:**
- Modify: `server/internal/modules/investment/dto/request.go`
- Modify: `server/internal/modules/investment/service/investment_service.go`
- Modify: `server/internal/modules/investment/handler/investment_handler.go`

- [ ] **Step 1: Failing test**

Append to `investment_service_test.go`:
```go
func TestGetAll_FiltersByProductType(t *testing.T) {
	ctx := context.Background()
	invRepo := &stubInvestmentRepository{
		byUserAndProduct: map[string][]model.Investment{
			"u1|trading": {{ID: "i1", ProductType: "trading"}},
			"u1|arbitrage": {{ID: "i2", ProductType: "arbitrage"}},
		},
	}
	uRepo := &stubUserRepo{users: map[string]*usermodel.User{"u1": {ID: "u1", InvestmentType: "trading"}}}
	wRepo := &stubWalletRepo{}
	svc := NewInvestmentService(invRepo, wRepo, uRepo, &spyBus{}, slog.Default())

	resp, err := svc.GetAll(ctx, "u1", "trading")
	if err != nil {
		t.Fatalf("GetAll error: %v", err)
	}
	if len(resp.Investments) != 1 || resp.Investments[0].ProductType != "trading" {
		t.Errorf("want only trading rows, got %v", resp.Investments)
	}
}
```

(Extend `stubInvestmentRepository` to implement `FindByUserIDAndProduct` — when `byUserAndProduct` is set, return rows for the `user|product` key.)

- [ ] **Step 2: Update service signature**

Change interface:
```go
	GetAll(ctx context.Context, userID, productType string) (*dto.InvestmentListResponse, error)
```

Implementation: pass `productType` to `invRepo.FindByUserIDAndProduct(ctx, userID, productType)`. If `productType == ""`, default to user's current `investment_type`.

- [ ] **Step 3: Handler accepts query param**

In `handler/investment_handler.go`, in the `GetAll` handler:
```go
	productType := c.Query("product_type") // "", "arbitrage", "trading", "all"
	resp, err := h.svc.GetAll(c.Request.Context(), userID, productType)
```

- [ ] **Step 4: Build + tests + commit**

```bash
cd /Users/johnny/Work/soveregin/server && go test ./internal/modules/investment/... -v && go build ./...
git add server/internal/modules/investment/
git commit -m "feat(investment): filter list by product_type with default-to-user-type"
```

---

### Task 9: Settlement service filter by product_type

**Files:**
- Modify: `server/internal/modules/settlement/service/settlement_service.go`
- Modify: `server/internal/modules/settlement/repository/settlement_repo.go`
- Modify: `server/internal/modules/settlement/dto/response.go`
- Modify: `server/internal/modules/settlement/handler/settlement_handler.go`

- [ ] **Step 1: Add repo method**

```go
	FindByUserIDAndProduct(ctx context.Context, userID, productType string, limit, offset int) ([]model.Settlement, int64, error)
```

Implementation pattern same as investment repo's product-aware query.

- [ ] **Step 2: Service + handler accept param**

Service signature change:
```go
	List(ctx context.Context, userID, productType string, page, perPage int) ([]dto.SettlementResponse, int64, error)
```

Handler reads `c.Query("product_type")`.

- [ ] **Step 3: Add field to settlement response DTO**

Add `ProductType string `json:"product_type"`` to `SettlementResponse`.

- [ ] **Step 4: Build + tests + commit**

```bash
cd /Users/johnny/Work/soveregin/server && go test ./internal/modules/settlement/... -v && go build ./...
git add server/internal/modules/settlement/
git commit -m "feat(settlement): filter by product_type and surface in DTO"
```

---

## Phase 4 — Settlement worker double loop (TDD)

### Task 10: Settlement Job per-product loop

**Files:**
- Modify: `server/internal/worker/settlement_job.go`
- Create or extend: `server/internal/worker/settlement_job_test.go`

- [ ] **Step 1: Inspect existing test (if any)**

```bash
ls server/internal/worker/ | grep -i test
```

If no test exists, this task creates one for the new behavior.

- [ ] **Step 2: Refactor SettlementJob constructor**

Add a `tradingTradeRepo` field of type `tradingrepo.TradingTradeRepository`. Update `NewSettlementJob` and `NewSettlementJobFromDB` accordingly:

```go
func NewSettlementJob(
	ir investRepo.InvestmentRepository,
	tr tradeRepo.TradeRepository,
	ttr tradingrepo.TradingTradeRepository,  // NEW
	utr tradeRepo.UserTradeRepository,
	sr settlRepo.SettlementRepository,
	wr walletRepo.WalletRepository,
	bus events.Bus,
	logger *slog.Logger,
) *SettlementJob {
	return &SettlementJob{
		invRepo:          ir,
		tradeRepo:        tr,
		tradingTradeRepo: ttr,  // NEW
		userTradeRepo:    utr,
		settlRepo:        sr,
		walletRepo:       wr,
		eventBus:         bus,
		logger:           logger,
		feeRate:          decimal.NewFromFloat(0.5),
	}
}

func NewSettlementJobFromDB(db *gorm.DB, bus events.Bus, logger *slog.Logger) *SettlementJob {
	return NewSettlementJob(
		investRepo.NewInvestmentRepository(db),
		tradeRepo.NewTradeRepository(db),
		tradingrepo.NewTradingTradeRepository(db),
		tradeRepo.NewUserTradeRepository(db),
		settlRepo.NewSettlementRepository(db),
		walletRepo.NewWalletRepository(db),
		bus,
		logger,
	)
}
```

Add import: `tradingrepo "github.com/sovereign-fund/sovereign/internal/modules/trading_tradelog/repository"`.

- [ ] **Step 3: Refactor RunForDate to per-product loop**

Restructure: extract the existing body into a new method `settleProduct(ctx, date, productType)`. The new `RunForDate` becomes:

```go
func (j *SettlementJob) RunForDate(ctx context.Context, date time.Time) error {
	products := []string{
		invModel.ProductTypeArbitrage,
		invModel.ProductTypeTrading,
	}
	var firstErr error
	for _, productType := range products {
		if err := j.settleProduct(ctx, date, productType); err != nil {
			j.logger.Error("settle product failed",
				slog.String("product", productType),
				slog.String("error", err.Error()),
			)
			if firstErr == nil {
				firstErr = err
			}
			// One product's failure doesn't block the other.
		}
	}
	return firstErr
}

func (j *SettlementJob) settleProduct(ctx context.Context, date time.Time, productType string) error {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)
	period := date.Format("2006-01-02")

	j.logger.Info("settlement started", slog.String("period", period), slog.String("product", productType))

	activeInvs, err := j.invRepo.FindAllActiveBeforeDateByProduct(ctx, dayStart, productType)
	if err != nil {
		return fmt.Errorf("find active investments: %w", err)
	}
	if len(activeInvs) == 0 {
		j.logger.Info("no active investments", slog.String("product", productType))
		return nil
	}

	var summary *tradeRepo.TradeSummaryResult
	if productType == invModel.ProductTypeArbitrage {
		summary, err = j.tradeRepo.SummarizeByPeriod(ctx, dayStart, dayEnd)
	} else {
		summary, err = j.tradingTradeRepo.SummarizeByPeriod(ctx, dayStart, dayEnd)
	}
	if err != nil {
		return fmt.Errorf("summarize trades: %w", err)
	}

	totalPnL := decimal.NewFromFloat(summary.TotalPnL)
	if totalPnL.LessThanOrEqual(decimal.Zero) {
		j.logger.Info("no profit to distribute", slog.String("period", period), slog.String("product", productType))
		return nil
	}

	// PRESERVE the existing distribution algorithm verbatim from the previous body.
	// Inside the per-investment loop:
	// - Set settlement.ProductType = productType
	// - When BatchCreate'ing user_trades, set ProductType = productType for each
	// - When walletRepo.AddEarnings, pass productType as the third arg
	// - When fetching dayTrades to build user_trades, branch by productType:
	//     if arbitrage: dayTrades, _ := j.tradeRepo.FindByPeriod(ctx, dayStart, dayEnd)
	//     else:         tradingDayTrades, _ := j.tradingTradeRepo.FindByPeriod(ctx, dayStart, dayEnd)
	//                   convert to []tradelog/model.UserTrade compatible inputs (same fields, different source struct)
	//
	// Trade-shape adapter: define a tiny local helper to translate *both* trade types
	// into a common struct used by the per-investment loop, OR duplicate the inner
	// loop once per type. Pick whichever keeps file readable; duplication is fine
	// since the algorithm is identical and short. Keep DRY-via-helper IF the helper
	// fits in 30 lines; otherwise duplicate.

	... (rest of original algorithm with the changes above)

	return nil
}
```

The full algorithm body is the original `RunForDate`. Walk through and apply the four changes above.

- [ ] **Step 4: Failing test for two-product isolation**

Create or extend `server/internal/worker/settlement_job_test.go`:

```go
package worker

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	invmodel "github.com/sovereign-fund/sovereign/internal/modules/investment/model"
	tradinglogmodel "github.com/sovereign-fund/sovereign/internal/modules/trading_tradelog/model"
	traderepo "github.com/sovereign-fund/sovereign/internal/modules/tradelog/repository"
	walletmodel "github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	"gorm.io/gorm"
)

// All stubs minimal — satisfy interfaces only with what the test exercises.

// ... define stubInvRepo, stubTradeRepo (returning summary by period),
//     stubTradingTradeRepo, stubUserTradeRepo, stubSettlRepo, stubWalletRepo,
//     spyBus.

func TestSettleProduct_IsolatesArbitrageFromTrading(t *testing.T) {
	ctx := context.Background()
	day := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)

	invRepo := &stubInvRepo{
		byProduct: map[string][]invmodel.Investment{
			"arbitrage": {{ID: "a1", UserID: "ua", Amount: decimal.NewFromInt(100), Currency: "USDT", ProductType: "arbitrage"}},
			"trading":   {{ID: "t1", UserID: "ut", Amount: decimal.NewFromInt(100), Currency: "USDT", ProductType: "trading"}},
		},
	}
	arbTradeRepo := &stubTradeRepo{summary: &traderepo.TradeSummaryResult{TotalTrades: 1, TotalPnL: 100}}
	tradingTradeRepo := &stubTradingTradeRepo{summary: &traderepo.TradeSummaryResult{TotalTrades: 1, TotalPnL: 200}}
	userTradeRepo := &stubUserTradeRepo{}
	settlRepo := &stubSettlRepo{}
	walletRepo := &stubWalletRepoSj{
		walletByUser: map[string]*walletmodel.Wallet{
			"ua": {ID: "wa", UserID: "ua", Currency: "USDT"},
			"ut": {ID: "wt", UserID: "ut", Currency: "USDT"},
		},
	}
	bus := &spyBus{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	job := NewSettlementJob(invRepo, arbTradeRepo, tradingTradeRepo, userTradeRepo, settlRepo, walletRepo, bus, logger)
	if err := job.RunForDate(ctx, day); err != nil {
		t.Fatalf("RunForDate error = %v", err)
	}

	if walletRepo.lastEarningsAdd["wa"].productType != "arbitrage" {
		t.Errorf("ua wallet earnings should be added with productType=arbitrage")
	}
	if walletRepo.lastEarningsAdd["wt"].productType != "trading" {
		t.Errorf("ut wallet earnings should be added with productType=trading")
	}
	// arb user must not receive trading payout
	if walletRepo.lastEarningsAdd["wa"].productType == "trading" {
		t.Error("isolation broken: arb user got trading earnings")
	}
}

func TestSettleProduct_SkipWhenNoActiveInvestments(t *testing.T) {
	ctx := context.Background()
	invRepo := &stubInvRepo{byProduct: map[string][]invmodel.Investment{}}
	job := NewSettlementJob(invRepo, &stubTradeRepo{}, &stubTradingTradeRepo{}, &stubUserTradeRepo{}, &stubSettlRepo{}, &stubWalletRepoSj{}, &spyBus{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := job.RunForDate(ctx, time.Now()); err != nil {
		t.Fatalf("error = %v", err)
	}
	// No-op success
}

func TestSettleProduct_OneProductFailureDoesNotBlockOther(t *testing.T) {
	ctx := context.Background()
	invRepo := &stubInvRepo{
		byProduct: map[string][]invmodel.Investment{
			"arbitrage": {{ID: "a1", UserID: "ua"}},
			"trading":   {{ID: "t1", UserID: "ut"}},
		},
	}
	arbTradeRepo := &stubTradeRepo{err: gorm.ErrInvalidDB} // simulate arb settle failure
	tradingTradeRepo := &stubTradingTradeRepo{summary: &traderepo.TradeSummaryResult{TotalTrades: 1, TotalPnL: 100}}
	walletRepo := &stubWalletRepoSj{walletByUser: map[string]*walletmodel.Wallet{"ut": {ID: "wt", UserID: "ut", Currency: "USDT"}}}
	job := NewSettlementJob(invRepo, arbTradeRepo, tradingTradeRepo, &stubUserTradeRepo{}, &stubSettlRepo{}, walletRepo, &spyBus{}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	_ = job.RunForDate(ctx, time.Now())  // returns first error but should still settle trading
	if walletRepo.lastEarningsAdd["wt"].productType != "trading" {
		t.Error("trading should still settle even when arb fails")
	}
}
```

Stubs are mechanical; copy-paste from the test scaffolding pattern in the wallet/admin service tests. Each stub satisfies only the methods used here.

- [ ] **Step 5: Run, expect FAIL → implement → PASS**

```bash
cd /Users/johnny/Work/soveregin/server && go test ./internal/worker/ -v
```

After implementing settleProduct, all three tests pass.

- [ ] **Step 6: Patch worker bootstrap call site**

```bash
grep -rn "NewSettlementJob\b\|NewSettlementJobFromDB" server/cmd/ server/internal/app/
```

If callers use `NewSettlementJobFromDB`, no change needed (it auto-builds from db). If anywhere uses `NewSettlementJob` directly, add the new arg.

- [ ] **Step 7: Commit**

```bash
git add server/internal/worker/
git commit -m "feat(worker): SettlementJob loops per product_type with isolated failure"
```

---

## Phase 5 — Admin user marking (TDD)

### Task 11: UserProductChangeLog model + repo

**Files:**
- Create: `server/internal/modules/admin/model/user_product_change_log.go`
- Create: `server/internal/modules/admin/repository/user_product_change_log_repo.go`

- [ ] **Step 1: Model**

```go
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserProductChangeLog struct {
	ID         string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     string    `gorm:"type:uuid;index;not null" json:"user_id"`
	FromType   string    `gorm:"type:varchar(20);not null" json:"from_type"`
	ToType     string    `gorm:"type:varchar(20);not null" json:"to_type"`
	AdminID    string    `gorm:"type:uuid;not null" json:"admin_id"`
	AdminEmail string    `gorm:"type:varchar(255);not null" json:"admin_email"`
	Reason     string    `gorm:"type:varchar(500);not null" json:"reason"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (l *UserProductChangeLog) BeforeCreate(_ *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}

func (UserProductChangeLog) TableName() string {
	return "user_product_change_logs"
}
```

- [ ] **Step 2: Repo**

```go
package repository

import (
	"context"

	"github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	"gorm.io/gorm"
)

type UserProductChangeLogRepository interface {
	Create(ctx context.Context, l *model.UserProductChangeLog) error
	ListByUser(ctx context.Context, userID string) ([]model.UserProductChangeLog, error)
}

type userProductChangeLogRepository struct {
	db *gorm.DB
}

func NewUserProductChangeLogRepository(db *gorm.DB) UserProductChangeLogRepository {
	return &userProductChangeLogRepository{db: db}
}

func (r *userProductChangeLogRepository) Create(ctx context.Context, l *model.UserProductChangeLog) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *userProductChangeLogRepository) ListByUser(ctx context.Context, userID string) ([]model.UserProductChangeLog, error) {
	var rows []model.UserProductChangeLog
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error
	return rows, err
}
```

- [ ] **Step 3: Build + commit**

```bash
cd /Users/johnny/Work/soveregin/server && go build ./...
git add server/internal/modules/admin/model/user_product_change_log.go server/internal/modules/admin/repository/user_product_change_log_repo.go
git commit -m "feat(admin): UserProductChangeLog model + repo"
```

---

### Task 12: UserProductService (TDD)

**Files:**
- Create: `server/internal/modules/admin/service/user_product_service.go`
- Create: `server/internal/modules/admin/service/user_product_service_test.go`
- Modify: `server/internal/shared/errors/codes.go`

- [ ] **Step 1: Add error codes**

In `codes.go`:
```go
	ErrInvalidProductType = New(http.StatusBadRequest, "INVALID_PRODUCT_TYPE", "invalid product type")
	ErrSameProductType    = New(http.StatusUnprocessableEntity, "SAME_PRODUCT_TYPE", "user is already on this product type")
```

- [ ] **Step 2: Failing test (stdlib testing, no testify)**

```go
package service

import (
	"context"
	"errors"
	"testing"

	usermodel "github.com/sovereign-fund/sovereign/internal/modules/auth/model"
	logmodel "github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
)

type stubUserRepo struct {
	users    map[string]*usermodel.User
	updated  []*usermodel.User
	updateErr error
}
func (s *stubUserRepo) FindByID(ctx context.Context, id string) (*usermodel.User, error) {
	if u, ok := s.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}
func (s *stubUserRepo) Update(ctx context.Context, u *usermodel.User) error {
	if s.updateErr != nil { return s.updateErr }
	s.updated = append(s.updated, u)
	if existing, ok := s.users[u.ID]; ok {
		existing.InvestmentType = u.InvestmentType
	}
	return nil
}
func (s *stubUserRepo) FindByEmail(context.Context, string) (*usermodel.User, error) { panic("unused") }
func (s *stubUserRepo) FindByGoogleID(context.Context, string) (*usermodel.User, error) { panic("unused") }
func (s *stubUserRepo) Create(context.Context, *usermodel.User) error { panic("unused") }
func (s *stubUserRepo) ExistsByEmail(context.Context, string) (bool, error) { panic("unused") }

type stubLogRepo struct {
	created []logmodel.UserProductChangeLog
}
func (s *stubLogRepo) Create(ctx context.Context, l *logmodel.UserProductChangeLog) error {
	s.created = append(s.created, *l)
	return nil
}
func (s *stubLogRepo) ListByUser(context.Context, string) ([]logmodel.UserProductChangeLog, error) {
	return nil, nil
}

func TestChange_FromArbitrageToTrading(t *testing.T) {
	ctx := context.Background()
	uRepo := &stubUserRepo{users: map[string]*usermodel.User{
		"u1": {ID: "u1", Email: "u@x.com", InvestmentType: "arbitrage"},
	}}
	lRepo := &stubLogRepo{}
	svc := NewUserProductService(uRepo, lRepo)

	if err := svc.Change(ctx, "u1", "trading", "internal pilot", "admin-1", "admin@x.com"); err != nil {
		t.Fatalf("Change error = %v", err)
	}
	if uRepo.users["u1"].InvestmentType != "trading" {
		t.Errorf("user.InvestmentType = %s, want trading", uRepo.users["u1"].InvestmentType)
	}
	if len(lRepo.created) != 1 {
		t.Fatalf("logs = %d, want 1", len(lRepo.created))
	}
	l := lRepo.created[0]
	if l.FromType != "arbitrage" || l.ToType != "trading" || l.Reason != "internal pilot" || l.AdminID != "admin-1" {
		t.Errorf("log = %+v, mismatch", l)
	}
}

func TestChange_RejectsInvalidType(t *testing.T) {
	uRepo := &stubUserRepo{users: map[string]*usermodel.User{"u1": {ID: "u1", InvestmentType: "arbitrage"}}}
	svc := NewUserProductService(uRepo, &stubLogRepo{})
	err := svc.Change(context.Background(), "u1", "garbage", "x", "a", "a@x.com")
	var ae *apperr.AppError
	if !errors.As(err, &ae) || ae.Code != "INVALID_PRODUCT_TYPE" {
		t.Errorf("error code = %v", ae)
	}
}

func TestChange_RejectsSameType(t *testing.T) {
	uRepo := &stubUserRepo{users: map[string]*usermodel.User{"u1": {ID: "u1", InvestmentType: "arbitrage"}}}
	svc := NewUserProductService(uRepo, &stubLogRepo{})
	err := svc.Change(context.Background(), "u1", "arbitrage", "x", "a", "a@x.com")
	var ae *apperr.AppError
	if !errors.As(err, &ae) || ae.Code != "SAME_PRODUCT_TYPE" {
		t.Errorf("error code = %v", ae)
	}
}

func TestBulkChange_PartialFailures(t *testing.T) {
	uRepo := &stubUserRepo{users: map[string]*usermodel.User{
		"u1": {ID: "u1", InvestmentType: "arbitrage"},
		"u2": {ID: "u2", InvestmentType: "trading"}, // already trading - same-type fail
	}}
	svc := NewUserProductService(uRepo, &stubLogRepo{})
	res, err := svc.BulkChange(context.Background(), []string{"u1", "u2", "u3"}, "trading", "x", "a", "a@x.com")
	if err != nil {
		t.Fatalf("BulkChange should not error: %v", err)
	}
	if len(res.Succeeded) != 1 || res.Succeeded[0] != "u1" {
		t.Errorf("succeeded = %v, want [u1]", res.Succeeded)
	}
	if len(res.Failed) != 2 {
		t.Errorf("failed = %v, want 2 entries", res.Failed)
	}
}
```

- [ ] **Step 3: Implement service**

```go
package service

import (
	"context"
	"fmt"

	usermodel "github.com/sovereign-fund/sovereign/internal/modules/auth/model"
	usermodelrepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	logmodel "github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	logrepo "github.com/sovereign-fund/sovereign/internal/modules/admin/repository"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
)

type UserProductService interface {
	Change(ctx context.Context, userID, toType, reason, adminID, adminEmail string) error
	BulkChange(ctx context.Context, userIDs []string, toType, reason, adminID, adminEmail string) (*BulkChangeResult, error)
	History(ctx context.Context, userID string) ([]logmodel.UserProductChangeLog, error)
}

type BulkChangeResult struct {
	Succeeded []string                `json:"succeeded"`
	Failed    []BulkChangeFailedEntry `json:"failed"`
}

type BulkChangeFailedEntry struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}

type userProductService struct {
	uRepo usermodelrepo.UserRepository
	lRepo logrepo.UserProductChangeLogRepository
}

func NewUserProductService(uRepo usermodelrepo.UserRepository, lRepo logrepo.UserProductChangeLogRepository) UserProductService {
	return &userProductService{uRepo: uRepo, lRepo: lRepo}
}

func (s *userProductService) Change(ctx context.Context, userID, toType, reason, adminID, adminEmail string) error {
	if toType != usermodel.InvestmentTypeArbitrage && toType != usermodel.InvestmentTypeTrading {
		return apperr.ErrInvalidProductType
	}
	user, err := s.uRepo.FindByID(ctx, userID)
	if err != nil {
		return apperr.ErrAccountNotFound
	}
	from := user.InvestmentType
	if from == "" {
		from = usermodel.InvestmentTypeArbitrage
	}
	if from == toType {
		return apperr.ErrSameProductType
	}
	user.InvestmentType = toType
	if err := s.uRepo.Update(ctx, user); err != nil {
		return apperr.Wrap(apperr.ErrInternal, err)
	}
	if err := s.lRepo.Create(ctx, &logmodel.UserProductChangeLog{
		UserID:     userID,
		FromType:   from,
		ToType:     toType,
		AdminID:    adminID,
		AdminEmail: adminEmail,
		Reason:     reason,
	}); err != nil {
		return apperr.Wrap(apperr.ErrInternal, fmt.Errorf("write change log: %w", err))
	}
	return nil
}

func (s *userProductService) BulkChange(ctx context.Context, userIDs []string, toType, reason, adminID, adminEmail string) (*BulkChangeResult, error) {
	res := &BulkChangeResult{}
	for _, uid := range userIDs {
		if err := s.Change(ctx, uid, toType, reason, adminID, adminEmail); err != nil {
			res.Failed = append(res.Failed, BulkChangeFailedEntry{UserID: uid, Reason: err.Error()})
			continue
		}
		res.Succeeded = append(res.Succeeded, uid)
	}
	return res, nil
}

func (s *userProductService) History(ctx context.Context, userID string) ([]logmodel.UserProductChangeLog, error) {
	return s.lRepo.ListByUser(ctx, userID)
}
```

- [ ] **Step 4: Run + commit**

```bash
cd /Users/johnny/Work/soveregin/server && go test ./internal/modules/admin/service/ -run TestChange -v && go test ./internal/modules/admin/service/ -run TestBulkChange -v
git add server/internal/modules/admin/service/user_product_service.go server/internal/modules/admin/service/user_product_service_test.go server/internal/shared/errors/codes.go
git commit -m "feat(admin): UserProductService for tagging users + bulk + history"
```

---

### Task 13: UserProduct handler + routes

**Files:**
- Create: `server/internal/modules/admin/handler/user_product_handler.go`
- Modify: `server/internal/modules/admin/dto/{request,response}.go`
- Modify: `server/internal/modules/admin/routes.go`
- Modify: `server/internal/modules/admin/module.go`

- [ ] **Step 1: DTOs**

In `dto/request.go`:
```go
type ChangeUserProductRequest struct {
	Type   string `json:"type" binding:"required,oneof=arbitrage trading"`
	Reason string `json:"reason" binding:"required,min=5,max=500"`
}

type BulkChangeUserProductRequest struct {
	UserIDs []string `json:"user_ids" binding:"required,min=1,max=200,dive,uuid"`
	Type    string   `json:"type" binding:"required,oneof=arbitrage trading"`
	Reason  string   `json:"reason" binding:"required,min=5,max=500"`
}
```

In `dto/response.go` (and any user-detail response struct), add `InvestmentType string `json:"investment_type"`` field.

- [ ] **Step 2: Handler**

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

type UserProductHandler struct {
	svc      service.UserProductService
	auditSvc service.AuditService
}

func NewUserProductHandler(svc service.UserProductService, auditSvc service.AuditService) *UserProductHandler {
	return &UserProductHandler{svc: svc, auditSvc: auditSvc}
}

func (h *UserProductHandler) Change(c *gin.Context) {
	id := c.Param("id")
	var req dto.ChangeUserProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	adminID := c.GetString("admin_id")
	adminEmail := c.GetString("admin_email")
	err := h.svc.Change(c.Request.Context(), id, req.Type, req.Reason, adminID, adminEmail)
	h.audit(c, "change_investment_type", id, "to="+req.Type+" reason="+req.Reason)
	h.respond(c, err)
}

func (h *UserProductHandler) BulkChange(c *gin.Context) {
	var req dto.BulkChangeUserProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	adminID := c.GetString("admin_id")
	adminEmail := c.GetString("admin_email")
	res, err := h.svc.BulkChange(c.Request.Context(), req.UserIDs, req.Type, req.Reason, adminID, adminEmail)
	h.audit(c, "bulk_change_investment_type", "", fmt.Sprintf("to=%s count=%d reason=%s", req.Type, len(req.UserIDs), req.Reason))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	response.OK(c, res)
}

func (h *UserProductHandler) History(c *gin.Context) {
	id := c.Param("id")
	rows, err := h.svc.History(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "HISTORY_FAILED", err.Error())
		return
	}
	response.OK(c, rows)
}

func (h *UserProductHandler) audit(c *gin.Context, action, id, detail string) {
	if err := h.auditSvc.Log(c.Request.Context(), c.GetString("admin_id"), c.GetString("admin_email"), action, "user", id, detail, c.ClientIP()); err != nil {
		slog.Error("audit log failed", slog.String("action", action), slog.String("error", err.Error()))
	}
}

func (h *UserProductHandler) respond(c *gin.Context, err error) {
	if err == nil {
		response.OK(c, gin.H{"message": "ok"})
		return
	}
	var ae *apperr.AppError
	if errors.As(err, &ae) {
		response.Fail(c, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
}
```

Add `import "fmt"`.

- [ ] **Step 3: Wire module + routes**

In `module.go` add `UserProductHandler *handler.UserProductHandler` field; construct `userProductSvc := service.NewUserProductService(userRepo, repository.NewUserProductChangeLogRepository(db))` (you'll need to thread `userRepo` into the admin module — pass it from app.go like the wallet/withdraw plumbing).

In `routes.go` add inside the `users` group:
```go
		users.POST("/:id/investment-type",
			middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator),
			m.UserProductHandler.Change,
		)
		users.GET("/:id/product-changes",
			middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator),
			m.UserProductHandler.History,
		)
```

And as a sibling outside the group:
```go
	protected.POST("/users/bulk-investment-type",
		middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator),
		m.UserProductHandler.BulkChange,
	)
```

- [ ] **Step 4: Patch UserService.Detail to include InvestmentType in user detail response**

```bash
grep -n "InvestmentType" server/internal/modules/admin/service/user_service.go
```

If absent, add `InvestmentType: user.InvestmentType` when the detail is built.

- [ ] **Step 5: Build + test + commit**

```bash
cd /Users/johnny/Work/soveregin/server && go build ./... && go test ./internal/modules/admin/...
git add server/internal/modules/admin/ server/internal/app/
git commit -m "feat(admin): user product type change endpoints + audit"
```

---

## Phase 6 — Admin trading_trades module

### Task 14: Trading trade service + handler

**Files:**
- Create: `server/internal/modules/admin/service/trading_trade_service.go`
- Create: `server/internal/modules/admin/handler/trading_trade_handler.go`
- Modify: `server/internal/modules/admin/dto/{request,response}.go`
- Modify: `server/internal/modules/admin/routes.go`
- Modify: `server/internal/modules/admin/module.go`

The pattern is a near-copy of the existing `trade_service.go` + `trade_handler.go`. To avoid duplication:

- [ ] **Step 1: Export reusable parsing**

In `server/internal/modules/admin/service/trade_service.go`, rename `parseTradeRow` to `ParseTradeRow` (export). Same for `parseImportRows` → `ParseImportRows`. Keep their signatures (return `[]trademodel.Trade` because the column shape is identical).

After exporting, update internal callers in trade_service.go (`parseImportRows(file)` → `ParseImportRows(file)`).

Run `go build ./...` to confirm nothing breaks.

- [ ] **Step 2: Trading trade service**

Create `trading_trade_service.go`:
```go
package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	tradingmodel "github.com/sovereign-fund/sovereign/internal/modules/trading_tradelog/model"
	tradingrepo "github.com/sovereign-fund/sovereign/internal/modules/trading_tradelog/repository"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type TradingTradeService interface {
	List(ctx context.Context, query dto.TradeListQuery) ([]dto.TradeListItem, int64, error)
	Stats(ctx context.Context) (*dto.TradeStats, error)
	DownloadTemplate(ctx context.Context) (*excelize.File, error)
	ImportFromExcel(ctx context.Context, file multipart.File) (int, []string, error)
	Delete(ctx context.Context, tradeID string) error
}

type tradingTradeService struct {
	db   *gorm.DB
	repo tradingrepo.TradingTradeRepository
}

func NewTradingTradeService(db *gorm.DB) TradingTradeService {
	return &tradingTradeService{db: db, repo: tradingrepo.NewTradingTradeRepository(db)}
}

func (s *tradingTradeService) List(ctx context.Context, query dto.TradeListQuery) ([]dto.TradeListItem, int64, error) {
	db := s.db.WithContext(ctx).Model(&tradingmodel.TradingTrade{})
	if query.Pair != "" {
		db = db.Where("pair ILIKE ?", "%"+query.Pair+"%")
	}
	if query.DateFrom != "" {
		db = db.Where("executed_at >= ?", query.DateFrom)
	}
	if query.DateTo != "" {
		db = db.Where("executed_at < ?", query.DateTo+" 23:59:59")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count trading_trades: %w", err)
	}
	var trades []tradingmodel.TradingTrade
	offset := (query.Page - 1) * query.Limit
	if err := db.Order("executed_at DESC").Offset(offset).Limit(query.Limit).Find(&trades).Error; err != nil {
		return nil, 0, fmt.Errorf("find trading_trades: %w", err)
	}
	items := make([]dto.TradeListItem, len(trades))
	for i, t := range trades {
		items[i] = dto.TradeListItem{
			ID:           t.ID,
			Pair:         t.Pair,
			BuyExchange:  t.BuyExchange,
			SellExchange: t.SellExchange,
			BuyPrice:     t.BuyPrice.StringFixed(4),
			SellPrice:    t.SellPrice.StringFixed(4),
			Amount:       t.Amount.StringFixed(2),
			PremiumPct:   t.PremiumPct.StringFixed(2),
			PnL:          t.PnL.StringFixed(2),
			Fee:          t.Fee.StringFixed(2),
			Source:       t.Source,
			ExecutedAt:   t.ExecutedAt.Format(time.RFC3339),
		}
	}
	return items, total, nil
}

func (s *tradingTradeService) DownloadTemplate(ctx context.Context) (*excelize.File, error) {
	// Same template; pair-identical fields. Reuse arb's template builder via a small helper:
	// easiest path — call (s2 *tradeService).DownloadTemplate via a dedicated helper.
	// Or duplicate the 6 lines. Choose duplication here to avoid coupling.
	file := excelize.NewFile()
	sheet := file.GetSheetList()[0]
	file.SetSheetName(sheet, "TradingTrades")
	for col, value := range tradeTemplateHeaders {
		_ = file.SetCellValue("TradingTrades", fmt.Sprintf("%s1", tradeTemplateColumns[col]), value)
	}
	for col, value := range tradeTemplateSampleRow {
		_ = file.SetCellValue("TradingTrades", fmt.Sprintf("%s2", tradeTemplateColumns[col]), value)
	}
	for _, w := range tradeTemplateWidths {
		_ = file.SetColWidth("TradingTrades", w.Start, w.End, w.Width)
	}
	return file, nil
}

func (s *tradingTradeService) ImportFromExcel(ctx context.Context, file multipart.File) (int, []string, error) {
	parsed, rowErrors, err := ParseImportRows(file) // exported in Task 14 step 1
	if err != nil {
		return 0, nil, err
	}
	if len(parsed) == 0 {
		return 0, rowErrors, nil
	}
	trading := make([]tradingmodel.TradingTrade, len(parsed))
	for i, p := range parsed {
		trading[i] = tradingmodel.TradingTrade{
			Pair:         p.Pair,
			BuyExchange:  p.BuyExchange,
			SellExchange: p.SellExchange,
			BuyPrice:     p.BuyPrice,
			SellPrice:    p.SellPrice,
			Amount:       p.Amount,
			PremiumPct:   p.PremiumPct,
			PnL:          p.PnL,
			Fee:          p.Fee,
			Source:       p.Source,
			ExecutedAt:   p.ExecutedAt,
		}
	}
	if err := s.repo.BatchCreate(ctx, trading); err != nil {
		return 0, rowErrors, fmt.Errorf("import trading_trades: %w", err)
	}
	return len(trading), rowErrors, nil
}

func (s *tradingTradeService) Stats(ctx context.Context) (*dto.TradeStats, error) {
	// Same shape as tradeService.Stats but query trading_trades. Copy the pattern.
	// (Implement by translating each SQL select to use tradingmodel.TradingTrade.)
	// Detailed code identical to tradeService.Stats; substitute the model type.
	return nil, fmt.Errorf("stats not implemented")  // TODO replaced by full impl below
}

func (s *tradingTradeService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
```

NOTE: replace the `Stats` stub with the full body from `tradeService.Stats` swapping `&trademodel.Trade{}` → `&tradingmodel.TradingTrade{}`. Do not leave it stubbed.

- [ ] **Step 3: Trading trade handler**

```go
package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type TradingTradeHandler struct {
	svc      service.TradingTradeService
	auditSvc service.AuditService
}

func NewTradingTradeHandler(svc service.TradingTradeService, auditSvc service.AuditService) *TradingTradeHandler {
	return &TradingTradeHandler{svc: svc, auditSvc: auditSvc}
}

func (h *TradingTradeHandler) List(c *gin.Context) {
	var query dto.TradeListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_QUERY", err.Error())
		return
	}
	if query.Page < 1 { query.Page = 1 }
	if query.Limit < 1 || query.Limit > 100 { query.Limit = 20 }
	items, total, err := h.svc.List(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "TRADING_LIST_FAILED", err.Error())
		return
	}
	response.Paginated(c, items, response.Meta{Total: total, Page: query.Page, PerPage: query.Limit})
}

func (h *TradingTradeHandler) Stats(c *gin.Context) {
	stats, err := h.svc.Stats(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "TRADING_STATS_FAILED", err.Error())
		return
	}
	response.OK(c, stats)
}

func (h *TradingTradeHandler) DownloadTemplate(c *gin.Context) {
	file, err := h.svc.DownloadTemplate(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "TRADING_TEMPLATE_FAILED", err.Error())
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="trading-trade-import-template.xlsx"`)
	if err := file.Write(c.Writer); err != nil {
		response.Fail(c, http.StatusInternalServerError, "TRADING_TEMPLATE_WRITE_FAILED", err.Error())
	}
}

func (h *TradingTradeHandler) Import(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "FILE_REQUIRED", err.Error())
		return
	}
	defer file.Close()
	count, rowErrors, err := h.svc.ImportFromExcel(c.Request.Context(), file)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "TRADING_IMPORT_FAILED", err.Error())
		return
	}
	if err := h.auditSvc.Log(c.Request.Context(), c.GetString("admin_id"), c.GetString("admin_email"), "import_trading_trades", "trading_trade", "", fmt.Sprintf("count=%d errors=%d", count, len(rowErrors)), c.ClientIP()); err != nil {
		log.Printf("audit log failed: %v", err)
	}
	response.OK(c, gin.H{"count": count, "row_errors": rowErrors})
}

func (h *TradingTradeHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, http.StatusBadRequest, "DELETE_TRADING_FAILED", err.Error())
		return
	}
	if err := h.auditSvc.Log(c.Request.Context(), c.GetString("admin_id"), c.GetString("admin_email"), "delete_trading_trade", "trading_trade", id, "", c.ClientIP()); err != nil {
		log.Printf("audit log failed: %v", err)
	}
	response.OK(c, gin.H{"message": "deleted"})
}
```

Add imports `"fmt"` etc.

- [ ] **Step 4: Wire module + routes**

In `module.go` add:
- field `TradingTradeHandler *handler.TradingTradeHandler`
- construct `tradingTradeSvc := service.NewTradingTradeService(db)` and assign

In `routes.go` add:
```go
	tradingTrades := protected.Group("/trading-trades")
	{
		tradingTrades.GET("",
			middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator, model.RoleViewer),
			m.TradingTradeHandler.List,
		)
		tradingTrades.GET("/stats",
			middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator, model.RoleViewer),
			m.TradingTradeHandler.Stats,
		)
		tradingTrades.GET("/template",
			middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator),
			m.TradingTradeHandler.DownloadTemplate,
		)
		tradingTrades.POST("/import",
			middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator),
			m.TradingTradeHandler.Import,
		)
		tradingTrades.DELETE("/:id",
			middleware.RequireRole(model.RoleSuperAdmin),
			m.TradingTradeHandler.Delete,
		)
	}
```

- [ ] **Step 5: Build + commit**

```bash
cd /Users/johnny/Work/soveregin/server && go build ./... && go test ./internal/modules/admin/...
git add server/internal/modules/admin/ server/internal/app/
git commit -m "feat(admin): trading_trades CRUD/import endpoints + audit"
```

---

## Phase 7 — Admin frontend

### Task 15: User detail page — InvestmentType card

**Files:**
- Create: `admin/src/pages/UserDetail/InvestmentTypeCard.tsx`
- Modify: `admin/src/pages/UserDetail/index.tsx`
- Modify: `admin/src/services/api.ts`

- [ ] **Step 1: API helpers**

Append to `admin/src/services/api.ts`:
```typescript
export interface UserProductChangeLog {
  id: string;
  user_id: string;
  from_type: string;
  to_type: string;
  admin_id: string;
  admin_email: string;
  reason: string;
  created_at: string;
}

export async function changeUserInvestmentType(id: string, type: 'arbitrage' | 'trading', reason: string) {
  return request<API.ApiResponse<{ message: string }>>(`/users/${id}/investment-type`, {
    method: 'POST',
    data: { type, reason },
  });
}

export async function getUserProductChanges(id: string) {
  return request<API.ApiResponse<UserProductChangeLog[]>>(`/users/${id}/product-changes`, {
    method: 'GET',
  });
}

export async function bulkChangeUserInvestmentType(user_ids: string[], type: 'arbitrage' | 'trading', reason: string) {
  return request<API.ApiResponse<{ succeeded: string[]; failed: { user_id: string; reason: string }[] }>>(`/users/bulk-investment-type`, {
    method: 'POST',
    data: { user_ids, type, reason },
  });
}
```

- [ ] **Step 2: InvestmentTypeCard component**

```tsx
import { Card, Tag, Button, Modal, Form, Input, message, Timeline } from 'antd';
import { useState } from 'react';
import { useRequest } from 'ahooks';
import { changeUserInvestmentType, getUserProductChanges } from '@/services/api';

interface Props {
  userId: string;
  currentType: string;
  onChanged: () => void;
}

const LABEL: Record<string, { text: string; color: string }> = {
  arbitrage: { text: '套利投资', color: 'blue' },
  trading: { text: '交易投资', color: 'gold' },
};

export default function InvestmentTypeCard({ userId, currentType, onChanged }: Props) {
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();
  const { data: history, refresh: refreshHistory } = useRequest(() => getUserProductChanges(userId), { refreshDeps: [userId] });

  const target = currentType === 'trading' ? 'arbitrage' : 'trading';
  const meta = LABEL[currentType] ?? { text: currentType, color: 'default' };

  const handleOk = async () => {
    const values = await form.validateFields();
    try {
      await changeUserInvestmentType(userId, target, values.reason);
      message.success('已更新');
      form.resetFields();
      setModalOpen(false);
      onChanged();
      refreshHistory();
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message ?? '操作失败');
    }
  };

  return (
    <Card title="投资产品类型" extra={<Button type="primary" onClick={() => setModalOpen(true)}>切换为 {LABEL[target].text}</Button>}>
      <div style={{ marginBottom: 16 }}>
        当前类型：<Tag color={meta.color}>{meta.text}</Tag>
      </div>
      <h4>变更历史</h4>
      {history?.data?.length ? (
        <Timeline items={history.data.map((l) => ({
          children: (
            <div>
              <div>{l.from_type} → {l.to_type}</div>
              <div style={{ color: '#888', fontSize: 12 }}>{l.admin_email} · {new Date(l.created_at).toLocaleString()}</div>
              <div style={{ marginTop: 4 }}>原因：{l.reason}</div>
            </div>
          ),
        }))} />
      ) : <span style={{ color: '#888' }}>暂无变更记录</span>}

      <Modal title={`切换为 ${LABEL[target].text}`} open={modalOpen} onOk={handleOk} onCancel={() => { form.resetFields(); setModalOpen(false); }} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="reason" label="变更原因" rules={[{ required: true, min: 5, max: 500 }]}>
            <Input.TextArea rows={3} placeholder="例：参与交易投资内测" />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}
```

- [ ] **Step 3: Embed in UserDetail page**

In `admin/src/pages/UserDetail/index.tsx`, locate the layout where existing cards live (e.g. balance card, security card). Add:
```tsx
import InvestmentTypeCard from './InvestmentTypeCard';
// ...
<InvestmentTypeCard
  userId={user.id}
  currentType={user.investment_type ?? 'arbitrage'}
  onChanged={() => refreshUser()}
/>
```

(Adapt to whatever refresh function the page uses.)

- [ ] **Step 4: Build + commit**

```bash
cd /Users/johnny/Work/soveregin/admin && pnpm tsc --noEmit
git add admin/
git commit -m "feat(admin-ui): user detail InvestmentTypeCard with history + change modal"
```

---

### Task 16: Users list — column + filter + bulk-tag

**Files:**
- Modify: `admin/src/pages/Users/index.tsx`
- Create: `admin/src/pages/Users/BulkTagModal.tsx`

- [ ] **Step 1: BulkTagModal**

```tsx
import { Modal, Form, Select, Input, message } from 'antd';
import { bulkChangeUserInvestmentType } from '@/services/api';

interface Props {
  open: boolean;
  selectedUserIds: string[];
  onClose: () => void;
  onDone: () => void;
}

export default function BulkTagModal({ open, selectedUserIds, onClose, onDone }: Props) {
  const [form] = Form.useForm();

  const handleOk = async () => {
    const values = await form.validateFields();
    try {
      const res = await bulkChangeUserInvestmentType(selectedUserIds, values.type, values.reason);
      const data = (res as any).data;
      message.success(`成功 ${data.succeeded.length}, 失败 ${data.failed.length}`);
      form.resetFields();
      onDone();
    } catch (e: any) {
      message.error(e?.response?.data?.error?.message ?? '操作失败');
    }
  };

  return (
    <Modal title={`批量打标 (${selectedUserIds.length} 个用户)`} open={open} onOk={handleOk} onCancel={() => { form.resetFields(); onClose(); }} destroyOnClose>
      <Form form={form} layout="vertical">
        <Form.Item name="type" label="目标类型" rules={[{ required: true }]}>
          <Select options={[{ value: 'arbitrage', label: '套利投资' }, { value: 'trading', label: '交易投资' }]} />
        </Form.Item>
        <Form.Item name="reason" label="变更原因" rules={[{ required: true, min: 5, max: 500 }]}>
          <Input.TextArea rows={3} />
        </Form.Item>
      </Form>
    </Modal>
  );
}
```

- [ ] **Step 2: Modify Users page**

Add a column `投资类型` rendering a `<Tag>` for `record.investment_type`. Add column-level filter (Select) and a top-right `<Button>` "批量打标" enabled when rows are selected. Use `rowSelection` from `ProTable`. On click open BulkTagModal.

```tsx
const [selectedKeys, setSelectedKeys] = useState<string[]>([]);
const [bulkOpen, setBulkOpen] = useState(false);
// ...
<ProTable
  rowSelection={{ selectedRowKeys: selectedKeys, onChange: (keys) => setSelectedKeys(keys as string[]) }}
  toolBarRender={() => [
    <Button key="bulk" disabled={selectedKeys.length === 0} onClick={() => setBulkOpen(true)}>批量打标</Button>,
  ]}
  // ...existing columns...
/>
<BulkTagModal
  open={bulkOpen}
  selectedUserIds={selectedKeys}
  onClose={() => setBulkOpen(false)}
  onDone={() => { setBulkOpen(false); setSelectedKeys([]); actionRef.current?.reload(); }}
/>
```

- [ ] **Step 3: Build + commit**

```bash
cd /Users/johnny/Work/soveregin/admin && pnpm tsc --noEmit
git add admin/
git commit -m "feat(admin-ui): users list bulk-tag investment_type"
```

---

### Task 17: Trading Trades pages

**Files:**
- Create: `admin/src/pages/TradingTrades/index.tsx`
- Create: `admin/src/pages/TradingTrades/Import.tsx`
- Modify: `admin/src/services/api.ts`
- Modify: `admin/config/routes.ts`

- [ ] **Step 1: API helpers**

Append to `admin/src/services/api.ts`:
```typescript
export async function listTradingTrades(params: { page?: number; limit?: number; pair?: string; date_from?: string; date_to?: string }) {
  return request<API.ApiResponse<API.PaginatedItem[]>>('/trading-trades', { method: 'GET', params });
}
export async function getTradingTradeStats() {
  return request<API.ApiResponse<any>>('/trading-trades/stats', { method: 'GET' });
}
export async function downloadTradingTradeTemplate() {
  return request<Blob>('/trading-trades/template', { method: 'GET', responseType: 'blob' });
}
export async function importTradingTrades(file: File) {
  const formData = new FormData();
  formData.append('file', file);
  return request<API.ApiResponse<{ count: number; row_errors: string[] }>>('/trading-trades/import', { method: 'POST', data: formData });
}
export async function deleteTradingTrade(id: string) {
  return request<API.ApiResponse<{ message: string }>>(`/trading-trades/${id}`, { method: 'DELETE' });
}
```

- [ ] **Step 2: TradingTrades index page**

Mirror `admin/src/pages/Trades/index.tsx` exactly — copy that file, swap `listTrades` → `listTradingTrades`, `deleteTrade` → `deleteTradingTrade`, page title to "交易记录（交易策略）". Open the existing file:
```bash
cat admin/src/pages/Trades/index.tsx
```
and produce a parallel `TradingTrades/index.tsx`.

- [ ] **Step 3: TradingTrades Import page**

Mirror existing `Trades/Import.tsx` (or wherever the import UI lives).

- [ ] **Step 4: Routes**

In `admin/config/routes.ts` add:
```ts
{
  path: '/trading-trades',
  name: '交易策略',
  icon: 'LineChartOutlined',
  routes: [
    { path: '/trading-trades', redirect: '/trading-trades/list' },
    { path: '/trading-trades/list', name: '交易记录', component: './TradingTrades' },
    { path: '/trading-trades/import', name: '导入交易记录', component: './TradingTrades/Import' },
  ],
},
```

- [ ] **Step 5: Build + commit**

```bash
cd /Users/johnny/Work/soveregin/admin && pnpm tsc --noEmit
git add admin/
git commit -m "feat(admin-ui): trading-trades list + import pages"
```

---

## Phase 8 — User frontend

### Task 18: Tabs on investments + reports

**Files:**
- Modify: `front/src/types/api.ts`
- Modify: `front/src/hooks/use-api.ts`
- Modify: `front/src/app/(app)/investments/page.tsx` (path may differ — locate it)
- Modify: `front/src/app/(app)/reports/page.tsx` (or settlements equivalent)
- Modify: `front/src/i18n/{en,zh,ko}.json`

- [ ] **Step 1: Type updates**

In `front/src/types/api.ts`:
```typescript
export interface User {
  // ...
  investment_type?: 'arbitrage' | 'trading';
}

export interface Investment {
  // ...
  product_type?: 'arbitrage' | 'trading';
}

export interface Settlement {
  // ...
  product_type?: 'arbitrage' | 'trading';
}
```

- [ ] **Step 2: Hook updates**

```typescript
export function useInvestments(productType?: string) {
  return useQuery({
    queryKey: ['investments', productType],
    queryFn: () => api.get<Investment[]>(`/investments`, { params: { product_type: productType } }),
  })
}

export function useSettlements(productType?: string) {
  return useQuery({
    queryKey: ['settlements', productType],
    queryFn: () => api.get<Settlement[]>(`/settlements`, { params: { product_type: productType } }),
  })
}
```

(Match the project's actual api wrapper signature.)

- [ ] **Step 3: i18n keys**

Add to wallet/investment block in each locale:
- zh: `tabCurrent: "当前产品"`, `tabHistory: "历史产品"`
- en: `tabCurrent: "Current"`, `tabHistory: "Historical"`
- ko: `tabCurrent: "현재"`, `tabHistory: "이력"`

- [ ] **Step 4: Investments page**

Read first to understand structure:
```bash
find front/src/app -name "page.tsx" | xargs grep -l investment
```

Surgical edit:
- Fetch investments twice (or once with `all`) and group client-side by `product_type`
- If `current.length > 0 && historical.length > 0`, render tab bar; otherwise render the single list
- "Current" = those matching `user.investment_type`; "Historical" = the others

```tsx
const { data: user } = useUser()
const currentType = user?.investment_type ?? 'arbitrage'
const { data: all } = useInvestments('all')
const current = (all ?? []).filter(i => i.product_type === currentType)
const historical = (all ?? []).filter(i => i.product_type !== currentType)
const [tab, setTab] = useState<'current' | 'historical'>('current')
const showTabs = current.length > 0 && historical.length > 0
const visible = !showTabs ? (all ?? []) : (tab === 'current' ? current : historical)
```

Render `<Tabs>` or simple buttons depending on existing UI components in the page.

- [ ] **Step 5: Reports / Settlements page — same pattern**

- [ ] **Step 6: Build + commit**

```bash
cd /Users/johnny/Work/soveregin/front && pnpm tsc --noEmit
git add front/
git commit -m "feat(front): tab investments and reports by product_type"
```

---

## Phase 9 — Final verification

### Task 19: Full test + build sweep + PR

- [ ] **Step 1: Backend full test**

```bash
cd /Users/johnny/Work/soveregin/server && go test ./... -race -count=1
```
Expect: all PASS.

- [ ] **Step 2: Builds**

```bash
cd /Users/johnny/Work/soveregin/server && go build ./...
cd /Users/johnny/Work/soveregin/admin && pnpm tsc --noEmit
cd /Users/johnny/Work/soveregin/front && pnpm tsc --noEmit
```

- [ ] **Step 3: Commit count + diff stat sanity**

```bash
cd /Users/johnny/Work/soveregin && git log --oneline <base-of-this-feature>..HEAD
git diff --stat <base-of-this-feature>..HEAD
```

- [ ] **Step 4: Manual staging walkthrough (REQUIRED before PR)**

1. Apply migration 024 on staging.
2. Verify `users.investment_type` exists; all rows backfilled to `arbitrage`.
3. Verify `wallets.earnings` is gone; `earnings_arbitrage` matches the previous earnings amount.
4. Pick a test user. As admin → User detail → switch to trading (reason: "internal pilot test"). Verify badge updates and history shows the entry.
5. As that test user → click invest → verify the new investment lands with `product_type=trading` (DB or API GET).
6. Import a sample `trading_trades` row dated yesterday.
7. Trigger settlement: `./deployments/deploy.sh exec server ./settle <yesterday-date>` (or whatever manual trigger exists).
8. Verify only the trading user's `wallets.earnings_trading` increased; arbitrage users untouched.
9. Verify the settlement row in `settlements` has `product_type='trading'`.
10. As that user, claim earnings; verify `available` increased by the sum.
11. Bulk-tag two more users. Verify history rows in `user_product_change_logs`.

- [ ] **Step 5: Open PR**

```bash
git push origin main   # or feature branch if you migrated
gh pr create --title "feat: trading investment product (internal beta)" --body "$(cat <<'EOF'
## Summary
- New "trading" investment product alongside existing arbitrage.
- Users tagged by admins (single or bulk); same user-facing button routes to the matching product pool.
- Trading fund-level trades live in a new `trading_trades` table; investments / user_trades / settlements share schema with `product_type` column.
- Settlement job loops both products daily with isolated failure semantics.
- Wallet earnings stored in two columns (`earnings_arbitrage` + `earnings_trading`), aggregated to a single number on the user frontend.
- Admin UI: per-user + bulk tagging on user detail/list; new `/admin/trading-trades` import + list pages.

Spec: `docs/superpowers/specs/2026-05-06-trading-product-design.md`
Plan: `docs/superpowers/plans/2026-05-06-trading-product.md`

## Test plan
- [ ] Migration 024 applies cleanly + backfills (`users.investment_type='arbitrage'`, `wallets.earnings_arbitrage=earnings`)
- [ ] Tag a user trading → user creates investment → `product_type='trading'`
- [ ] Import trading_trades → settlement → only trading users get earnings_trading
- [ ] Arbitrage users untouched after trading settlement
- [ ] Reject same-type / invalid type errors surfaced
- [ ] Bulk tag returns succeeded/failed correctly
- [ ] Frontend tab appears only when user has both products
- [ ] Wallet earnings card shows aggregated total
EOF
)"
```

---

## Self-Review Notes

- **Spec coverage:**
  - §3 decisions table: each row → tasks (1, 2, 3, 4, 5, 6, 10, 11, 12, 13, 14, 15-17, 18)
  - §4 Schema: Task 1 (migration) + Tasks 2/3/4/5 (model fields)
  - §5 Investment routing: Task 7
  - §6 Settlement double loop: Task 10
  - §7 User APIs: Tasks 8, 9 + Task 5 (wallet aggregation)
  - §8 Admin APIs: Tasks 12, 13, 14
  - §9 Service refactors: Tasks 5, 7, 8, 9
  - §10 Frontend: Tasks 15, 16, 17 (admin) + Task 18 (user)
  - §11 Audit: Tasks 12, 13 (audit_log calls)
  - §12 Tests: TDD steps embedded in 7, 10, 12

- **Placeholder scan:** No "TBD" / "implement later" remaining in code blocks. The Stats stub in Task 14 is annotated "replace with full impl from tradeService.Stats" — engineer must execute, not leave stubbed.

- **Type consistency:**
  - `ProductType` constant naming consistent: `model.ProductTypeArbitrage`/`Trading` (investment package), mirrored as `usermodel.InvestmentTypeArbitrage`/`Trading`
  - `AddEarnings` signature change (3-arg with productType) propagated through callers in Task 5
  - `NewInvestmentService` signature change (added userRepo) propagated in Task 7
  - `NewSettlementJob` signature change (added tradingTradeRepo) propagated in Task 10

- **Known follow-ups:**
  - E2E (deferred — same situation as withdraw-review plan; no Playwright in repo yet)
  - Manual staging walkthrough (REQUIRED, listed in Task 19)
  - Migration apply on real DB (REQUIRED, can't run in sandbox)
