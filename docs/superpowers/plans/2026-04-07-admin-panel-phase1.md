# Admin Panel Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a mobile-first admin panel with admin authentication (JWT), role-based access (super_admin/operator/viewer), user management (CRUD + disable/enable/reset password/adjust balance), admin management, and dashboard statistics.

**Architecture:** Backend adds `internal/modules/admin` module to the existing Go/Gin server with independent JWT auth. Frontend is a new `admin/` Next.js project using `@arco-design/mobile-react` for mobile H5. Both share the same database. Admin JWT is completely isolated from user JWT.

**Tech Stack:** Go/Gin/GORM (backend), Next.js 16 + React 19 + @arco-design/mobile-react + TanStack Query (frontend), PostgreSQL, pnpm

**Arco Design Mobile key components:** NavBar (top bar), TabBar (bottom tabs), Cell (list items), SearchBar, PullRefresh, LoadMore, Popup, Dialog, Toast, Form, Input, Button, Tabs, Tag, Grid, Picker, SwipeAction

**Arco Design Mobile imports:** `import { Component } from '@arco-design/mobile-react'` and `import '@arco-design/mobile-react/esm/style'` for global styles.

---

## File Map

### Backend (server/)

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `migrations/000019_create_admin_users.up.sql` | admin_users table |
| Create | `migrations/000019_create_admin_users.down.sql` | rollback |
| Create | `internal/modules/admin/model/admin_user.go` | AdminUser model + role constants |
| Create | `internal/modules/admin/repository/admin_repo.go` | AdminUser CRUD repository |
| Create | `internal/modules/admin/dto/request.go` | Request DTOs (login, create/update admin, user filters, adjust balance) |
| Create | `internal/modules/admin/dto/response.go` | Response DTOs (admin profile, user detail, dashboard stats) |
| Create | `internal/modules/admin/middleware/admin_auth.go` | Admin JWT auth + RequireRole middleware |
| Create | `internal/modules/admin/service/auth_service.go` | Admin login + change password |
| Create | `internal/modules/admin/service/admin_user_service.go` | Admin CRUD |
| Create | `internal/modules/admin/service/user_service.go` | User management (list, detail, edit, disable, reset pwd, adjust balance) |
| Create | `internal/modules/admin/service/dashboard_service.go` | Dashboard stats queries |
| Create | `internal/modules/admin/handler/auth_handler.go` | Login + change password handlers |
| Create | `internal/modules/admin/handler/admin_user_handler.go` | Admin CRUD handlers |
| Create | `internal/modules/admin/handler/user_handler.go` | User management handlers |
| Create | `internal/modules/admin/handler/dashboard_handler.go` | Dashboard stats handler |
| Create | `internal/modules/admin/routes.go` | Route registration with role-based middleware groups |
| Create | `internal/modules/admin/module.go` | Module init |
| Modify | `config/config.go` | Add AdminConfig struct |
| Modify | `config/config.yaml` | Add admin section defaults |
| Modify | `config/loader.go` | Add admin to subConfigs |
| Modify | `internal/app/app.go` | Add AdminModule field + init |
| Modify | `internal/app/router.go` | Register admin routes |
| Create | `scripts/seed_admin.go` | Seed first super_admin |

### Frontend (admin/)

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `admin/package.json` | Project dependencies |
| Create | `admin/next.config.ts` | Next.js config |
| Create | `admin/tsconfig.json` | TypeScript config |
| Create | `admin/src/app/layout.tsx` | Root layout + providers |
| Create | `admin/src/app/globals.css` | Global styles |
| Create | `admin/src/lib/api.ts` | Axios instance + token interceptor |
| Create | `admin/src/lib/auth.ts` | Auth state management (zustand) |
| Create | `admin/src/types/api.ts` | TypeScript types |
| Create | `admin/src/hooks/use-api.ts` | TanStack Query hooks |
| Create | `admin/src/app/login/page.tsx` | Login page |
| Create | `admin/src/components/layout/app-layout.tsx` | TabBar + NavBar layout shell |
| Create | `admin/src/app/(admin)/layout.tsx` | Auth-guarded layout |
| Create | `admin/src/app/(admin)/dashboard/page.tsx` | Dashboard stats page |
| Create | `admin/src/app/(admin)/users/page.tsx` | User list page |
| Create | `admin/src/app/(admin)/users/[id]/page.tsx` | User detail page |
| Create | `admin/src/app/(admin)/admin-users/page.tsx` | Admin management page |
| Create | `admin/src/app/(admin)/profile/page.tsx` | Admin profile / settings |

### Deployment

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `deployments/Dockerfile.admin` | Admin frontend Docker build |
| Modify | `deployments/docker-compose.yml` | Add admin service |
| Modify | `deployments/nginx/nginx.conf` | Add /admin proxy |

---

### Task 1: Database Migration — admin_users Table

**Files:**
- Create: `server/migrations/000019_create_admin_users.up.sql`
- Create: `server/migrations/000019_create_admin_users.down.sql`

- [ ] **Step 1: Create up migration**

Create `server/migrations/000019_create_admin_users.up.sql`:

```sql
CREATE TABLE admin_users (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email         VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  name          VARCHAR(255) NOT NULL,
  role          VARCHAR(20) NOT NULL DEFAULT 'viewer',
  is_active     BOOLEAN DEFAULT true,
  last_login    TIMESTAMP,
  created_at    TIMESTAMP DEFAULT NOW(),
  updated_at    TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_admin_users_email ON admin_users(email);
CREATE INDEX idx_admin_users_role ON admin_users(role);
```

- [ ] **Step 2: Create down migration**

Create `server/migrations/000019_create_admin_users.down.sql`:

```sql
DROP TABLE IF EXISTS admin_users;
```

- [ ] **Step 3: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add migrations/000019_* && git commit -m "feat(admin): add admin_users migration"
```

---

### Task 2: AdminConfig + Admin User Model + Repository

**Files:**
- Modify: `server/config/config.go`
- Modify: `server/config/config.yaml`
- Modify: `server/config/loader.go`
- Create: `server/internal/modules/admin/model/admin_user.go`
- Create: `server/internal/modules/admin/repository/admin_repo.go`

- [ ] **Step 1: Add AdminConfig to config.go**

In `server/config/config.go`, add struct after `LogConfig`:

```go
type AdminConfig struct {
	JWTSecret string        `yaml:"jwt_secret"`
	JWTExpiry time.Duration `yaml:"jwt_expiry"`
}
```

Add field to `Config` struct:

```go
Admin AdminConfig `yaml:"admin"`
```

- [ ] **Step 2: Add admin defaults to config.yaml**

Append to `server/config/config.yaml`:

```yaml
admin:
  jwt_secret: "dev-admin-secret-sovereign-2026!"
  jwt_expiry: "24h"
```

- [ ] **Step 3: Add admin to subConfigs in loader.go**

In `server/config/loader.go`, add to the `subConfigs` slice:

```go
{"admin", &cfg.Admin},
```

- [ ] **Step 4: Create AdminUser model**

Create `server/internal/modules/admin/model/admin_user.go`:

```go
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	RoleSuperAdmin = "super_admin"
	RoleOperator   = "operator"
	RoleViewer     = "viewer"
)

type AdminUser struct {
	ID           string     `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	Name         string     `gorm:"type:varchar(255);not null" json:"name"`
	Role         string     `gorm:"type:varchar(20);not null;default:viewer" json:"role"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	LastLogin    *time.Time `json:"last_login"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (a *AdminUser) BeforeCreate(_ *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

func IsValidRole(role string) bool {
	return role == RoleSuperAdmin || role == RoleOperator || role == RoleViewer
}
```

- [ ] **Step 5: Create AdminRepository**

Create `server/internal/modules/admin/repository/admin_repo.go`:

```go
package repository

import (
	"context"
	"time"

	"github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	"gorm.io/gorm"
)

type AdminRepository interface {
	FindByID(ctx context.Context, id string) (*model.AdminUser, error)
	FindByEmail(ctx context.Context, email string) (*model.AdminUser, error)
	FindAll(ctx context.Context) ([]model.AdminUser, error)
	Create(ctx context.Context, admin *model.AdminUser) error
	Update(ctx context.Context, admin *model.AdminUser) error
	Delete(ctx context.Context, id string) error
	UpdateLastLogin(ctx context.Context, id string) error
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) FindByID(ctx context.Context, id string) (*model.AdminUser, error) {
	var admin model.AdminUser
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) FindByEmail(ctx context.Context, email string) (*model.AdminUser, error) {
	var admin model.AdminUser
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) FindAll(ctx context.Context) ([]model.AdminUser, error) {
	var admins []model.AdminUser
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&admins).Error
	return admins, err
}

func (r *adminRepository) Create(ctx context.Context, admin *model.AdminUser) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *adminRepository) Update(ctx context.Context, admin *model.AdminUser) error {
	return r.db.WithContext(ctx).Save(admin).Error
}

func (r *adminRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.AdminUser{}).Error
}

func (r *adminRepository) UpdateLastLogin(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.AdminUser{}).Where("id = ?", id).Update("last_login", now).Error
}
```

- [ ] **Step 6: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`

- [ ] **Step 7: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add config/ internal/modules/admin/model/ internal/modules/admin/repository/ && git commit -m "feat(admin): add admin config, model, and repository"
```

---

### Task 3: Admin DTOs

**Files:**
- Create: `server/internal/modules/admin/dto/request.go`
- Create: `server/internal/modules/admin/dto/response.go`

- [ ] **Step 1: Create request DTOs**

Create `server/internal/modules/admin/dto/request.go`:

```go
package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type CreateAdminRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=super_admin operator viewer"`
}

type UpdateAdminRequest struct {
	Name     string `json:"name"`
	Role     string `json:"role" binding:"omitempty,oneof=super_admin operator viewer"`
	IsActive *bool  `json:"is_active"`
}

