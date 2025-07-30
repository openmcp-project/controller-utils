package smartrequeue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	store := NewStore(time.Second, time.Minute, 2)
	entry := newEntry(store)
	ctx := NewContext(context.Background(), entry)

	// Test that we get the entry back using FromContext
	got := FromContext(ctx)
	assert.Equal(t, entry, got, "Expected entry to be the same as the one set in context")
}

func TestFromContext(t *testing.T) {
	store := NewStore(time.Second, time.Minute, 2)
	entry := newEntry(store)
	ctx := NewContext(context.Background(), entry)

	// Retrieve entry from context
	got := FromContext(ctx)
	assert.Equal(t, entry, got, "Expected entry to be the same as the one set in context")

	// Test empty context
	got = FromContext(context.Background())
	assert.Nil(t, got, "Expected nil when no entry is set in context")
}
