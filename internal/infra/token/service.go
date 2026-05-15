package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"infiour.local/dms-api-server/internal/domain/user"
)

type Config struct {
	AccessTokenSecret    string
	AccessTokenTTL       time.Duration
	RefreshTokenTTL      time.Duration
	RefreshTokenByteSize int
}

type Service struct {
	config Config
	nowFn  func() time.Time
}

func NewService(config Config) (*Service, error) {
	if config.AccessTokenSecret == "" {
		return nil, errors.New("access token secret is required")
	}
	if config.AccessTokenTTL <= 0 {
		config.AccessTokenTTL = 15 * time.Minute
	}
	if config.RefreshTokenTTL <= 0 {
		config.RefreshTokenTTL = 7 * 24 * time.Hour
	}
	if config.RefreshTokenByteSize <= 0 {
		config.RefreshTokenByteSize = 32
	}

	return &Service{
		config: config,
		nowFn:  time.Now,
	}, nil
}

func (s *Service) Issue(userID uint64) (*user.TokenPair, error) {
	return s.issueForUser(userID)
}

func (s *Service) Rotate(userID uint64) (*user.TokenPair, error) {
	return s.issueForUser(userID)
}

func (s *Service) HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *Service) issueForUser(userID uint64) (*user.TokenPair, error) {
	now := s.nowFn()
	accessExpiry := now.Add(s.config.AccessTokenTTL)

	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(userID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(accessExpiry),
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.config.AccessTokenSecret))
	if err != nil {
		return nil, err
	}

	refreshRaw := make([]byte, s.config.RefreshTokenByteSize)
	if _, err := rand.Read(refreshRaw); err != nil {
		return nil, err
	}
	refreshToken := base64.RawURLEncoding.EncodeToString(refreshRaw)
	refreshHash := s.HashRefreshToken(refreshToken)

	return &user.TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessTokenTTL:   int64(s.config.AccessTokenTTL.Seconds()),
		RefreshTokenTTL:  int64(s.config.RefreshTokenTTL.Seconds()),
		RefreshTokenHash: refreshHash,
	}, nil
}