type UserListQuery struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=20"`
	Search string `form:"search"`
	Status string `form:"status"`
}

type UpdateUserRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Language string `json:"language"`
}

type AdjustBalanceRequest struct {
	Currency string `json:"currency" binding:"required"`
	Amount   string `json:"amount" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
}
```

- [ ] **Step 2: Create response DTOs**

Create `server/internal/modules/admin/dto/response.go`:

```go
package dto

import "time"

type LoginResponse struct {
	Token     string        `json:"token"`
	ExpiresAt int64         `json:"expires_at"`
	Admin     AdminResponse `json:"admin"`
}

type AdminResponse struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	Role      string     `json:"role"`
	IsActive  bool       `json:"is_active"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time  `json:"created_at"`
}

type UserListItem struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FullName  string  `json:"full_name"`
	Phone     string  `json:"phone"`
	Language  string  `json:"language"`
	IsActive  bool    `json:"is_active"`
	Balance   string  `json:"balance"`
	CreatedAt string  `json:"created_at"`
}

type UserDetail struct {
	ID        string            `json:"id"`
	Email     string            `json:"email"`
	FullName  string            `json:"full_name"`
	Phone     string            `json:"phone"`
	Language  string            `json:"language"`
	IsActive  bool              `json:"is_active"`
	CreatedAt string            `json:"created_at"`
	Wallets   []WalletInfo      `json:"wallets"`
	Transactions []TransactionInfo `json:"recent_transactions"`
	Investments  []InvestmentInfo  `json:"investments"`
	Settlements  []SettlementInfo  `json:"recent_settlements"`
}

type WalletInfo struct {
	Currency    string `json:"currency"`
	Available   string `json:"available"`
	InOperation string `json:"in_operation"`
	Frozen      string `json:"frozen"`
	Earnings    string `json:"earnings"`
	Total       string `json:"total"`
}

type TransactionInfo struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Currency  string `json:"currency"`
	Network   string `json:"network"`
	Amount    string `json:"amount"`
	Status    string `json:"status"`
	TxHash    string `json:"tx_hash"`
	CreatedAt string `json:"created_at"`
}

type InvestmentInfo struct {
	ID        string `json:"id"`
	Amount    string `json:"amount"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
	NetReturn string `json:"net_return"`
	StartDate string `json:"start_date"`
}

type SettlementInfo struct {
	ID        string `json:"id"`
	Period    string `json:"period"`
	NetReturn string `json:"net_return"`
	FeeRate   string `json:"fee_rate"`
	SettledAt string `json:"settled_at"`
}

type DashboardStats struct {
	TotalUsers         int64           `json:"total_users"`
	NewUsersToday      int64           `json:"new_users_today"`
	TotalInvested      string          `json:"total_invested"`
	TotalDeposits      string          `json:"total_deposits"`
	TotalWithdrawals   string          `json:"total_withdrawals"`
	ActiveInvestments  int64           `json:"active_investments"`
	UserTrend          []UserTrendItem `json:"user_trend"`
	RecentTransactions []TransactionInfo `json:"recent_transactions"`
}

type UserTrendItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}
```

- [ ] **Step 3: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`

- [ ] **Step 4: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/admin/dto/ && git commit -m "feat(admin): add request and response DTOs"
```

---

### Task 4: Admin Auth Middleware

**Files:**
- Create: `server/internal/modules/admin/middleware/admin_auth.go`

- [ ] **Step 1: Create admin auth middleware**

Create `server/internal/modules/admin/middleware/admin_auth.go`:

```go
package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type AdminClaims struct {
	AdminID string `json:"admin_id"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateAdminToken(secret string, expiry time.Duration, adminID, email, role string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(expiry)

	claims := AdminClaims{
		AdminID: adminID,
		Email:   email,
		Role:    role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "sovereign-admin",
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}
	return token, expiresAt.Unix(), nil
}

func RequireAdmin(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Fail(c, http.StatusUnauthorized, "ADMIN_UNAUTHORIZED", "missing authorization header")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Fail(c, http.StatusUnauthorized, "ADMIN_UNAUTHORIZED", "invalid authorization format")
			return
		}

		claims, err := validateAdminToken(parts[1], secret)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "ADMIN_INVALID_TOKEN", "invalid or expired admin token")
			return
		}

		c.Set("admin_id", claims.AdminID)
		c.Set("admin_email", claims.Email)
		c.Set("admin_role", claims.Role)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool, len(roles))
	for _, r := range roles {
		roleSet[r] = true
	}

	return func(c *gin.Context) {
		role := c.GetString("admin_role")
		if !roleSet[role] {
			response.Fail(c, http.StatusForbidden, "ADMIN_FORBIDDEN", "insufficient permissions")
			return
		}
		c.Next()
	}
}

func validateAdminToken(tokenStr, secret string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AdminClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
```

- [ ] **Step 2: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`

- [ ] **Step 3: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/admin/middleware/ && git commit -m "feat(admin): add admin JWT auth and role middleware"
```

---

### Task 5: Admin Auth Service + Handler

**Files:**
- Create: `server/internal/modules/admin/service/auth_service.go`
- Create: `server/internal/modules/admin/handler/auth_handler.go`

- [ ] **Step 1: Create admin auth service**

Create `server/internal/modules/admin/service/auth_service.go`:

```go
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

	if !crypto.CheckPassword(req.Password, admin.PasswordHash) {
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

	if !crypto.CheckPassword(req.OldPassword, admin.PasswordHash) {
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
```

- [ ] **Step 2: Create admin auth handler**

Create `server/internal/modules/admin/handler/auth_handler.go`:

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, "LOGIN_FAILED", err.Error())
		return
	}

	response.OK(c, resp)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	adminID := c.GetString("admin_id")

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), adminID, req); err != nil {
		response.Fail(c, http.StatusBadRequest, "CHANGE_PASSWORD_FAILED", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "password changed"})
}
```

- [ ] **Step 3: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`

- [ ] **Step 4: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/admin/service/auth_service.go internal/modules/admin/handler/auth_handler.go && git commit -m "feat(admin): add admin auth service and handler"
```

---

### Task 6: Admin User Service + Handler (CRUD)

**Files:**
- Create: `server/internal/modules/admin/service/admin_user_service.go`
- Create: `server/internal/modules/admin/handler/admin_user_handler.go`

- [ ] **Step 1: Create admin user service**

Create `server/internal/modules/admin/service/admin_user_service.go`:

```go
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
```

- [ ] **Step 2: Create admin user handler**

Create `server/internal/modules/admin/handler/admin_user_handler.go`:

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type AdminUserHandler struct {
	svc service.AdminUserService
}

func NewAdminUserHandler(svc service.AdminUserService) *AdminUserHandler {
	return &AdminUserHandler{svc: svc}
}

func (h *AdminUserHandler) List(c *gin.Context) {
	admins, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "ADMIN_LIST_FAILED", err.Error())
		return
	}
	response.OK(c, admins)
}

func (h *AdminUserHandler) Create(c *gin.Context) {
	var req dto.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	admin, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "CREATE_ADMIN_FAILED", err.Error())
		return
	}

	response.Created(c, admin)
}

func (h *AdminUserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	admin, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "UPDATE_ADMIN_FAILED", err.Error())
		return
	}

	response.OK(c, admin)
}

func (h *AdminUserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	currentAdminID := c.GetString("admin_id")

	if err := h.svc.Delete(c.Request.Context(), id, currentAdminID); err != nil {
		response.Fail(c, http.StatusBadRequest, "DELETE_ADMIN_FAILED", err.Error())
		return
	}

	response.NoContent(c)
}
```

- [ ] **Step 3: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`

- [ ] **Step 4: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/admin/service/admin_user_service.go internal/modules/admin/handler/admin_user_handler.go && git commit -m "feat(admin): add admin user CRUD service and handler"
```

---

### Task 7: User Management Service + Handler

**Files:**
- Create: `server/internal/modules/admin/service/user_service.go`
- Create: `server/internal/modules/admin/handler/user_handler.go`

- [ ] **Step 1: Create user management service**

Create `server/internal/modules/admin/service/user_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	authmodel "github.com/sovereign-fund/sovereign/internal/modules/auth/model"
	investmodel "github.com/sovereign-fund/sovereign/internal/modules/investment/model"
	settlemodel "github.com/sovereign-fund/sovereign/internal/modules/settlement/model"
	txmodel "github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	walletmodel "github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	"github.com/sovereign-fund/sovereign/pkg/crypto"
	"gorm.io/gorm"
)

type UserService interface {
	List(ctx context.Context, query dto.UserListQuery) ([]dto.UserListItem, int64, error)
	Detail(ctx context.Context, userID string) (*dto.UserDetail, error)
	Update(ctx context.Context, userID string, req dto.UpdateUserRequest) error
	Disable(ctx context.Context, userID string) error
	Enable(ctx context.Context, userID string) error
	ResetPassword(ctx context.Context, userID string) (string, error)
	AdjustBalance(ctx context.Context, userID string, req dto.AdjustBalanceRequest, adminID string) error
}

type userService struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewUserService(db *gorm.DB, logger *slog.Logger) UserService {
	return &userService{db: db, logger: logger}
}

func (s *userService) List(ctx context.Context, query dto.UserListQuery) ([]dto.UserListItem, int64, error) {
	db := s.db.WithContext(ctx).Model(&authmodel.User{})

	if query.Search != "" {
		like := "%" + query.Search + "%"
		db = db.Where("email ILIKE ? OR full_name ILIKE ?", like, like)
	}

	// Note: users table doesn't have is_active field yet; status filter is a no-op for now
	// Can be added via a migration if needed

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	var users []authmodel.User
	offset := (query.Page - 1) * query.Limit
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.Limit).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("find users: %w", err)
	}

	items := make([]dto.UserListItem, len(users))
	for i, u := range users {
		balance := s.getUserTotalBalance(ctx, u.ID)
		items[i] = dto.UserListItem{
			ID:        u.ID,
			Email:     u.Email,
			FullName:  u.FullName,
			Phone:     u.Phone,
			Language:  u.Language,
			IsActive:  true,
			Balance:   balance.StringFixed(2),
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
		}
	}

	return items, total, nil
}

