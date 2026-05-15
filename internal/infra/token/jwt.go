package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	tokenprovider "infiour.local/dms-api-server/internal/providers/token"
	"infiour.local/dms-api-server/pkg/config"
)

type JWTProvider struct {
	config config.AuthConfig
	nowFn  func() time.Time
}

var _ tokenprovider.Provider = (*JWTProvider)(nil)

func NewJWTProvider(cfg config.AuthConfig) *JWTProvider {
	if cfg.AccessTokenTTL <= 0 {
		cfg.AccessTokenTTL = 15 * time.Minute
	}
	if cfg.RefreshTokenTTL <= 0 {
		cfg.RefreshTokenTTL = 7 * 24 * time.Hour
	}
	return &JWTProvider{
		config: cfg,
		nowFn:  time.Now,
	}
}

func (p *JWTProvider) Issue(userID uint64) (*tokenprovider.TokenPair, error) {
	return p.issueForUser(userID)
}

func (p *JWTProvider) Rotate(userID uint64) (*tokenprovider.TokenPair, error) {
	return p.issueForUser(userID)
}

func (p *JWTProvider) HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (p *JWTProvider) ParseAccessToken(token string) (uint64, error) {
	if p.config.AccessTokenSecret == "" {
		return 0, errors.New("access token secret is required")
	}

	parsedToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method == nil || !strings.EqualFold(t.Method.Alg(), jwt.SigningMethodHS256.Alg()) {
			return nil, errors.New("invalid signing method")
		}
		return []byte(p.config.AccessTokenSecret), nil
	})
	if err != nil {
		return 0, err
	}
	if !parsedToken.Valid {
		return 0, errors.New("invalid access token")
	}

	claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return 0, errors.New("invalid access token claims")
	}
	userID, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (p *JWTProvider) issueForUser(userID uint64) (*tokenprovider.TokenPair, error) {
	if p.config.AccessTokenSecret == "" {
		return nil, errors.New("access token secret is required")
	}

	now := p.nowFn()
	accessExpiry := now.Add(p.config.AccessTokenTTL)

	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(userID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(accessExpiry),
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(p.config.AccessTokenSecret))
	if err != nil {
		return nil, err
	}

	refreshRaw := make([]byte, 32)
	if _, err := rand.Read(refreshRaw); err != nil {
		return nil, err
	}
	refreshToken := base64.RawURLEncoding.EncodeToString(refreshRaw)
	refreshHash := p.HashRefreshToken(refreshToken)

	return &tokenprovider.TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessTokenTTL:   int64(p.config.AccessTokenTTL.Seconds()),
		RefreshTokenTTL:  int64(p.config.RefreshTokenTTL.Seconds()),
		RefreshTokenHash: refreshHash,
	}, nil
}
