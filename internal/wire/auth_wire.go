package wire

import (
	"fmt"

	"gorm.io/gorm"
	"infiour.local/dms-api-server/internal/api/http/handler"
	"infiour.local/dms-api-server/internal/application/auth"
	"infiour.local/dms-api-server/internal/domain/user"
	"infiour.local/dms-api-server/internal/infra/config"
	otpinfra "infiour.local/dms-api-server/internal/infra/sms/otp"
	"infiour.local/dms-api-server/internal/infra/token"
	userotp "infiour.local/dms-api-server/internal/repository/user_otp"
	usersession "infiour.local/dms-api-server/internal/repository/user_session"
	"infiour.local/dms-api-server/internal/repository/users"
)

func BuildAuthHandler(database *gorm.DB) (*handler.AuthHandler, error) {
	authConfig := config.LoadAuthConfig()

	tokenService, err := token.NewService(token.Config{
		AccessTokenSecret: authConfig.AccessTokenSecret,
		AccessTokenTTL:    authConfig.AccessTokenTTL,
		RefreshTokenTTL:   authConfig.RefreshTokenTTL,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize token service: %w", err)
	}

	userRepo := users.NewRepository(database)
	otpRepo := userotp.NewRepository(database)
	sessionRepo := usersession.NewRepository(database)
	otpSender := otpinfra.NewDummySender()
	authService := auth.NewService(
		userRepo,
		otpRepo,
		sessionRepo,
		otpSender,
		tokenService,
		auth.Config{
			OTPTTL:         authConfig.OTPTTL,
			OTPMaxAttempts: authConfig.OTPMaxAttempts,
			OTPFor:         user.OTPForMobile,
		},
	)

	return handler.NewAuthHandler(authService), nil
}
