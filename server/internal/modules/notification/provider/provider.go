package provider

import "context"

// EmailProvider abstracts email sending.
// Implementations: SESProvider (production), MockProvider (dev/test).
type EmailProvider interface {
	Send(ctx context.Context, input SendInput) error
}

// SendInput contains the parameters for sending an email.
type SendInput struct {
	To      string
	Subject string
	HTML    string
}
