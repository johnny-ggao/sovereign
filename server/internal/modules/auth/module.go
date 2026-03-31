package auth

import (
	"log/slog"

	"github.com/redis/go-redis/v9"
	"github.com/sovereign-fund/sovereign/config"
	"github.com/sovereign-fund/sovereign/internal/modules/auth/handler"
	"github.com/sovereign-fund/sovereign/internal/modules/auth/repository"
	"github.com/sovereign-fund/sovereign/internal/modules/auth/service"
	"github.com/sovereign-fund/sovereign/internal/shared/events"
	jwtpkg "github.com/sovereign-fund/sovereign/pkg/jwt"
	"gorm.io/gorm"
)

type Module struct {
	Handler *handler.AuthHandler
	Service service.AuthService
}

func NewModule(db *gorm.DB, rdb *redis.Client, jwtMgr *jwtpkg.Manager, bus events.Bus, cfg *config.Config, logger *slog.Logger) *Module {
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	otpSvc := service.NewOTPService(rdb, cfg.OTP)
	googleVerifier := service.NewGoogleTokenVerifier(cfg.Google.ClientID)
	authSvc := service.NewAuthService(userRepo, tokenRepo, jwtMgr, otpSvc, googleVerifier, bus, logger)
	authHandler := handler.NewAuthHandler(authSvc)

	return &Module{
		Handler: authHandler,
		Service: authSvc,
	}
}
