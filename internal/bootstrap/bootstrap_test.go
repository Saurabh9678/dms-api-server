package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Compile-time check that exported names exist.
func TestExportsExist(t *testing.T) {
	assert.NotNil(t, NewRouter)
	assert.NotNil(t, BuildDependencies)
}
