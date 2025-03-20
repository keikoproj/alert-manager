package controllers_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/keikoproj/alert-manager/internal/controllers"
	"github.com/stretchr/testify/assert"
)

// TestContextKey ensures our custom context key implementation
// works correctly for request ID storage/retrieval
func TestContextKey(t *testing.T) {
	// Access the unexported contextKey and requestIDKey using reflection
	// This is for testing purposes only
	t.Run("context with custom key type works as expected", func(t *testing.T) {
		ctx := context.Background()

		// Create a UUID that we'll store in the context
		requestID := uuid.New()

		// Store the UUID in the context using the controllers package
		// which has access to the unexported contextKey type
		ctx = controllers.WithRequestID(ctx, requestID)

		// Retrieve the UUID from the context
		retrievedID, ok := controllers.GetRequestID(ctx)

		// Verify the UUID was stored and retrieved correctly
		assert.True(t, ok, "Expected to retrieve request ID from context")
		assert.Equal(t, requestID, retrievedID, "Expected retrieved ID to match original")
	})
}
