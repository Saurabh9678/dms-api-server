package user

import "context"

type OTPSender interface {
	Send(ctx context.Context, destination string, code string) error
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessTokenTTL   int64
	RefreshTokenTTL  int64
	RefreshTokenHash string
}

type TokenService interface {
	Issue(userID uint64) (*TokenPair, error)
	Rotate(userID uint64) (*TokenPair, error)
	HashRefreshToken(token string) string
}
