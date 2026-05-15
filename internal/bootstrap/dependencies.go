package bootstrap

import (
	"log/slog"

	"gorm.io/gorm"
	infraotp "infiour.local/dms-api-server/internal/infra/otp"
	infratoken "infiour.local/dms-api-server/internal/infra/token"
	"infiour.local/dms-api-server/internal/modules/auth"
	"infiour.local/dms-api-server/internal/modules/user"
	"infiour.local/dms-api-server/pkg/config"
)

type Dependencies struct {
	AuthHandler *auth.Handler
}

func buildDependencies(cfg *config.Config, db *gorm.DB, log *slog.Logger) *Dependencies {
	userRepo := user.NewRepository(db)

	otpProvider := infraotp.NewDummyProvider(log)
	tokenProvider := infratoken.NewJWTProvider(cfg.Auth)

	otpRepo := auth.NewOTPRepository(db)
	sessionRepo := auth.NewSessionRepository(db)
	authSvc := auth.NewService(userRepo, otpRepo, sessionRepo, otpProvider, tokenProvider, cfg.Auth, db)
	authHandler := auth.NewHandler(authSvc)

	return &Dependencies{
		AuthHandler: authHandler,
	}
}
