package wavefront_test

import (
	"context"
	"testing"

	"github.com/WavefrontHQ/go-wavefront-management-api"
	wf "github.com/keikoproj/alert-manager/pkg/wavefront"
	"github.com/stretchr/testify/assert"
)

// TestErrorMessageFormatting specifically tests our improvements to error formatting
// where we simplified the error message construction and avoided unnecessary fmt.Sprintf
func TestErrorMessageFormatting(t *testing.T) {
	ctx := context.Background()

	t.Run("missing conditions error has correct format", func(t *testing.T) {
		input := &wavefront.Alert{
			Name:      "test-alert",
			AlertType: wavefront.AlertTypeThreshold,
		}
		err := wf.ValidateAlertInput(ctx, input)
		assert.Error(t, err)
		assert.Equal(t, "validation failed: conditions must not be empty", err.Error())
	})

	t.Run("missing condition error has correct format", func(t *testing.T) {
		input := &wavefront.Alert{
			Name:      "test-alert",
			AlertType: wavefront.AlertTypeClassic,
		}
		err := wf.ValidateAlertInput(ctx, input)
		assert.Error(t, err)
		assert.Equal(t, "validation failed: condition must not be empty", err.Error())
	})

	t.Run("missing severity error has correct format", func(t *testing.T) {
		input := &wavefront.Alert{
			Name:      "test-alert",
			AlertType: wavefront.AlertTypeClassic,
			Condition: "ts(status.health)",
		}
		err := wf.ValidateAlertInput(ctx, input)
		assert.Error(t, err)
		assert.Equal(t, "validation failed: severity must not be empty", err.Error())
	})

	t.Run("invalid alert type error has correct format", func(t *testing.T) {
		input := &wavefront.Alert{
			Name:      "test-alert",
			AlertType: "INVALID_TYPE",
		}
		err := wf.ValidateAlertInput(ctx, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed: invalid alert type: INVALID_TYPE")
	})

	t.Run("missing exported param has correct format", func(t *testing.T) {
		params := []string{"required_param"}
		configValues := map[string]string{
			"other_param": "value",
		}
		err := wf.ValidateTemplateParams(ctx, params, configValues)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Required exported param required_param is not supplied")
	})
}
