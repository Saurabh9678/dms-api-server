package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustLoad_PanicOnLoadError(t *testing.T) {
	orig := loaderFn
	defer func() { loaderFn = orig }()

	loaderFn = func() (*Config, error) {
		return nil, errors.New("simulated load error")
	}

	assert.Panics(t, func() { MustLoad() })
}
