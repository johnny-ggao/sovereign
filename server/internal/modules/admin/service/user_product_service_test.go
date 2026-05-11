package service

import (
	"context"
	"errors"
	"testing"

	logmodel "github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	usermodel "github.com/sovereign-fund/sovereign/internal/modules/auth/model"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
)

// === stubs ===

type stubUserRepoForProduct struct {
	users     map[string]*usermodel.User
	updateErr error
}

func (s *stubUserRepoForProduct) FindByID(ctx context.Context, id string) (*usermodel.User, error) {
	if u, ok := s.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}
func (s *stubUserRepoForProduct) Update(ctx context.Context, u *usermodel.User) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	if existing, ok := s.users[u.ID]; ok {
		existing.InvestmentType = u.InvestmentType
	}
	return nil
}
func (s *stubUserRepoForProduct) FindByEmail(context.Context, string) (*usermodel.User, error) {
	panic("unused")
}
func (s *stubUserRepoForProduct) FindByGoogleID(context.Context, string) (*usermodel.User, error) {
	panic("unused")
}
func (s *stubUserRepoForProduct) Create(context.Context, *usermodel.User) error {
	panic("unused")
}
func (s *stubUserRepoForProduct) ExistsByEmail(context.Context, string) (bool, error) {
	panic("unused")
}

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

// === tests ===

func TestChange_FromArbitrageToTrading(t *testing.T) {
	ctx := context.Background()
	uRepo := &stubUserRepoForProduct{users: map[string]*usermodel.User{
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
	if l.FromType != "arbitrage" || l.ToType != "trading" || l.Reason != "internal pilot" || l.AdminID != "admin-1" || l.AdminEmail != "admin@x.com" {
		t.Errorf("log = %+v, mismatch", l)
	}
}

func TestChange_RejectsInvalidType(t *testing.T) {
	uRepo := &stubUserRepoForProduct{users: map[string]*usermodel.User{"u1": {ID: "u1", InvestmentType: "arbitrage"}}}
	svc := NewUserProductService(uRepo, &stubLogRepo{})
	err := svc.Change(context.Background(), "u1", "garbage", "x", "a", "a@x.com")
	var ae *apperr.AppError
	if !errors.As(err, &ae) || ae.Code != "INVALID_PRODUCT_TYPE" {
		t.Errorf("error code = %v, want INVALID_PRODUCT_TYPE", ae)
	}
}

func TestChange_RejectsSameType(t *testing.T) {
	uRepo := &stubUserRepoForProduct{users: map[string]*usermodel.User{"u1": {ID: "u1", InvestmentType: "arbitrage"}}}
	svc := NewUserProductService(uRepo, &stubLogRepo{})
	err := svc.Change(context.Background(), "u1", "arbitrage", "x", "a", "a@x.com")
	var ae *apperr.AppError
	if !errors.As(err, &ae) || ae.Code != "SAME_PRODUCT_TYPE" {
		t.Errorf("error code = %v, want SAME_PRODUCT_TYPE", ae)
	}
}

func TestChange_TreatsEmptyAsArbitrage(t *testing.T) {
	uRepo := &stubUserRepoForProduct{users: map[string]*usermodel.User{"u1": {ID: "u1", InvestmentType: ""}}}
	lRepo := &stubLogRepo{}
	svc := NewUserProductService(uRepo, lRepo)
	if err := svc.Change(context.Background(), "u1", "trading", "test", "admin-1", "admin@x.com"); err != nil {
		t.Fatalf("Change error = %v", err)
	}
	if lRepo.created[0].FromType != "arbitrage" {
		t.Errorf("FromType = %s, want arbitrage (empty -> arbitrage)", lRepo.created[0].FromType)
	}
}

func TestBulkChange_PartialFailures(t *testing.T) {
	uRepo := &stubUserRepoForProduct{users: map[string]*usermodel.User{
		"u1": {ID: "u1", InvestmentType: "arbitrage"},
		"u2": {ID: "u2", InvestmentType: "trading"}, // already trading - same-type fail
	}}
	svc := NewUserProductService(uRepo, &stubLogRepo{})
	res, err := svc.BulkChange(context.Background(), []string{"u1", "u2", "u3"}, "trading", "test reason ok", "a", "a@x.com")
	if err != nil {
		t.Fatalf("BulkChange should not return top-level error: %v", err)
	}
	if len(res.Succeeded) != 1 || res.Succeeded[0] != "u1" {
		t.Errorf("succeeded = %v, want [u1]", res.Succeeded)
	}
	if len(res.Failed) != 2 {
		t.Errorf("failed = %v, want 2 entries (u2 same-type, u3 not found)", res.Failed)
	}
}
