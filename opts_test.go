package at

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test the field access for an opt.
func TestOpt(t *testing.T) {
	t.Parallel()

	opt := ServiceStates.Restricted
	assert.Equal(t, 1, opt.ID)
	assert.Equal(t, "Restricted service", opt.Description)
}

// Test the resolve function of an opt.
func TestResolve(t *testing.T) {
	t.Parallel()

	opt := ServiceStates.Restricted
	assert.Equal(t, opt, ServiceStates.Resolve(1))
}

// TODO: complete this suite in case of 100% coverage needed.