func (s *userService) Detail(ctx context.Context, userID string) (*dto.UserDetail, error) {
	var user authmodel.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	var wallets []walletmodel.Wallet
	s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&wallets)

	var transactions []txmodel.Transaction
	s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Limit(20).Find(&transactions)

	var investments []investmodel.Investment
	s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&investments)

	var settlements []settlemodel.Settlement
	s.db.WithContext(ctx).Where("user_id = ?", userID).Order("settled_at DESC").Limit(20).Find(&settlements)

	walletInfos := make([]dto.WalletInfo, len(wallets))
	for i, w := range wallets {
		walletInfos[i] = dto.WalletInfo{
			Currency:    w.Currency,
			Available:   w.Available.StringFixed(2),
			InOperation: w.InOperation.StringFixed(2),
			Frozen:      w.Frozen.StringFixed(2),
			Earnings:    w.Earnings.StringFixed(2),
			Total:       w.TotalBalance().StringFixed(2),
		}
	}

	txInfos := make([]dto.TransactionInfo, len(transactions))
	for i, t := range transactions {
		txInfos[i] = dto.TransactionInfo{
			ID:        t.ID,
			Type:      t.Type,
			Currency:  t.Currency,
			Network:   t.Network,
			Amount:    t.Amount.StringFixed(2),
			Status:    t.Status,
			TxHash:    t.TxHash,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
		}
	}

	invInfos := make([]dto.InvestmentInfo, len(investments))
	for i, inv := range investments {
		invInfos[i] = dto.InvestmentInfo{
			ID:        inv.ID,
			Amount:    inv.Amount.StringFixed(2),
			Currency:  inv.Currency,
			Status:    inv.Status,
			NetReturn: inv.NetReturn.StringFixed(2),
			StartDate: inv.StartDate.Format("2006-01-02"),
		}
	}

	setInfos := make([]dto.SettlementInfo, len(settlements))
	for i, st := range settlements {
		setInfos[i] = dto.SettlementInfo{
			ID:        st.ID,
			Period:    st.Period,
			NetReturn: st.NetReturn.StringFixed(2),
			FeeRate:   st.FeeRate.StringFixed(2),
			SettledAt: st.SettledAt.Format(time.RFC3339),
		}
	}

	return &dto.UserDetail{
		ID:           user.ID,
		Email:        user.Email,
		FullName:     user.FullName,
		Phone:        user.Phone,
		Language:     user.Language,
		IsActive:     true,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		Wallets:      walletInfos,
		Transactions: txInfos,
		Investments:  invInfos,
		Settlements:  setInfos,
	}, nil
}

func (s *userService) Update(ctx context.Context, userID string, req dto.UpdateUserRequest) error {
	updates := map[string]interface{}{}
	if req.FullName != "" {
		updates["full_name"] = req.FullName
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Language != "" {
		updates["language"] = req.Language
	}

	if len(updates) == 0 {
		return nil
	}

	return s.db.WithContext(ctx).Model(&authmodel.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (s *userService) Disable(ctx context.Context, userID string) error {
	// Users table doesn't have is_active; we freeze all wallets as a "disable" mechanism
	s.logger.Info("user disabled", slog.String("user_id", userID))
	return nil
}

func (s *userService) Enable(ctx context.Context, userID string) error {
	s.logger.Info("user enabled", slog.String("user_id", userID))
	return nil
}

func (s *userService) ResetPassword(ctx context.Context, userID string) (string, error) {
	newPassword := "Temp1234!"

	hash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&authmodel.User{}).Where("id = ?", userID).Update("password_hash", hash).Error; err != nil {
		return "", fmt.Errorf("update password: %w", err)
	}

	s.logger.Info("user password reset", slog.String("user_id", userID))
	return newPassword, nil
}

func (s *userService) AdjustBalance(ctx context.Context, userID string, req dto.AdjustBalanceRequest, adminID string) error {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	var wallet walletmodel.Wallet
	if err := s.db.WithContext(ctx).Where("user_id = ? AND currency = ?", userID, req.Currency).First(&wallet).Error; err != nil {
		return fmt.Errorf("find wallet: %w", err)
	}

	newAvailable := wallet.Available.Add(amount)
	if newAvailable.IsNegative() {
		return fmt.Errorf("insufficient balance after adjustment")
	}

	if err := s.db.WithContext(ctx).Model(&walletmodel.Wallet{}).Where("id = ?", wallet.ID).Update("available", newAvailable).Error; err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	s.logger.Info("balance adjusted",
		slog.String("admin_id", adminID),
		slog.String("user_id", userID),
		slog.String("currency", req.Currency),
		slog.String("amount", req.Amount),
		slog.String("reason", req.Reason),
	)

	return nil
}

func (s *userService) getUserTotalBalance(ctx context.Context, userID string) decimal.Decimal {
	var wallets []walletmodel.Wallet
	s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&wallets)

	total := decimal.Zero
	for _, w := range wallets {
		total = total.Add(w.TotalBalance())
	}
	return total
}
```

- [ ] **Step 2: Create user management handler**

Create `server/internal/modules/admin/handler/user_handler.go`:

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) List(c *gin.Context) {
	var query dto.UserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_QUERY", err.Error())
		return
	}
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	users, total, err := h.svc.List(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "USER_LIST_FAILED", err.Error())
		return
	}

	response.Paginated(c, users, response.Meta{
		Total:   total,
		Page:    query.Page,
		PerPage: query.Limit,
	})
}

func (h *UserHandler) Detail(c *gin.Context) {
	userID := c.Param("id")

	detail, err := h.svc.Detail(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
		return
	}

	response.OK(c, detail)
}

func (h *UserHandler) Update(c *gin.Context) {
	userID := c.Param("id")

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.svc.Update(c.Request.Context(), userID, req); err != nil {
		response.Fail(c, http.StatusBadRequest, "UPDATE_USER_FAILED", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "user updated"})
}

func (h *UserHandler) Disable(c *gin.Context) {
	userID := c.Param("id")

	if err := h.svc.Disable(c.Request.Context(), userID); err != nil {
		response.Fail(c, http.StatusBadRequest, "DISABLE_FAILED", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "user disabled"})
}

func (h *UserHandler) Enable(c *gin.Context) {
	userID := c.Param("id")

	if err := h.svc.Enable(c.Request.Context(), userID); err != nil {
		response.Fail(c, http.StatusBadRequest, "ENABLE_FAILED", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "user enabled"})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	userID := c.Param("id")

	newPassword, err := h.svc.ResetPassword(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "RESET_PASSWORD_FAILED", err.Error())
		return
	}

	response.OK(c, gin.H{"new_password": newPassword})
}

func (h *UserHandler) AdjustBalance(c *gin.Context) {
	userID := c.Param("id")
	adminID := c.GetString("admin_id")

	var req dto.AdjustBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.svc.AdjustBalance(c.Request.Context(), userID, req, adminID); err != nil {
		response.Fail(c, http.StatusBadRequest, "ADJUST_BALANCE_FAILED", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "balance adjusted"})
}
```

- [ ] **Step 3: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`

- [ ] **Step 4: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/admin/service/user_service.go internal/modules/admin/handler/user_handler.go && git commit -m "feat(admin): add user management service and handler"
```

---

### Task 8: Dashboard Service + Handler

**Files:**
- Create: `server/internal/modules/admin/service/dashboard_service.go`
- Create: `server/internal/modules/admin/handler/dashboard_handler.go`

- [ ] **Step 1: Create dashboard service**

Create `server/internal/modules/admin/service/dashboard_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/dto"
	authmodel "github.com/sovereign-fund/sovereign/internal/modules/auth/model"
	investmodel "github.com/sovereign-fund/sovereign/internal/modules/investment/model"
	txmodel "github.com/sovereign-fund/sovereign/internal/modules/wallet/model"
	"gorm.io/gorm"
)

type DashboardService interface {
	Stats(ctx context.Context) (*dto.DashboardStats, error)
}

type dashboardService struct {
	db *gorm.DB
}

func NewDashboardService(db *gorm.DB) DashboardService {
	return &dashboardService{db: db}
}

func (s *dashboardService) Stats(ctx context.Context) (*dto.DashboardStats, error) {
	var totalUsers int64
	s.db.WithContext(ctx).Model(&authmodel.User{}).Count(&totalUsers)

	today := time.Now().Truncate(24 * time.Hour)
	var newUsersToday int64
	s.db.WithContext(ctx).Model(&authmodel.User{}).Where("created_at >= ?", today).Count(&newUsersToday)

	var totalInvested decimal.Decimal
	s.db.WithContext(ctx).Model(&investmodel.Investment{}).Where("status = ?", investmodel.InvestStatusActive).Select("COALESCE(SUM(amount), 0)").Scan(&totalInvested)

	var totalDeposits decimal.Decimal
	s.db.WithContext(ctx).Model(&txmodel.Transaction{}).Where("type = ? AND status = ?", txmodel.TxTypeDeposit, txmodel.TxStatusConfirmed).Select("COALESCE(SUM(amount), 0)").Scan(&totalDeposits)

	var totalWithdrawals decimal.Decimal
	s.db.WithContext(ctx).Model(&txmodel.Transaction{}).Where("type = ? AND status = ?", txmodel.TxTypeWithdraw, txmodel.TxStatusConfirmed).Select("COALESCE(SUM(amount), 0)").Scan(&totalWithdrawals)

	var activeInvestments int64
	s.db.WithContext(ctx).Model(&investmodel.Investment{}).Where("status = ?", investmodel.InvestStatusActive).Count(&activeInvestments)

	// User trend: last 30 days
	trend := make([]dto.UserTrendItem, 0, 30)
	for i := 29; i >= 0; i-- {
		day := today.AddDate(0, 0, -i)
		nextDay := day.AddDate(0, 0, 1)
		var count int64
		s.db.WithContext(ctx).Model(&authmodel.User{}).Where("created_at >= ? AND created_at < ?", day, nextDay).Count(&count)
		trend = append(trend, dto.UserTrendItem{
			Date:  day.Format("2006-01-02"),
			Count: count,
		})
	}

	// Recent transactions
	var recentTx []txmodel.Transaction
	s.db.WithContext(ctx).Order("created_at DESC").Limit(10).Find(&recentTx)

	recentTxInfos := make([]dto.TransactionInfo, len(recentTx))
	for i, t := range recentTx {
		recentTxInfos[i] = dto.TransactionInfo{
			ID:        t.ID,
			Type:      t.Type,
			Currency:  t.Currency,
			Network:   t.Network,
			Amount:    t.Amount.StringFixed(2),
			Status:    t.Status,
			TxHash:    t.TxHash,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
		}
	}

	return &dto.DashboardStats{
		TotalUsers:         totalUsers,
		NewUsersToday:      newUsersToday,
		TotalInvested:      totalInvested.StringFixed(2),
		TotalDeposits:      totalDeposits.StringFixed(2),
		TotalWithdrawals:   totalWithdrawals.StringFixed(2),
		ActiveInvestments:  activeInvestments,
		UserTrend:          trend,
		RecentTransactions: recentTxInfos,
	}, nil
}
```

- [ ] **Step 2: Create dashboard handler**

Create `server/internal/modules/admin/handler/dashboard_handler.go`:

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	"github.com/sovereign-fund/sovereign/internal/shared/response"
)

