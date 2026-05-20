package vehicle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRepository(t *testing.T) {
	repo := NewRepository(nil)
	assert.NotNil(t, repo)
}
