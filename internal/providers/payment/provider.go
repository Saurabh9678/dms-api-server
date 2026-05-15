package payment

import "context"

type Provider interface {
	CreatePayment(ctx context.Context, amount int64, currency string, reference string) (string, error)
}
