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
	"github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/service"
	apperr "github.com/sovereign-fund/sovereign/internal/shared/errors"
)

type fakeReviewSvc struct {
	approveErr error
	rejectErr  error
	retryErr   error
	lastReason string
	listItems  []dto.WithdrawReviewItem
	listTotal  int64
	listErr    error
}

func (f *fakeReviewSvc) List(ctx context.Context, q dto.WithdrawReviewListQuery) ([]dto.WithdrawReviewItem, int64, error) {
	return f.listItems, f.listTotal, f.listErr
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

// Compile-time interface checks
var _ service.WithdrawReviewService = (*fakeReviewSvc)(nil)
var _ service.AuditService = (*stubAudit)(nil)

func setup() (*gin.Engine, *fakeReviewSvc, *stubAudit) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := &fakeReviewSvc{listItems: []dto.WithdrawReviewItem{{ID: "tx1"}}, listTotal: 1}
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
	r, _, _ := setup()
	req := httptest.NewRequest("GET", "/admin/withdrawals?page=1&limit=20", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200, body = %s", rr.Code, rr.Body.String())
	}
}

func TestApprove_OnSuccess_LogsAudit(t *testing.T) {
	r, _, audit := setup()
	req := httptest.NewRequest("POST", "/admin/withdrawals/tx1/approve", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if audit.calls != 1 {
		t.Errorf("audit.calls = %d, want 1", audit.calls)
	}
}

func TestApprove_PropagatesAppError(t *testing.T) {
	r, svc, _ := setup()
	svc.approveErr = apperr.ErrWithdrawNotApprovable
	req := httptest.NewRequest("POST", "/admin/withdrawals/tx1/approve", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnprocessableEntity)
	}
}

func TestReject_RequiresReason(t *testing.T) {
	r, _, _ := setup()
	req := httptest.NewRequest("POST", "/admin/withdrawals/tx1/reject", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != 400 {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestReject_PassesReason(t *testing.T) {
	r, svc, _ := setup()
	body, _ := json.Marshal(dto.RejectWithdrawRequest{Reason: "address mismatch"})
	req := httptest.NewRequest("POST", "/admin/withdrawals/tx1/reject", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200, body = %s", rr.Code, rr.Body.String())
	}
	if svc.lastReason != "address mismatch" {
		t.Errorf("lastReason = %q, want 'address mismatch'", svc.lastReason)
	}
}

func TestRetry_PropagatesError(t *testing.T) {
	r, svc, _ := setup()
	svc.retryErr = errors.New("boom")
	req := httptest.NewRequest("POST", "/admin/withdrawals/tx1/retry", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != 500 {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}
