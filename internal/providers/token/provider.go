package token

type Provider interface {
	Issue(userID uint64) (*TokenPair, error)
	Rotate(userID uint64) (*TokenPair, error)
	HashRefreshToken(token string) string
	ParseAccessToken(token string) (uint64, error)
}
