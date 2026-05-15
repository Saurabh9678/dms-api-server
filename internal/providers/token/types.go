package token

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessTokenTTL   int64
	RefreshTokenTTL  int64
	RefreshTokenHash string
}