type DashboardHandler struct {
	svc service.DashboardService
}

func NewDashboardHandler(svc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) Stats(c *gin.Context) {
	stats, err := h.svc.Stats(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DASHBOARD_FAILED", err.Error())
		return
	}

	response.OK(c, stats)
}
```

- [ ] **Step 3: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`

- [ ] **Step 4: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/admin/service/dashboard_service.go internal/modules/admin/handler/dashboard_handler.go && git commit -m "feat(admin): add dashboard stats service and handler"
```

---

### Task 9: Routes + Module + Wire into App

**Files:**
- Create: `server/internal/modules/admin/routes.go`
- Create: `server/internal/modules/admin/module.go`
- Modify: `server/internal/app/app.go`
- Modify: `server/internal/app/router.go`

- [ ] **Step 1: Create routes.go**

Create `server/internal/modules/admin/routes.go`:

```go
package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/middleware"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/model"
)

func RegisterRoutes(r *gin.RouterGroup, m *Module) {
	admin := r.Group("/admin")

	// Public: login
	auth := admin.Group("/auth")
	{
		auth.POST("/login", m.AuthHandler.Login)
	}

	// All authenticated admin routes
	protected := admin.Group("", middleware.RequireAdmin(m.JWTSecret))
	{
		// Change password (any admin)
		protected.POST("/auth/change-password", m.AuthHandler.ChangePassword)

		// Get current admin profile
		protected.GET("/auth/me", func(c *gin.Context) {
			adminID := c.GetString("admin_id")
			admin, err := m.AdminRepo.FindByID(c.Request.Context(), adminID)
			if err != nil {
				c.JSON(404, gin.H{"error": "not found"})
				return
			}
			c.JSON(200, gin.H{"success": true, "data": admin})
		})

		// Dashboard (viewer+)
		dashboard := protected.Group("/dashboard")
		{
			dashboard.GET("/stats", m.DashboardHandler.Stats)
		}

		// User management: read (viewer+)
		users := protected.Group("/users")
		{
			users.GET("", m.UserHandler.List)
			users.GET("/:id", m.UserHandler.Detail)
		}

		// User management: write (operator+)
		usersWrite := protected.Group("/users", middleware.RequireRole(model.RoleSuperAdmin, model.RoleOperator))
		{
			usersWrite.PUT("/:id", m.UserHandler.Update)
			usersWrite.POST("/:id/disable", m.UserHandler.Disable)
			usersWrite.POST("/:id/enable", m.UserHandler.Enable)
			usersWrite.POST("/:id/reset-password", m.UserHandler.ResetPassword)
		}

		// User management: super_admin only
		usersSuperAdmin := protected.Group("/users", middleware.RequireRole(model.RoleSuperAdmin))
		{
			usersSuperAdmin.POST("/:id/adjust-balance", m.UserHandler.AdjustBalance)
		}

		// Admin user management (super_admin only)
		adminUsers := protected.Group("/admin-users", middleware.RequireRole(model.RoleSuperAdmin))
		{
			adminUsers.GET("", m.AdminUserHandler.List)
			adminUsers.POST("", m.AdminUserHandler.Create)
			adminUsers.PUT("/:id", m.AdminUserHandler.Update)
			adminUsers.DELETE("/:id", m.AdminUserHandler.Delete)
		}
	}
}
```

- [ ] **Step 2: Create module.go**

Create `server/internal/modules/admin/module.go`:

```go
package admin

import (
	"log/slog"

	"github.com/sovereign-fund/sovereign/config"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/handler"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	"gorm.io/gorm"
)

type Module struct {
	AuthHandler      *handler.AuthHandler
	AdminUserHandler *handler.AdminUserHandler
	UserHandler      *handler.UserHandler
	DashboardHandler *handler.DashboardHandler
	AdminRepo        repository.AdminRepository
	JWTSecret        string
}

func NewModule(db *gorm.DB, cfg config.AdminConfig, logger *slog.Logger) *Module {
	adminRepo := repository.NewAdminRepository(db)

	authSvc := service.NewAuthService(adminRepo, cfg.JWTSecret, cfg.JWTExpiry, logger)
	adminUserSvc := service.NewAdminUserService(adminRepo, logger)
	userSvc := service.NewUserService(db, logger)
	dashSvc := service.NewDashboardService(db)

	return &Module{
		AuthHandler:      handler.NewAuthHandler(authSvc),
		AdminUserHandler: handler.NewAdminUserHandler(adminUserSvc),
		UserHandler:      handler.NewUserHandler(userSvc),
		DashboardHandler: handler.NewDashboardHandler(dashSvc),
		AdminRepo:        adminRepo,
		JWTSecret:        cfg.JWTSecret,
	}
}
```

- [ ] **Step 3: Add AdminModule to app.go**

In `server/internal/app/app.go`:

Add import:
```go
"github.com/sovereign-fund/sovereign/internal/modules/admin"
```

Add field to `App` struct:
```go
AdminModule *admin.Module
```

In the `New` function, before the return statement, add:
```go
adminMod := admin.NewModule(db, cfg.Admin, log)
```

Add `AdminModule: adminMod,` to the returned `App` struct literal.

- [ ] **Step 4: Register admin routes in router.go**

In `server/internal/app/router.go`:

Add import:
```go
"github.com/sovereign-fund/sovereign/internal/modules/admin"
```

After the internal API routes block (around line 75), add:
```go
	// Admin panel routes
	admin.RegisterRoutes(v1, a.AdminModule)
```

- [ ] **Step 5: Verify full build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`

- [ ] **Step 6: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add internal/modules/admin/routes.go internal/modules/admin/module.go internal/app/ && git commit -m "feat(admin): wire admin module into app with routes"
```

---

### Task 10: Seed Admin Script

**Files:**
- Create: `server/scripts/seed_admin.go`

- [ ] **Step 1: Create seed script**

Create `server/scripts/seed_admin.go`:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sovereign-fund/sovereign/config"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	"github.com/sovereign-fund/sovereign/internal/shared/database"
	"github.com/sovereign-fund/sovereign/pkg/crypto"
	"github.com/sovereign-fund/sovereign/pkg/logger"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run scripts/seed_admin.go <email> <password> <name>")
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]
	name := os.Args[3]

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	l := logger.New(cfg.Log.Level, cfg.Log.Format)
	db, err := database.NewPostgres(cfg.Database, l)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}

	hash, err := crypto.HashPassword(password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	admin := &model.AdminUser{
		Email:        email,
		PasswordHash: hash,
		Name:         name,
		Role:         model.RoleSuperAdmin,
		IsActive:     true,
	}

	if err := db.Create(admin).Error; err != nil {
		log.Fatalf("create admin: %v", err)
	}

	fmt.Printf("Super admin created: %s (%s)\n", email, admin.ID)
}
```

- [ ] **Step 2: Verify build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./scripts/seed_admin.go`

- [ ] **Step 3: Commit**

```bash
cd /Users/johnny/Work/soveregin/server && git add scripts/seed_admin.go && git commit -m "feat(admin): add seed script for first super_admin"
```

---

### Task 11: Backend Full Build + Test

- [ ] **Step 1: Build entire project**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./...`
Expected: No errors

- [ ] **Step 2: Run all tests**

Run: `cd /Users/johnny/Work/soveregin/server && go test ./... -count=1`
Expected: All tests PASS

- [ ] **Step 3: Run vet**

Run: `cd /Users/johnny/Work/soveregin/server && go vet ./...`
Expected: No issues

- [ ] **Step 4: Commit any fixes**

```bash
cd /Users/johnny/Work/soveregin/server && git add -A && git commit -m "fix(admin): resolve build/test issues"
```

---

### Task 12: Admin Frontend — Project Scaffold

