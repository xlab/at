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

// Test DeleteOptions field access and resolve function.
func TestDeleteOptions(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, DeleteOptions.Index.ID)
	assert.Equal(t, "Delete message by index", DeleteOptions.Index.Description)
	assert.Equal(t, 4, DeleteOptions.All.ID)
	assert.Equal(t, "Delete all messages", DeleteOptions.All.Description)
	assert.Equal(t, DeleteOptions.Index, DeleteOptions.Resolve(0))
	assert.Equal(t, DeleteOptions.All, DeleteOptions.Resolve(4))
}

// Test MessageFlags field access and resolve function.
func TestMessageFlags(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, MessageFlags.Unread.ID)
	assert.Equal(t, "Unread", MessageFlags.Unread.Description)
	assert.Equal(t, 4, MessageFlags.Any.ID)
	assert.Equal(t, "Any", MessageFlags.Any.Description)
	assert.Equal(t, MessageFlags.Unread, MessageFlags.Resolve(0))
	assert.Equal(t, MessageFlags.Any, MessageFlags.Resolve(4))
}

// TODO: complete this suite in case of 100% coverage needed.
