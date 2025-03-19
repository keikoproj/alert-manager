package k8s

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewK8sSelfClientDoOrDie(t *testing.T) {
	// Save original KUBECONFIG
	originalKubeConfig := os.Getenv("KUBECONFIG")
	defer os.Setenv("KUBECONFIG", originalKubeConfig)

	t.Run("handles local environment properly", func(t *testing.T) {
		// Set KUBECONFIG to a non-existent file to force local path
		// We'll catch the panic and verify it's the expected panic
		os.Setenv("KUBECONFIG", "/non/existent/path")

		// Expect a panic but recover from it
		defer func() {
			if r := recover(); r != nil {
				// This is expected - we set KUBECONFIG to a non-existent path
				// We just want to verify the code path is executed properly
				assert.Contains(t, r.(error).Error(), "stat /non/existent/path",
					"Expected panic due to non-existent KUBECONFIG path")
			} else {
				t.Fatal("Expected a panic due to non-existent KUBECONFIG but none occurred")
			}
		}()

		// This should trigger our fix for the ineffectual assignment
		// We expect it to panic because the KUBECONFIG is invalid
		// We just want to make sure it follows the correct path
		NewK8sSelfClientDoOrDie()
	})
}