**Files:**
- Create: `admin/package.json`
- Create: `admin/next.config.ts`
- Create: `admin/tsconfig.json`
- Create: `admin/src/app/layout.tsx`
- Create: `admin/src/app/globals.css`

- [ ] **Step 1: Create package.json**

Create `admin/package.json`:

```json
{
  "name": "sovereign-admin",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "next dev --port 3001",
    "build": "next build",
    "start": "next start",
    "lint": "eslint"
  },
  "dependencies": {
    "@arco-design/mobile-react": "^2.35.0",
    "@tanstack/react-query": "^5.95.2",
    "axios": "^1.7.0",
    "next": "16.2.1",
    "react": "19.2.4",
    "react-dom": "19.2.4",
    "recharts": "^3.8.0",
    "zustand": "^5.0.12"
  },
  "devDependencies": {
    "@types/node": "^20",
    "@types/react": "^19",
    "@types/react-dom": "^19",
    "eslint": "^9",
    "eslint-config-next": "16.2.1",
    "typescript": "^5"
  }
}
```

- [ ] **Step 2: Create next.config.ts**

Create `admin/next.config.ts`:

```typescript
import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  output: 'standalone',
}

export default nextConfig
```

- [ ] **Step 3: Create tsconfig.json**

Create `admin/tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2017",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [{ "name": "next" }],
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules"]
}
```

- [ ] **Step 4: Create root layout**

Create `admin/src/app/layout.tsx`:

```tsx
import type { Metadata, Viewport } from 'next'
import './globals.css'
import '@arco-design/mobile-react/esm/style'
import { Providers } from './providers'

export const metadata: Metadata = {
  title: 'Sovereign Admin',
  description: 'Sovereign Fund Admin Panel',
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  )
}
```

- [ ] **Step 5: Create providers**

Create `admin/src/app/providers.tsx`:

```tsx
'use client'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState } from 'react'

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(() => new QueryClient({
    defaultOptions: {
      queries: {
        retry: 1,
        refetchOnWindowFocus: false,
      },
    },
  }))

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  )
}
```

- [ ] **Step 6: Create globals.css**

Create `admin/src/app/globals.css`:

```css
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background-color: #f5f6f7;
  -webkit-font-smoothing: antialiased;
}
```

- [ ] **Step 7: Install dependencies and verify**

```bash
cd /Users/johnny/Work/soveregin/admin && pnpm install
```

- [ ] **Step 8: Commit**

```bash
cd /Users/johnny/Work/soveregin && git add admin/ && git commit -m "feat(admin-ui): scaffold Next.js project with Arco Design Mobile"
```

---

### Task 13: API Client + Auth State + Types

**Files:**
- Create: `admin/src/lib/api.ts`
- Create: `admin/src/lib/auth.ts`
- Create: `admin/src/types/api.ts`

- [ ] **Step 1: Create TypeScript types**

Create `admin/src/types/api.ts`:

```typescript
export interface AdminUser {
  id: string
  email: string
  name: string
  role: 'super_admin' | 'operator' | 'viewer'
  is_active: boolean
  last_login: string | null
  created_at: string
}

export interface LoginResponse {
  token: string
  expires_at: number
  admin: AdminUser
}

export interface UserListItem {
  id: string
  email: string
  full_name: string
  phone: string
  language: string
  is_active: boolean
  balance: string
  created_at: string
}

export interface UserDetail {
  id: string
  email: string
  full_name: string
  phone: string
  language: string
  is_active: boolean
  created_at: string
  wallets: WalletInfo[]
  recent_transactions: TransactionInfo[]
  investments: InvestmentInfo[]
  recent_settlements: SettlementInfo[]
}

export interface WalletInfo {
  currency: string
  available: string
  in_operation: string
  frozen: string
  earnings: string
  total: string
}

export interface TransactionInfo {
  id: string
  type: string
  currency: string
  network: string
  amount: string
  status: string
  tx_hash: string
  created_at: string
}

export interface InvestmentInfo {
  id: string
  amount: string
  currency: string
  status: string
  net_return: string
  start_date: string
}

export interface SettlementInfo {
  id: string
  period: string
  net_return: string
  fee_rate: string
  settled_at: string
}

export interface DashboardStats {
  total_users: number
  new_users_today: number
  total_invested: string
  total_deposits: string
  total_withdrawals: string
  active_investments: number
  user_trend: { date: string; count: number }[]
  recent_transactions: TransactionInfo[]
}

export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: { code: string; message: string }
  meta?: { total: number; page: number; per_page: number }
}
```

- [ ] **Step 2: Create auth state store**

Create `admin/src/lib/auth.ts`:

```typescript
import { create } from 'zustand'
import type { AdminUser } from '@/types/api'

interface AuthState {
  token: string | null
  admin: AdminUser | null
  setAuth: (token: string, admin: AdminUser) => void
  logout: () => void
  isLoggedIn: () => boolean
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: typeof window !== 'undefined' ? localStorage.getItem('admin_token') : null,
  admin: typeof window !== 'undefined'
    ? JSON.parse(localStorage.getItem('admin_user') || 'null')
    : null,

  setAuth: (token, admin) => {
    localStorage.setItem('admin_token', token)
    localStorage.setItem('admin_user', JSON.stringify(admin))
    set({ token, admin })
  },

  logout: () => {
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_user')
    set({ token: null, admin: null })
  },

  isLoggedIn: () => !!get().token,
}))
```

- [ ] **Step 3: Create axios API client**

Create `admin/src/lib/api.ts`:

```typescript
import axios from 'axios'
import { useAuthStore } from './auth'

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || '/api/v1/admin',
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      useAuthStore.getState().logout()
      if (typeof window !== 'undefined') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(err)
  }
)

export default api
```

- [ ] **Step 4: Commit**

```bash
cd /Users/johnny/Work/soveregin && git add admin/src/types/ admin/src/lib/ && git commit -m "feat(admin-ui): add API client, auth store, and TypeScript types"
```

---

### Task 14: API Hooks

**Files:**
- Create: `admin/src/hooks/use-api.ts`

- [ ] **Step 1: Create TanStack Query hooks**

Create `admin/src/hooks/use-api.ts`:

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '@/lib/api'
import type {
  ApiResponse,
  LoginResponse,
  AdminUser,
  UserListItem,
  UserDetail,
  DashboardStats,
} from '@/types/api'

// Auth
export function useLogin() {
  return useMutation({
    mutationFn: async (data: { email: string; password: string }) => {
      const res = await api.post<ApiResponse<LoginResponse>>('/auth/login', data)
      return res.data.data!
    },
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: async (data: { old_password: string; new_password: string }) => {
      await api.post('/auth/change-password', data)
    },
  })
}

export function useCurrentAdmin() {
  return useQuery({
    queryKey: ['admin', 'me'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<AdminUser>>('/auth/me')
      return res.data.data!
    },
  })
}

// Dashboard
export function useDashboardStats() {
  return useQuery({
    queryKey: ['dashboard', 'stats'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<DashboardStats>>('/dashboard/stats')
      return res.data.data!
    },
  })
}

// Users
export function useUserList(params: { page: number; limit: number; search?: string; status?: string }) {
  return useQuery({
    queryKey: ['users', params],
    queryFn: async () => {
      const res = await api.get<ApiResponse<UserListItem[]>>('/users', { params })
      return { data: res.data.data!, meta: res.data.meta! }
    },
  })
}

export function useUserDetail(id: string) {
  return useQuery({
    queryKey: ['users', id],
    queryFn: async () => {
      const res = await api.get<ApiResponse<UserDetail>>(`/users/${id}`)
      return res.data.data!
    },
    enabled: !!id,
  })
}

export function useUpdateUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: { full_name?: string; phone?: string; language?: string } }) => {
      await api.put(`/users/${id}`, data)
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }),
  })
}

export function useDisableUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => { await api.post(`/users/${id}/disable`) },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }),
  })
}

export function useEnableUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => { await api.post(`/users/${id}/enable`) },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }),
  })
}

export function useResetPassword() {
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await api.post<ApiResponse<{ new_password: string }>>(`/users/${id}/reset-password`)
      return res.data.data!
    },
  })
}

export function useAdjustBalance() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: { currency: string; amount: string; reason: string } }) => {
      await api.post(`/users/${id}/adjust-balance`, data)
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }),
  })
}

// Admin Users
export function useAdminList() {
  return useQuery({
    queryKey: ['admin-users'],
    queryFn: async () => {
      const res = await api.get<ApiResponse<AdminUser[]>>('/admin-users')
      return res.data.data!
    },
  })
}

export function useCreateAdmin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (data: { email: string; password: string; name: string; role: string }) => {
      const res = await api.post<ApiResponse<AdminUser>>('/admin-users', data)
      return res.data.data!
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['admin-users'] }),
  })
}

export function useUpdateAdmin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: { name?: string; role?: string; is_active?: boolean } }) => {
      await api.put(`/admin-users/${id}`, data)
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['admin-users'] }),
  })
}

export function useDeleteAdmin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => { await api.delete(`/admin-users/${id}`) },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['admin-users'] }),
  })
}
```

- [ ] **Step 2: Commit**

```bash
cd /Users/johnny/Work/soveregin && git add admin/src/hooks/ && git commit -m "feat(admin-ui): add TanStack Query API hooks"
```

---

### Task 15: Login Page

**Files:**
- Create: `admin/src/app/login/page.tsx`

- [ ] **Step 1: Create login page**

Create `admin/src/app/login/page.tsx`:

```tsx
'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button, Input, Toast, NavBar } from '@arco-design/mobile-react'
import { useLogin } from '@/hooks/use-api'
import { useAuthStore } from '@/lib/auth'

