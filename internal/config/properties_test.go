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

package config

import (
	"os"
	"testing"

	"github.com/keikoproj/alert-manager/internal/config/common"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLoadProperties(t *testing.T) {
	// Setup test environment
	os.Setenv("TEST_MODE", "true")

	t.Run("loads default test properties without ConfigMap", func(t *testing.T) {
		err := LoadProperties("test", nil)
		assert.NoError(t, err, "Should load test properties without error")
		assert.NotNil(t, Props, "Properties should be initialized")
	})

	t.Run("loads properties from ConfigMap", func(t *testing.T) {
		// Create a test ConfigMap
		testCM := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "alert-manager-config",
				Namespace: "alert-manager-system",
			},
			Data: map[string]string{
				common.WavefrontAPITokenK8sSecretName: "test-token-secret",
				common.WavefrontAPIUrl:                "https://test.wavefront.com",
			},
		}

		err := LoadProperties("", testCM)
		assert.NoError(t, err, "Should load properties from ConfigMap without error")
		assert.Equal(t, "test-token-secret", Props.WavefrontAPITokenSecretName())
		assert.Equal(t, "https://test.wavefront.com", Props.WavefrontAPIUrl())
	})
}

func TestUpdateProperties(t *testing.T) {
	// Setup test environment
	os.Setenv("TEST_MODE", "true")

	t.Run("skips update when ResourceVersion is the same", func(t *testing.T) {
		oldCM := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				ResourceVersion: "123",
			},
		}
		newCM := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				ResourceVersion: "123",
			},
		}

		// This should not cause any errors
		updateProperties(oldCM, newCM)
	})

	t.Run("updates properties when ResourceVersion changes", func(t *testing.T) {
		// Setup initial props
		initialCM := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				ResourceVersion: "123",
			},
			Data: map[string]string{
				common.WavefrontAPITokenK8sSecretName: "initial-token",
				common.WavefrontAPIUrl:                "https://initial.wavefront.com",
			},
		}
		LoadProperties("", initialCM)

		// Create updated CM with new ResourceVersion
		updatedCM := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				ResourceVersion: "124",
			},
			Data: map[string]string{
				common.WavefrontAPITokenK8sSecretName: "updated-token",
				common.WavefrontAPIUrl:                "https://updated.wavefront.com",
			},
		}

		// Update properties
		updateProperties(initialCM, updatedCM)

		// Verify properties were updated
		assert.Equal(t, "updated-token", Props.WavefrontAPITokenSecretName())
		assert.Equal(t, "https://updated.wavefront.com", Props.WavefrontAPIUrl())
	})
}
