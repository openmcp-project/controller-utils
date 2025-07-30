package smartrequeue

import "context"

// contextKey is a type used as a key for storing and retrieving the Entry from the context.
type contextKey struct{}

// NewContext creates a new context with the given Entry.
// This is a utility function for passing Entry instances through context.
func NewContext(ctx context.Context, entry *Entry) context.Context {
	return context.WithValue(ctx, contextKey{}, entry)
}

// FromContext retrieves the Entry from the context, if it exists.
// Returns nil if no Entry is found in the context.
func FromContext(ctx context.Context) *Entry {
	entry, ok := ctx.Value(contextKey{}).(*Entry)
	if !ok {
		return nil
	}
	return entry
}
