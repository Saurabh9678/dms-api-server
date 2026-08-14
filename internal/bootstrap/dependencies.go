package bootstrap

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	infraotp "infiour.local/dms-api-server/internal/infra/otp"
	infrastorage "infiour.local/dms-api-server/internal/infra/storage"
	infratoken "infiour.local/dms-api-server/internal/infra/token"
	"infiour.local/dms-api-server/internal/modules/auth"
	"infiour.local/dms-api-server/internal/modules/dashboard"
	"infiour.local/dms-api-server/internal/modules/showroom"
	"infiour.local/dms-api-server/internal/modules/user"
	"infiour.local/dms-api-server/internal/modules/vehicle"
	tokenprovider "infiour.local/dms-api-server/internal/providers/token"
	"infiour.local/dms-api-server/pkg/config"
	"infiour.local/dms-api-server/pkg/middleware"
)

type Dependencies struct {
	AuthHandler             *auth.Handler
	UserHandler             *user.Handler
	VehicleHandler          *vehicle.Handler
	DashboardHandler        *dashboard.Handler
	ShowroomHandler         *showroom.Handler
	TokenProvider           tokenprovider.Provider
	ShowroomRolesMiddleware gin.HandlerFunc
}

func buildDependencies(cfg *config.Config, db *gorm.DB, log *slog.Logger) *Dependencies {
	userRepo := user.NewRepository(db)

	otpProvider := infraotp.NewDummyProvider(log)
	tokenProvider := infratoken.NewJWTProvider(cfg.Auth)

	otpRepo := auth.NewOTPRepository(db)
	sessionRepo := auth.NewSessionRepository(db)
	authSvc := auth.NewService(userRepo, otpRepo, sessionRepo, otpProvider, tokenProvider, cfg.Auth, db, cfg.Env)
	authHandler := auth.NewHandler(authSvc)

	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc)

	dashboardRepo := dashboard.NewRepository(db)
	dashboardSvc := dashboard.NewService(dashboardRepo)
	dashboardHandler := dashboard.NewHandler(dashboardSvc)

	storageProvider, err := infrastorage.NewProvider(cfg.Storage)
	if err != nil {
		log.Error("failed to init storage provider", "error", err)
		panic(err)
	}

	vehicleRepo := vehicle.NewRepository(db)
	vehicleSvc := vehicle.NewService(vehicleRepo, storageProvider, vehicle.WithSignedURLTTL(cfg.Storage.SignedURLTTL))
	vehicleHandler := vehicle.NewHandler(vehicleSvc)

	showroomRepo := showroom.NewRepository(db)
	showroomSvc := showroom.NewService(showroomRepo, storageProvider, showroom.WithSignedURLTTL(cfg.Storage.SignedURLTTL))
	showroomHandler := showroom.NewHandler(showroomSvc)

	return &Dependencies{
		AuthHandler:             authHandler,
		UserHandler:             userHandler,
		VehicleHandler:          vehicleHandler,
		DashboardHandler:        dashboardHandler,
		ShowroomHandler:         showroomHandler,
		TokenProvider:           tokenProvider,
		ShowroomRolesMiddleware: middleware.RequireShowroomRoles(userRepo),
	}
}
