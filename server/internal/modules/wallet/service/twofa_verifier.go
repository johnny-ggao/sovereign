package service

import (
	"context"

	authRepo "github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	"github.com/sovereign-fund/sovereign/pkg/crypto"
)

type totpVerifier struct {
	userRepo authRepo.UserRepository
}

func NewTOTPVerifier(userRepo authRepo.UserRepository) TwoFAVerifier {
	return &totpVerifier{userRepo: userRepo}
}

func (v *totpVerifier) Verify2FA(ctx context.Context, userID, code string) (bool, error) {
	user, err := v.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}

	// 用户未启用 2FA 时直接通过
	if !user.TwoFAEnabled || user.TwoFASecret == "" {
		return true, nil
	}

	if code == "" {
		return false, nil
	}

	return crypto.VerifyTOTP(user.TwoFASecret, code)
}