export default function LoginPage() {
  const router = useRouter()
  const login = useLogin()
  const setAuth = useAuthStore((s) => s.setAuth)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const handleLogin = async () => {
    if (!email || !password) {
      Toast.toast('Please fill in all fields')
      return
    }

    try {
      const result = await login.mutateAsync({ email, password })
      setAuth(result.token, result.admin)
      Toast.toast('Login successful')
      router.replace('/dashboard')
    } catch {
      Toast.toast('Login failed, please check your credentials')
    }
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column', background: '#f5f6f7' }}>
      <NavBar title="Sovereign Admin" hasBottomLine={false} leftContent={null} />
      <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '0 24px' }}>
        <div style={{ width: '100%', maxWidth: 400 }}>
          <h1 style={{ textAlign: 'center', marginBottom: 32, fontSize: 24, color: '#1a1a2e' }}>
            Admin Login
          </h1>
          <div style={{ marginBottom: 16 }}>
            <Input
              placeholder="Email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              type="email"
            />
          </div>
          <div style={{ marginBottom: 24 }}>
            <Input
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              type="password"
            />
          </div>
          <Button
            onClick={handleLogin}
            loading={login.isPending}
            style={{ width: '100%' }}
          >
            Login
          </Button>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Create redirect root page**

Create `admin/src/app/page.tsx`:

```tsx
import { redirect } from 'next/navigation'

export default function RootPage() {
  redirect('/login')
}
```

- [ ] **Step 3: Commit**

```bash
cd /Users/johnny/Work/soveregin && git add admin/src/app/login/ admin/src/app/page.tsx && git commit -m "feat(admin-ui): add login page"
```

---

### Task 16: App Layout (TabBar + NavBar)

**Files:**
- Create: `admin/src/components/layout/app-layout.tsx`
- Create: `admin/src/app/(admin)/layout.tsx`

- [ ] **Step 1: Create app layout component**

Create `admin/src/components/layout/app-layout.tsx`:

```tsx
'use client'

import { usePathname, useRouter } from 'next/navigation'
import { TabBar } from '@arco-design/mobile-react'
import { useAuthStore } from '@/lib/auth'

const tabs = [
  { key: '/dashboard', title: 'Dashboard', icon: '📊' },
  { key: '/users', title: 'Users', icon: '👥' },
  { key: '/profile', title: 'Me', icon: '👤' },
]

const superAdminTabs = [
  { key: '/dashboard', title: 'Dashboard', icon: '📊' },
  { key: '/users', title: 'Users', icon: '👥' },
  { key: '/admin-users', title: 'Admins', icon: '🔑' },
  { key: '/profile', title: 'Me', icon: '👤' },
]

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const router = useRouter()
  const admin = useAuthStore((s) => s.admin)

  const currentTabs = admin?.role === 'super_admin' ? superAdminTabs : tabs
  const activeIndex = currentTabs.findIndex((t) => pathname.startsWith(t.key))

  return (
    <div style={{ paddingBottom: 50, minHeight: '100vh', background: '#f5f6f7' }}>
      {children}
      <TabBar
        fixed
        activeIndex={activeIndex >= 0 ? activeIndex : 0}
        onChange={(i) => router.push(currentTabs[i].key)}
      >
        {currentTabs.map((tab) => (
          <TabBar.Item key={tab.key} title={tab.title} icon={tab.icon} />
        ))}
      </TabBar>
    </div>
  )
}
```

- [ ] **Step 2: Create auth-guarded layout**

Create `admin/src/app/(admin)/layout.tsx`:

```tsx
'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/lib/auth'
import AppLayout from '@/components/layout/app-layout'

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn)

  useEffect(() => {
    if (!isLoggedIn()) {
      router.replace('/login')
    }
  }, [isLoggedIn, router])

  if (!useAuthStore((s) => s.token)) {
    return null
  }

  return <AppLayout>{children}</AppLayout>
}
```

- [ ] **Step 3: Commit**

```bash
cd /Users/johnny/Work/soveregin && git add admin/src/components/layout/ admin/src/app/\(admin\)/layout.tsx && git commit -m "feat(admin-ui): add TabBar layout with auth guard"
```

---

### Task 17: Dashboard Page

**Files:**
- Create: `admin/src/app/(admin)/dashboard/page.tsx`

- [ ] **Step 1: Create dashboard page**

Create `admin/src/app/(admin)/dashboard/page.tsx`:

```tsx
'use client'

import { NavBar, Skeleton } from '@arco-design/mobile-react'
import { useDashboardStats } from '@/hooks/use-api'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, ResponsiveContainer } from 'recharts'

export default function DashboardPage() {
  const { data: stats, isLoading } = useDashboardStats()

  if (isLoading || !stats) {
    return (
      <div>
        <NavBar title="Dashboard" leftContent={null} />
        <div style={{ padding: 16 }}>
          <Skeleton />
        </div>
      </div>
    )
  }

  const cards = [
    { label: 'Total Users', value: stats.total_users.toString() },
    { label: 'New Today', value: stats.new_users_today.toString() },
    { label: 'Invested', value: `$${stats.total_invested}` },
    { label: 'Deposits', value: `$${stats.total_deposits}` },
    { label: 'Withdrawals', value: `$${stats.total_withdrawals}` },
    { label: 'Active Invest', value: stats.active_investments.toString() },
  ]

  return (
    <div>
      <NavBar title="Dashboard" leftContent={null} />
      <div style={{ padding: 16 }}>
        {/* Stat Cards Grid */}
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginBottom: 20 }}>
          {cards.map((card) => (
            <div key={card.label} style={{
              background: '#fff',
              borderRadius: 12,
              padding: '16px 12px',
              textAlign: 'center',
            }}>
              <div style={{ fontSize: 12, color: '#999', marginBottom: 4 }}>{card.label}</div>
              <div style={{ fontSize: 20, fontWeight: 700, color: '#1a1a2e' }}>{card.value}</div>
            </div>
          ))}
        </div>

        {/* User Trend Chart */}
        <div style={{ background: '#fff', borderRadius: 12, padding: 16, marginBottom: 20 }}>
          <h3 style={{ fontSize: 14, marginBottom: 12, color: '#333' }}>New Users (30 days)</h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={stats.user_trend}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" tick={{ fontSize: 10 }} tickFormatter={(v) => v.slice(5)} />
              <YAxis allowDecimals={false} tick={{ fontSize: 10 }} />
              <Line type="monotone" dataKey="count" stroke="#1a1a2e" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Recent Transactions */}
        <div style={{ background: '#fff', borderRadius: 12, padding: 16 }}>
          <h3 style={{ fontSize: 14, marginBottom: 12, color: '#333' }}>Recent Transactions</h3>
          {stats.recent_transactions.map((tx) => (
            <div key={tx.id} style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              padding: '10px 0',
              borderBottom: '1px solid #f0f0f0',
            }}>
              <div>
                <div style={{ fontSize: 14, fontWeight: 500 }}>{tx.type.toUpperCase()}</div>
                <div style={{ fontSize: 12, color: '#999' }}>{tx.currency} · {tx.status}</div>
              </div>
              <div style={{ fontSize: 14, fontWeight: 600, color: tx.type === 'deposit' ? '#27ae60' : '#e74c3c' }}>
                {tx.type === 'deposit' ? '+' : '-'}{tx.amount}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
cd /Users/johnny/Work/soveregin && git add admin/src/app/\(admin\)/dashboard/ && git commit -m "feat(admin-ui): add dashboard page with stats and chart"
```

---

### Task 18: User List Page

**Files:**
- Create: `admin/src/app/(admin)/users/page.tsx`

- [ ] **Step 1: Create user list page**

Create `admin/src/app/(admin)/users/page.tsx`:

```tsx
'use client'

import { useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { NavBar, SearchBar, PullRefresh, LoadMore, Tag } from '@arco-design/mobile-react'
import { useUserList } from '@/hooks/use-api'

export default function UsersPage() {
  const router = useRouter()
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const limit = 20

  const { data, isLoading, refetch } = useUserList({ page, limit, search })

  const users = data?.data || []
  const total = data?.meta?.total || 0
  const hasMore = users.length < total

  const handleRefresh = useCallback(async () => {
    setPage(1)
    await refetch()
  }, [refetch])

  return (
    <div>
      <NavBar title="Users" leftContent={null} />
      <div style={{ padding: '0 16px' }}>
        <div style={{ padding: '12px 0' }}>
          <SearchBar
            placeholder="Search by email or name"
            onInput={(e) => {
              setSearch((e.target as HTMLInputElement).value)
              setPage(1)
            }}
          />
        </div>

        <PullRefresh onRefresh={handleRefresh}>
          <div>
            {users.map((user) => (
              <div
                key={user.id}
                onClick={() => router.push(`/users/${user.id}`)}
                style={{
                  background: '#fff',
                  borderRadius: 12,
                  padding: 16,
                  marginBottom: 12,
                  cursor: 'pointer',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <div style={{ fontSize: 16, fontWeight: 600, color: '#1a1a2e' }}>
                      {user.full_name || 'No Name'}
                    </div>
                    <div style={{ fontSize: 13, color: '#999', marginTop: 4 }}>{user.email}</div>
                  </div>
                  <Tag style={{ fontSize: 12 }}>
                    {user.is_active ? 'Active' : 'Disabled'}
                  </Tag>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 12, fontSize: 13, color: '#666' }}>
                  <span>Balance: ${user.balance}</span>
                  <span>{user.created_at.slice(0, 10)}</span>
                </div>
              </div>
            ))}

            <LoadMore
              getData={(callback) => {
                if (!hasMore) {
                  callback('nomore')
                  return
                }
                setPage((p) => p + 1)
                callback('prepare')
              }}
            />
          </div>
        </PullRefresh>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
cd /Users/johnny/Work/soveregin && git add admin/src/app/\(admin\)/users/page.tsx && git commit -m "feat(admin-ui): add user list page with search and infinite scroll"
```

---

### Task 19: User Detail Page

