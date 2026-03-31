package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestOK(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	OK(c, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("success should be true")
	}
}

func TestCreated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Created(c, map[string]string{"id": "123"})

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestFail(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Fail(c, http.StatusBadRequest, "BAD_REQ", "invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Success {
		t.Error("success should be false")
	}
	if resp.Error == nil || resp.Error.Code != "BAD_REQ" {
		t.Error("error code should be BAD_REQ")
	}
}

func TestPaginated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Paginated(c, []string{"a", "b"}, Meta{Total: 10, Page: 1, PerPage: 20})

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Meta == nil {
		t.Fatal("meta should not be nil")
	}
	if resp.Meta.Total != 10 {
		t.Errorf("total = %d, want 10", resp.Meta.Total)
	}
}
