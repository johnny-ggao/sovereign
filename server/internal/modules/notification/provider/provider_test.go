package provider

import (
	"context"
	"testing"
)

func TestMockProviderSend(t *testing.T) {
	mock := NewMockProvider()
	input := SendInput{
		To:      "user@example.com",
		Subject: "Test Subject",
		HTML:    "<p>Hello</p>",
	}

	err := mock.Send(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sent := mock.(*MockProvider).Sent
	if len(sent) != 1 {
		t.Fatalf("expected 1 sent email, got %d", len(sent))
	}
	if sent[0].To != "user@example.com" {
		t.Errorf("to = %q, want %q", sent[0].To, "user@example.com")
	}
	if sent[0].Subject != "Test Subject" {
		t.Errorf("subject = %q, want %q", sent[0].Subject, "Test Subject")
	}
}

func TestMockProviderMultipleSends(t *testing.T) {
	mock := NewMockProvider()

	for i := 0; i < 3; i++ {
		err := mock.Send(context.Background(), SendInput{
			To:      "user@example.com",
			Subject: "Test",
			HTML:    "<p>Hello</p>",
		})
		if err != nil {
			t.Fatalf("send %d: unexpected error: %v", i, err)
		}
	}

	sent := mock.(*MockProvider).Sent
	if len(sent) != 3 {
		t.Fatalf("expected 3 sent emails, got %d", len(sent))
	}
}