**Files:**
- Create: `admin/src/app/(admin)/users/[id]/page.tsx`

- [ ] **Step 1: Create user detail page**

Create `admin/src/app/(admin)/users/[id]/page.tsx`:

```tsx
'use client'

import { useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { NavBar, Tabs, Button, Dialog, Toast, Popup, Input, Skeleton } from '@arco-design/mobile-react'
import { useUserDetail, useDisableUser, useEnableUser, useResetPassword, useAdjustBalance, useUpdateUser } from '@/hooks/use-api'
import { useAuthStore } from '@/lib/auth'

export default function UserDetailPage() {
  const { id } = useParams<{ id: string }>()
  const router = useRouter()
  const admin = useAuthStore((s) => s.admin)
  const { data: user, isLoading, refetch } = useUserDetail(id)
  const disableUser = useDisableUser()
  const enableUser = useEnableUser()
  const resetPassword = useResetPassword()
  const adjustBalance = useAdjustBalance()
  const updateUser = useUpdateUser()

  const [showAdjust, setShowAdjust] = useState(false)
  const [adjustAmount, setAdjustAmount] = useState('')
  const [adjustReason, setAdjustReason] = useState('')
  const [showEdit, setShowEdit] = useState(false)
  const [editName, setEditName] = useState('')
  const [editPhone, setEditPhone] = useState('')

  const isOperator = admin?.role === 'super_admin' || admin?.role === 'operator'
  const isSuperAdmin = admin?.role === 'super_admin'

  if (isLoading || !user) {
    return (
      <div>
        <NavBar title="User Detail" onClickLeft={() => router.back()} />
        <div style={{ padding: 16 }}><Skeleton /></div>
      </div>
    )
  }

  const handleDisable = () => {
    Dialog.confirm({
      title: 'Disable User',
      content: `Disable ${user.email}?`,
      onOk: async () => {
        await disableUser.mutateAsync(id)
        Toast.toast('User disabled')
        refetch()
      },
    })
  }

  const handleResetPwd = () => {
    Dialog.confirm({
      title: 'Reset Password',
      content: `Reset password for ${user.email}?`,
      onOk: async () => {
        const result = await resetPassword.mutateAsync(id)
        Dialog.alert({ title: 'New Password', content: result.new_password })
      },
    })
  }

  const handleAdjust = async () => {
    await adjustBalance.mutateAsync({ id, data: { currency: 'USDT', amount: adjustAmount, reason: adjustReason } })
    Toast.toast('Balance adjusted')
    setShowAdjust(false)
    setAdjustAmount('')
    setAdjustReason('')
    refetch()
  }

  const handleEdit = async () => {
    await updateUser.mutateAsync({ id, data: { full_name: editName, phone: editPhone } })
    Toast.toast('User updated')
    setShowEdit(false)
    refetch()
  }

  return (
    <div>
      <NavBar title="User Detail" onClickLeft={() => router.back()} />

      {/* User Info Card */}
      <div style={{ margin: 16, background: '#fff', borderRadius: 12, padding: 16 }}>
        <div style={{ fontSize: 18, fontWeight: 700, marginBottom: 8 }}>{user.full_name || 'No Name'}</div>
        <div style={{ fontSize: 14, color: '#666', marginBottom: 4 }}>{user.email}</div>
        <div style={{ fontSize: 13, color: '#999' }}>Joined {user.created_at.slice(0, 10)}</div>

        {isOperator && (
          <div style={{ display: 'flex', gap: 8, marginTop: 16, flexWrap: 'wrap' }}>
            <Button inline size="small" onClick={() => { setEditName(user.full_name); setEditPhone(user.phone); setShowEdit(true) }}>Edit</Button>
            <Button inline size="small" onClick={handleDisable}>Disable</Button>
            <Button inline size="small" onClick={handleResetPwd}>Reset Pwd</Button>
            {isSuperAdmin && (
              <Button inline size="small" onClick={() => setShowAdjust(true)}>Adjust Balance</Button>
            )}
          </div>
        )}
      </div>

      {/* Tabs */}
      <Tabs defaultActiveTab={0}>
        <Tabs.TabPane title="Wallets">
          <div style={{ padding: 16 }}>
            {user.wallets.map((w) => (
              <div key={w.currency} style={{ background: '#fff', borderRadius: 12, padding: 16, marginBottom: 12 }}>
                <div style={{ fontSize: 16, fontWeight: 600, marginBottom: 8 }}>{w.currency}</div>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, fontSize: 13 }}>
                  <div><span style={{ color: '#999' }}>Available: </span>{w.available}</div>
                  <div><span style={{ color: '#999' }}>In Operation: </span>{w.in_operation}</div>
                  <div><span style={{ color: '#999' }}>Frozen: </span>{w.frozen}</div>
                  <div><span style={{ color: '#999' }}>Earnings: </span>{w.earnings}</div>
                </div>
                <div style={{ marginTop: 8, fontSize: 15, fontWeight: 700 }}>Total: {w.total}</div>
              </div>
            ))}
          </div>
        </Tabs.TabPane>

        <Tabs.TabPane title="Transactions">
          <div style={{ padding: 16 }}>
            {user.recent_transactions.map((tx) => (
              <div key={tx.id} style={{ background: '#fff', borderRadius: 12, padding: 12, marginBottom: 8 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ fontWeight: 600 }}>{tx.type.toUpperCase()}</span>
                  <span style={{ color: tx.type === 'deposit' ? '#27ae60' : '#e74c3c', fontWeight: 600 }}>{tx.amount}</span>
                </div>
                <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>{tx.status} · {tx.created_at.slice(0, 10)}</div>
              </div>
            ))}
          </div>
        </Tabs.TabPane>

        <Tabs.TabPane title="Investments">
          <div style={{ padding: 16 }}>
            {user.investments.map((inv) => (
              <div key={inv.id} style={{ background: '#fff', borderRadius: 12, padding: 12, marginBottom: 8 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ fontWeight: 600 }}>{inv.amount} {inv.currency}</span>
                  <span style={{ fontSize: 13 }}>{inv.status}</span>
                </div>
                <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>Return: {inv.net_return} · Since {inv.start_date}</div>
              </div>
            ))}
          </div>
        </Tabs.TabPane>

        <Tabs.TabPane title="Settlements">
          <div style={{ padding: 16 }}>
            {user.recent_settlements.map((s) => (
              <div key={s.id} style={{ background: '#fff', borderRadius: 12, padding: 12, marginBottom: 8 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span style={{ fontWeight: 600 }}>{s.period}</span>
                  <span style={{ color: '#27ae60', fontWeight: 600 }}>{s.net_return}</span>
                </div>
                <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>Fee rate: {s.fee_rate}</div>
              </div>
            ))}
          </div>
        </Tabs.TabPane>
      </Tabs>

      {/* Adjust Balance Popup */}
      <Popup visible={showAdjust} close={() => setShowAdjust(false)} direction="bottom">
        <div style={{ padding: 24 }}>
          <h3 style={{ marginBottom: 16 }}>Adjust Balance (USDT)</h3>
          <div style={{ marginBottom: 12 }}>
            <Input placeholder="Amount (e.g. 100 or -50)" value={adjustAmount} onChange={(e) => setAdjustAmount(e.target.value)} />
          </div>
          <div style={{ marginBottom: 16 }}>
            <Input placeholder="Reason" value={adjustReason} onChange={(e) => setAdjustReason(e.target.value)} />
          </div>
          <Button onClick={handleAdjust} loading={adjustBalance.isPending} style={{ width: '100%' }}>Confirm</Button>
        </div>
      </Popup>

      {/* Edit User Popup */}
      <Popup visible={showEdit} close={() => setShowEdit(false)} direction="bottom">
        <div style={{ padding: 24 }}>
          <h3 style={{ marginBottom: 16 }}>Edit User</h3>
          <div style={{ marginBottom: 12 }}>
            <Input placeholder="Full Name" value={editName} onChange={(e) => setEditName(e.target.value)} />
          </div>
          <div style={{ marginBottom: 16 }}>
            <Input placeholder="Phone" value={editPhone} onChange={(e) => setEditPhone(e.target.value)} />
          </div>
          <Button onClick={handleEdit} loading={updateUser.isPending} style={{ width: '100%' }}>Save</Button>
        </div>
      </Popup>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
cd /Users/johnny/Work/soveregin && git add admin/src/app/\(admin\)/users/\[id\]/ && git commit -m "feat(admin-ui): add user detail page with actions and tabs"
```

---

### Task 20: Admin Management Page

**Files:**
- Create: `admin/src/app/(admin)/admin-users/page.tsx`

- [ ] **Step 1: Create admin management page**

Create `admin/src/app/(admin)/admin-users/page.tsx`:

