package provider

import (
	"context"
	"log/slog"
)

// MockProvider is a test/dev implementation of EmailProvider
// that records all sent emails in the Sent slice.
type MockProvider struct {
	Sent []SendInput
}

// NewMockProvider creates a new MockProvider.
func NewMockProvider() EmailProvider {
	return &MockProvider{}
}

// Send records the email in the Sent slice and logs it.
func (m *MockProvider) Send(_ context.Context, input SendInput) error {
	m.Sent = append(m.Sent, input)
	slog.Info("mock email sent",
		slog.String("to", input.To),
		slog.String("subject", input.Subject),
	)
	return nil
}
