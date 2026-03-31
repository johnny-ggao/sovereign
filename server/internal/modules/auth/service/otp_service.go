package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sovereign-fund/sovereign/config"
)

type OTPService interface {
	Generate(ctx context.Context, email, purpose string) (string, error)
	Verify(ctx context.Context, email, purpose, code string) (bool, error)
}

type otpService struct {
	redis    *redis.Client
	length   int
	expiry   time.Duration
	coolDown time.Duration
}

func NewOTPService(rdb *redis.Client, cfg config.OTPConfig) OTPService {
	return &otpService{
		redis:    rdb,
		length:   cfg.Length,
		expiry:   cfg.Expiry,
		coolDown: cfg.CoolDown,
	}
}

func (s *otpService) Generate(ctx context.Context, email, purpose string) (string, error) {
	cooldownKey := fmt.Sprintf("otp_cooldown:%s:%s", email, purpose)
	if s.redis.Exists(ctx, cooldownKey).Val() > 0 {
		return "", fmt.Errorf("please wait before requesting a new OTP")
	}

	code := generateNumericCode(s.length)

	otpKey := fmt.Sprintf("otp:%s:%s", email, purpose)
	if err := s.redis.Set(ctx, otpKey, code, s.expiry).Err(); err != nil {
		return "", fmt.Errorf("store OTP: %w", err)
	}

	if err := s.redis.Set(ctx, cooldownKey, "1", s.coolDown).Err(); err != nil {
		return "", fmt.Errorf("set cooldown: %w", err)
	}

	return code, nil
}

// verifyOTPScript atomically checks and deletes an OTP to prevent race conditions.
// Returns 1 if code matches (and key is deleted), 0 if not found, -1 if code mismatch.
var verifyOTPScript = redis.NewScript(`
local stored = redis.call("GET", KEYS[1])
if not stored then
	return 0
end
if stored ~= ARGV[1] then
	return -1
end
redis.call("DEL", KEYS[1])
return 1
`)

func (s *otpService) Verify(ctx context.Context, email, purpose, code string) (bool, error) {
	otpKey := fmt.Sprintf("otp:%s:%s", email, purpose)

	result, err := verifyOTPScript.Run(ctx, s.redis, []string{otpKey}, code).Int()
	if err != nil {
		return false, fmt.Errorf("verify OTP: %w", err)
	}

	return result == 1, nil
}

func generateNumericCode(length int) string {
	code := ""
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code
}
