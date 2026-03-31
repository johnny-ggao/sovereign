package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestParseDefaults(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)

	p := Parse(c)
	if p.Page != 1 {
		t.Errorf("Page = %d, want 1", p.Page)
	}
	if p.PerPage != 20 {
		t.Errorf("PerPage = %d, want 20", p.PerPage)
	}
}

func TestParseCustomValues(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test?page=3&per_page=50", nil)

	p := Parse(c)
	if p.Page != 3 {
		t.Errorf("Page = %d, want 3", p.Page)
	}
	if p.PerPage != 50 {
		t.Errorf("PerPage = %d, want 50", p.PerPage)
	}
}

func TestParseMaxPerPage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test?per_page=999", nil)

	p := Parse(c)
	if p.PerPage != 100 {
		t.Errorf("PerPage should be capped at 100, got %d", p.PerPage)
	}
}

func TestOffset(t *testing.T) {
	p := Params{Page: 3, PerPage: 20}
	if p.Offset() != 40 {
		t.Errorf("Offset = %d, want 40", p.Offset())
	}
}
