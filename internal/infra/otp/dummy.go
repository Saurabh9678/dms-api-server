package otp

import (
	"context"
	"log/slog"

	otpprovider "infiour.local/dms-api-server/internal/providers/otp"
)

type DummyProvider struct {
	log *slog.Logger
}

var _ otpprovider.Provider = (*DummyProvider)(nil)

func NewDummyProvider(log *slog.Logger) *DummyProvider {
	return &DummyProvider{log: log}
}

func (p *DummyProvider) Send(ctx context.Context, destination string, code string) error {
	p.log.InfoContext(ctx, "dummy otp sender", "destination", destination, "otp", code)
	return nil
}
