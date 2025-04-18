/*
Copyright 2025 Keikoproj authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

	t.Run("missing display expression error has correct format", func(t *testing.T) {
		input := &wavefront.Alert{
			Name:                "test-alert",
			Minutes:             5,
			ResolveAfterMinutes: 5,
		}
		err := wf.ValidateAlertInput(ctx, input)
		assert.Error(t, err)
		assert.Equal(t, "validation failed: displayExpression must not be empty", err.Error())
	})

	t.Run("minutes value error has correct format", func(t *testing.T) {
		input := &wavefront.Alert{
			Name:                "test-alert",
			DisplayExpression:   "ts(status.health)",
			Minutes:             0, // Invalid value
			ResolveAfterMinutes: 5,
		}
		err := wf.ValidateAlertInput(ctx, input)
		assert.Error(t, err)
		assert.Equal(t, "validation failed: minutes must be greater than 0", err.Error())
	})

	t.Run("resolve after minutes value error has correct format", func(t *testing.T) {
		input := &wavefront.Alert{
			Name:                "test-alert",
			DisplayExpression:   "ts(status.health)",
			Minutes:             5,
			ResolveAfterMinutes: 0, // Invalid value
		}
		err := wf.ValidateAlertInput(ctx, input)
		assert.Error(t, err)
		assert.Equal(t, "validation failed: resolveAfterMinutes must be greater than 0", err.Error())
	})

	t.Run("missing conditions error has correct format", func(t *testing.T) {
		input := &wavefront.Alert{
			Name:                "test-alert",
			AlertType:           wavefront.AlertTypeThreshold,
			DisplayExpression:   "ts(status.health)",
			Minutes:             5,
			ResolveAfterMinutes: 5,
		}
		err := wf.ValidateAlertInput(ctx, input)
		assert.Error(t, err)
		assert.Equal(t, "validation failed: conditions must not be empty", err.Error())
	})

	t.Run("missing condition error has correct format", func(t *testing.T) {
		input := &wavefront.Alert{
			Name:                "test-alert",
			AlertType:           wavefront.AlertTypeClassic,
			DisplayExpression:   "ts(status.health)",
			Minutes:             5,
			ResolveAfterMinutes: 5,
		}
		err := wf.ValidateAlertInput(ctx, input)
		assert.Error(t, err)
		assert.Equal(t, "validation failed: condition must not be empty", err.Error())
	})

	t.Run("missing severity error has correct format", func(t *testing.T) {
		input := &wavefront.Alert{
			Name:                "test-alert",
			AlertType:           wavefront.AlertTypeClassic,
			Condition:           "ts(status.health)",
			DisplayExpression:   "ts(status.health)",
			Minutes:             5,
			ResolveAfterMinutes: 5,
		}
		err := wf.ValidateAlertInput(ctx, input)
		assert.Error(t, err)
		assert.Equal(t, "validation failed: severity must not be empty", err.Error())
	})

	t.Run("invalid alert type error has correct format", func(t *testing.T) {
		input := &wavefront.Alert{
			Name:                "test-alert",
			AlertType:           "INVALID_TYPE",
			DisplayExpression:   "ts(status.health)",
			Minutes:             5,
			ResolveAfterMinutes: 5,
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
