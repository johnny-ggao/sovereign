package template

import (
	"strings"
	"testing"
)

func TestRendererRenderSuccess(t *testing.T) {
	r, err := NewRenderer("emails")
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	data := map[string]string{
		"Amount":   "100.00",
		"Currency": "USDT",
		"Network":  "Ethereum",
		"TxHash":   "0xabc123",
	}

	subject, html, err := r.Render("deposit_confirmed", "en", data)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if subject == "" {
		t.Error("subject should not be empty")
	}
	if html == "" {
		t.Error("html should not be empty")
	}

	wantSubject := "Your deposit of 100.00 USDT has been confirmed"
	if subject != wantSubject {
		t.Errorf("subject = %q, want %q", subject, wantSubject)
	}

	if !strings.Contains(html, "100.00 USDT") {
		t.Error("html should contain amount and currency")
	}
	if !strings.Contains(html, "Ethereum") {
		t.Error("html should contain network")
	}
	if !strings.Contains(html, "0xabc123") {
		t.Error("html should contain txhash")
	}
}

func TestRendererFallbackToEnglish(t *testing.T) {
	r, err := NewRenderer("emails")
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	data := map[string]string{
		"Amount":   "50.00",
		"Currency": "USDT",
		"Network":  "Tron",
		"TxHash":   "0xdef456",
	}

	subject, _, err := r.Render("deposit_confirmed", "ja", data)
	if err != nil {
		t.Fatalf("Render with fallback: %v", err)
	}

	if subject == "" {
		t.Error("should fallback to English template")
	}
}

func TestRendererUnknownEvent(t *testing.T) {
	r, err := NewRenderer("emails")
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	_, _, err = r.Render("nonexistent_event", "en", nil)
	if err == nil {
		t.Error("expected error for unknown event type")
	}
}

func TestRendererInvalidDir(t *testing.T) {
	_, err := NewRenderer("nonexistent_dir")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}
