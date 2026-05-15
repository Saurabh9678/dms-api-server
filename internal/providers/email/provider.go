package email

import "context"

type Provider interface {
	Send(ctx context.Context, to string, subject string, body string) error
}
