package otp

import (
	"context"
	"log"
)

type DummySender struct{}

func NewDummySender() *DummySender {
	return &DummySender{}
}

func (s *DummySender) Send(_ context.Context, destination string, code string) error {
	log.Printf("dummy otp sender destination=%s otp=%s", destination, code)
	return nil
}