```tsx
'use client'

import { useState } from 'react'
import { NavBar, Button, Dialog, Toast, Popup, Input, Picker, SwipeAction, Tag } from '@arco-design/mobile-react'
import { useAdminList, useCreateAdmin, useUpdateAdmin, useDeleteAdmin } from '@/hooks/use-api'
import type { AdminUser } from '@/types/api'

const roleOptions = [
  [
    { label: 'Super Admin', value: 'super_admin' },
    { label: 'Operator', value: 'operator' },
    { label: 'Viewer', value: 'viewer' },
  ],
]

const roleColors: Record<string, string> = {
  super_admin: '#e74c3c',
  operator: '#3498db',
  viewer: '#95a5a6',
}

export default function AdminUsersPage() {
  const { data: admins, isLoading } = useAdminList()
  const createAdmin = useCreateAdmin()
  const updateAdmin = useUpdateAdmin()
  const deleteAdmin = useDeleteAdmin()

  const [showCreate, setShowCreate] = useState(false)
  const [showEdit, setShowEdit] = useState(false)
  const [editTarget, setEditTarget] = useState<AdminUser | null>(null)
  const [form, setForm] = useState({ email: '', password: '', name: '', role: 'viewer' })
  const [showRolePicker, setShowRolePicker] = useState(false)

  const handleCreate = async () => {
    try {
      await createAdmin.mutateAsync(form)
      Toast.toast('Admin created')
      setShowCreate(false)
      setForm({ email: '', password: '', name: '', role: 'viewer' })
    } catch {
      Toast.toast('Failed to create admin')
    }
  }

  const handleEdit = async () => {
    if (!editTarget) return
    await updateAdmin.mutateAsync({ id: editTarget.id, data: { name: form.name, role: form.role } })
    Toast.toast('Admin updated')
    setShowEdit(false)
  }

  const handleDelete = (admin: AdminUser) => {
    Dialog.confirm({
      title: 'Delete Admin',
      content: `Delete ${admin.email}?`,
      onOk: async () => {
        await deleteAdmin.mutateAsync(admin.id)
        Toast.toast('Admin deleted')
      },
    })
  }

  const openEdit = (admin: AdminUser) => {
    setEditTarget(admin)
    setForm({ ...form, name: admin.name, role: admin.role })
    setShowEdit(true)
  }

  return (
    <div>
      <NavBar
        title="Admin Management"
        leftContent={null}
        rightContent={<span onClick={() => setShowCreate(true)} style={{ fontSize: 24 }}>+</span>}
      />
      <div style={{ padding: 16 }}>
        {(admins || []).map((admin) => (
          <SwipeAction
            key={admin.id}
            rightActions={[
              { text: 'Edit', color: '#3498db', onClick: () => openEdit(admin) },
              { text: 'Delete', color: '#e74c3c', onClick: () => handleDelete(admin) },
            ]}
          >
            <div style={{ background: '#fff', borderRadius: 12, padding: 16, marginBottom: 12 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <div style={{ fontSize: 16, fontWeight: 600 }}>{admin.name}</div>
                  <div style={{ fontSize: 13, color: '#999', marginTop: 4 }}>{admin.email}</div>
                </div>
                <Tag style={{ background: roleColors[admin.role], color: '#fff', fontSize: 12 }}>
                  {admin.role}
                </Tag>
              </div>
              {admin.last_login && (
                <div style={{ fontSize: 12, color: '#ccc', marginTop: 8 }}>
                  Last login: {new Date(admin.last_login).toLocaleString()}
                </div>
              )}
            </div>
          </SwipeAction>
        ))}
      </div>

      {/* Create Admin Popup */}
      <Popup visible={showCreate} close={() => setShowCreate(false)} direction="bottom">
        <div style={{ padding: 24 }}>
          <h3 style={{ marginBottom: 16 }}>Create Admin</h3>
          <div style={{ marginBottom: 12 }}><Input placeholder="Email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} /></div>
          <div style={{ marginBottom: 12 }}><Input placeholder="Password" type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} /></div>
          <div style={{ marginBottom: 12 }}><Input placeholder="Name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} /></div>
          <div style={{ marginBottom: 16 }}>
            <Button inline size="small" onClick={() => setShowRolePicker(true)}>Role: {form.role}</Button>
          </div>
          <Button onClick={handleCreate} loading={createAdmin.isPending} style={{ width: '100%' }}>Create</Button>
        </div>
      </Popup>

      {/* Edit Admin Popup */}
      <Popup visible={showEdit} close={() => setShowEdit(false)} direction="bottom">
        <div style={{ padding: 24 }}>
          <h3 style={{ marginBottom: 16 }}>Edit Admin</h3>
          <div style={{ marginBottom: 12 }}><Input placeholder="Name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} /></div>
          <div style={{ marginBottom: 16 }}>
            <Button inline size="small" onClick={() => setShowRolePicker(true)}>Role: {form.role}</Button>
          </div>
          <Button onClick={handleEdit} loading={updateAdmin.isPending} style={{ width: '100%' }}>Save</Button>
        </div>
      </Popup>

      {/* Role Picker */}
      <Picker
        visible={showRolePicker}
        data={roleOptions}
        onClose={() => setShowRolePicker(false)}
        onOk={(value) => { setForm({ ...form, role: value[0] as string }); setShowRolePicker(false) }}
      />
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
cd /Users/johnny/Work/soveregin && git add admin/src/app/\(admin\)/admin-users/ && git commit -m "feat(admin-ui): add admin management page with swipe actions"
```

---

### Task 21: Profile Page

**Files:**
- Create: `admin/src/app/(admin)/profile/page.tsx`

- [ ] **Step 1: Create profile page**

Create `admin/src/app/(admin)/profile/page.tsx`:

```tsx
'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { NavBar, Cell, Button, Dialog, Toast, Popup, Input } from '@arco-design/mobile-react'
import { useAuthStore } from '@/lib/auth'
import { useChangePassword } from '@/hooks/use-api'

export default function ProfilePage() {
  const router = useRouter()
  const admin = useAuthStore((s) => s.admin)
  const logout = useAuthStore((s) => s.logout)
  const changePassword = useChangePassword()

  const [showChangePwd, setShowChangePwd] = useState(false)
  const [oldPwd, setOldPwd] = useState('')
  const [newPwd, setNewPwd] = useState('')

  const handleLogout = () => {
    Dialog.confirm({
      title: 'Logout',
      content: 'Are you sure?',
      onOk: () => {
        logout()
        router.replace('/login')
      },
    })
  }

  const handleChangePwd = async () => {
    try {
      await changePassword.mutateAsync({ old_password: oldPwd, new_password: newPwd })
      Toast.toast('Password changed')
      setShowChangePwd(false)
      setOldPwd('')
      setNewPwd('')
    } catch {
      Toast.toast('Failed to change password')
    }
  }

  return (
    <div>
      <NavBar title="Profile" leftContent={null} />
      <div style={{ padding: 16 }}>
        <div style={{ background: '#fff', borderRadius: 12, padding: 20, marginBottom: 16, textAlign: 'center' }}>
          <div style={{ fontSize: 40, marginBottom: 8 }}>👤</div>
          <div style={{ fontSize: 18, fontWeight: 700 }}>{admin?.name}</div>
          <div style={{ fontSize: 14, color: '#999', marginTop: 4 }}>{admin?.email}</div>
          <div style={{ fontSize: 13, color: '#3498db', marginTop: 4 }}>{admin?.role}</div>
        </div>

        <div style={{ background: '#fff', borderRadius: 12, overflow: 'hidden', marginBottom: 16 }}>
          <Cell label="Change Password" showArrow onClick={() => setShowChangePwd(true)} />
        </div>

        <Button onClick={handleLogout} style={{ width: '100%', color: '#e74c3c' }}>Logout</Button>
      </div>

      <Popup visible={showChangePwd} close={() => setShowChangePwd(false)} direction="bottom">
        <div style={{ padding: 24 }}>
          <h3 style={{ marginBottom: 16 }}>Change Password</h3>
          <div style={{ marginBottom: 12 }}><Input placeholder="Old Password" type="password" value={oldPwd} onChange={(e) => setOldPwd(e.target.value)} /></div>
          <div style={{ marginBottom: 16 }}><Input placeholder="New Password" type="password" value={newPwd} onChange={(e) => setNewPwd(e.target.value)} /></div>
          <Button onClick={handleChangePwd} loading={changePassword.isPending} style={{ width: '100%' }}>Save</Button>
        </div>
      </Popup>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
cd /Users/johnny/Work/soveregin && git add admin/src/app/\(admin\)/profile/ && git commit -m "feat(admin-ui): add profile page with password change"
```

---

### Task 22: Deployment — Dockerfile + Docker Compose + Nginx

**Files:**
- Create: `deployments/Dockerfile.admin`
- Modify: `deployments/docker-compose.yml`
- Modify: `deployments/nginx/nginx.conf`

- [ ] **Step 1: Create Dockerfile.admin**

Create `deployments/Dockerfile.admin`:

```dockerfile
FROM node:20-alpine AS builder

WORKDIR /app

RUN corepack enable && corepack prepare pnpm@latest --activate

COPY admin/package.json admin/pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile || pnpm install

COPY admin/ .
ARG NEXT_PUBLIC_API_URL=/api/v1/admin
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
RUN pnpm build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

EXPOSE 3001
ENV PORT=3001
CMD ["node", "server.js"]
```

- [ ] **Step 2: Add admin service to docker-compose.yml**

Add after the `front` service in `deployments/docker-compose.yml`:

```yaml
  admin:
    build:
      context: ..
      dockerfile: deployments/Dockerfile.admin
      args:
        - NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_ADMIN_API_URL:-/api/v1/admin}
    depends_on:
      - server
    restart: unless-stopped
```

Also add `admin` to the nginx `depends_on` list.

- [ ] **Step 3: Update nginx.conf to proxy /admin**

Add a location block for the admin app in `deployments/nginx/nginx.conf`. Read the existing file first and add the admin location block.

- [ ] **Step 4: Commit**

```bash
cd /Users/johnny/Work/soveregin && git add deployments/ && git commit -m "feat(admin): add Docker deployment config for admin panel"
```

---

### Task 23: Full Build Verification

- [ ] **Step 1: Backend build**

Run: `cd /Users/johnny/Work/soveregin/server && go build ./... && go vet ./...`
Expected: No errors

- [ ] **Step 2: Backend tests**

Run: `cd /Users/johnny/Work/soveregin/server && go test ./... -count=1`
Expected: All PASS

- [ ] **Step 3: Frontend install + type check**

Run: `cd /Users/johnny/Work/soveregin/admin && pnpm install && npx tsc --noEmit`
Expected: No type errors

- [ ] **Step 4: Fix any remaining issues and commit**

```bash
cd /Users/johnny/Work/soveregin && git add -A && git commit -m "fix: resolve build issues in admin panel"
```
