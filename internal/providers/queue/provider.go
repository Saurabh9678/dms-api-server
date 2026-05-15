package queue

import "context"

type Provider interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}
