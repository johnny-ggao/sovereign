package service

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"
)

type googleTokenVerifier struct {
	clientID string
}

func NewGoogleTokenVerifier(clientID string) GoogleTokenVerifier {
	return &googleTokenVerifier{clientID: clientID}
}

func (v *googleTokenVerifier) Verify(ctx context.Context, token string) (*GoogleClaims, error) {
	payload, err := idtoken.Validate(ctx, token, v.clientID)
	if err != nil {
		return nil, fmt.Errorf("validate id token: %w", err)
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)

	return &GoogleClaims{
		Sub:           payload.Subject,
		Email:         email,
		Name:          name,
		Picture:       picture,
		EmailVerified: emailVerified,
	}, nil
}
