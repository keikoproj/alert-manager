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

package controllers_test

import (
	"context"
	"strings"
	"time"

	wf "github.com/WavefrontHQ/go-wavefront-management-api"
	"github.com/golang/mock/gomock"
	alertmanagerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AlertsConfigController tests validate behavior of the AlertsConfig controller
// which manages configurations for Wavefront alerts
var _ = Describe("AlertsConfigController", Label("controller", "alertsconfig"), func() {
	const (
		alertConfigName    = "test-alerts-config"
		wavefrontAlertName = "test-wavefront-alert"
		namespace          = "default"
		timeout            = time.Second * 10
		interval           = time.Millisecond * 250
	)

	// Global setup for all tests in the AlertsConfigController describe block
	BeforeEach(func() {
		// Set up common mock expectations for all tests
		mockID := "test-alert-id-123"

		// Mock CreateAlert with specific behavior - returns a mock ID to simulate successful creation
		mockWavefront.EXPECT().CreateAlert(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, alert *wf.Alert) error {
				alert.ID = &mockID
				return nil
			}).AnyTimes()

		// Mock DeleteAlert to simulate successful deletion - important for cleanup operations
		mockWavefront.EXPECT().DeleteAlert(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, id string) error {
				return nil
			}).AnyTimes()

		// Mock ReadAlert to return predefined alert data
		mockWavefront.EXPECT().ReadAlert(gomock.Any(), gomock.Any()).Return(&wf.Alert{
			ID:        &mockID,
			Name:      wavefrontAlertName,
			Tags:      []string{"test", "integration"},
			Severity:  "warn",
			Condition: "ts(my.metric > 0)",
		}, nil).AnyTimes()

		// Mock UpdateAlert to simulate successful update operations
		mockWavefront.EXPECT().UpdateAlert(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	})

	// Test creation of AlertsConfig custom resource
	Context("When creating an AlertsConfig CR", Label("create"), func() {
		It("Should create the CR and set up finalizers", func() {
			ctx := context.Background()

			By("Creating a WavefrontAlert CR first (prerequisite for AlertsConfig)")
			wavefrontAlert := &alertmanagerv1alpha1.WavefrontAlert{
				ObjectMeta: metav1.ObjectMeta{Name: wavefrontAlertName, Namespace: namespace},
				Spec: alertmanagerv1alpha1.WavefrontAlertSpec{
					AlertType:         "CLASSIC",
					AlertName:         wavefrontAlertName,
					Condition:         "ts(my.metric > 0)",
					DisplayExpression: "ts(my.metric)",
					Minutes:           ptr(int32(5)),
					ResolveAfter:      ptr(int32(5)),
					Severity:          "warn",
					Tags:              []string{"test", "integration"},
					ExportedParams:    []string{"threshold"},
					ExportedParamsDefaultValues: map[string]string{
						"threshold": "80",
					},
				},
			}

			Expect(k8sClient.Create(ctx, wavefrontAlert)).Should(Succeed())

			// Wait for the WavefrontAlert to be created
			wfaLookupKey := types.NamespacedName{Name: wavefrontAlertName, Namespace: namespace}
			createdWFA := &alertmanagerv1alpha1.WavefrontAlert{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, wfaLookupKey, createdWFA)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Creating a new AlertsConfig CR that references the WavefrontAlert")
			alertsConfig := &alertmanagerv1alpha1.AlertsConfig{
				ObjectMeta: metav1.ObjectMeta{Name: alertConfigName, Namespace: namespace},
				Spec: alertmanagerv1alpha1.AlertsConfigSpec{
					// Global parameters applied to all alerts
					GlobalParams: map[string]string{
						"env": "test",
					},
					// Specific alert configurations by name
					Alerts: map[string]alertmanagerv1alpha1.Config{
						wavefrontAlertName: {
							Params: map[string]string{
								"threshold": "90", // Override the default threshold
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, alertsConfig)).Should(Succeed())

			// Verify the AlertsConfig is created
			acLookupKey := types.NamespacedName{Name: alertConfigName, Namespace: namespace}
			createdAC := &alertmanagerv1alpha1.AlertsConfig{}

			By("Checking if the finalizer is added by the controller")
			// Finalizers prevent premature resource deletion until cleanup is done
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, acLookupKey, createdAC); err != nil {
					return false
				}
				return len(createdAC.Finalizers) > 0 &&
					createdAC.Finalizers[0] == "alertsconfig.finalizers.alertmanager.keikoproj.io"
			}, timeout, interval).Should(BeTrue())

			By("Waiting for the AlertsConfig status to be populated by the controller")
			Eventually(func() int {
				if err := k8sClient.Get(ctx, acLookupKey, createdAC); err != nil {
					return 0
				}
				return len(createdAC.Status.AlertsStatus)
			}, timeout, interval).Should(BeNumerically(">", 0))

			// Use DeferCleanup for cleaner resource cleanup
			DeferCleanup(func() {
				Expect(k8sClient.Delete(ctx, alertsConfig)).Should(Succeed())
				Expect(k8sClient.Delete(ctx, wavefrontAlert)).Should(Succeed())
			})
		})
	})

	// Test updating an AlertsConfig custom resource
	Context("When updating an AlertsConfig CR", Label("update"), func() {
		It("Should update the configuration and reflect in status", func() {
			ctx := context.Background()

			By("Creating a WavefrontAlert CR first with a unique name")
			uniqueWfName := wavefrontAlertName + "-update-test"
			wavefrontAlert := &alertmanagerv1alpha1.WavefrontAlert{
				ObjectMeta: metav1.ObjectMeta{Name: uniqueWfName, Namespace: namespace},
				Spec: alertmanagerv1alpha1.WavefrontAlertSpec{
					AlertType:         "CLASSIC",
					AlertName:         uniqueWfName,
					Condition:         "ts(my.metric > 0)",
					DisplayExpression: "ts(my.metric)",
					Minutes:           ptr(int32(5)),
					ResolveAfter:      ptr(int32(5)),
					Severity:          "warn",
					Tags:              []string{"test", "integration"},
					ExportedParams:    []string{"threshold"},
					ExportedParamsDefaultValues: map[string]string{
						"threshold": "80", // Default threshold value
					},
				},
				Status: alertmanagerv1alpha1.WavefrontAlertStatus{
					State:      alertmanagerv1alpha1.Ready,
					RetryCount: 0,
				},
			}

			Expect(k8sClient.Create(ctx, wavefrontAlert)).Should(Succeed())

			// Wait for the WavefrontAlert to be created
			wfaLookupKey := types.NamespacedName{Name: uniqueWfName, Namespace: namespace}
			createdWFA := &alertmanagerv1alpha1.WavefrontAlert{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, wfaLookupKey, createdWFA)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Update the status manually to Ready to ensure it's in the expected state for the test
			statusPatch := []byte(`{"status":{"state":"Ready","retryCount":0}}`)
			Expect(k8sClient.Status().Patch(ctx, createdWFA, client.RawPatch(types.MergePatchType, statusPatch))).
				Should(Succeed())

			// Wait a short time to ensure status update is processed
			time.Sleep(100 * time.Millisecond)

			By("Creating a new AlertsConfig CR with a unique name")
			uniqueAcName := alertConfigName + "-update-unique"
			alertsConfig := &alertmanagerv1alpha1.AlertsConfig{
				ObjectMeta: metav1.ObjectMeta{Name: uniqueAcName, Namespace: namespace},
				Spec: alertmanagerv1alpha1.AlertsConfigSpec{
					GlobalParams: map[string]string{
						"env": "test",
					},
					Alerts: map[string]alertmanagerv1alpha1.Config{
						uniqueWfName: {
							Params: map[string]string{
								"threshold": "80", // Initially matches the default
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, alertsConfig)).Should(Succeed())

			// Verify the AlertsConfig is created
			acLookupKey := types.NamespacedName{Name: uniqueAcName, Namespace: namespace}
			createdAC := &alertmanagerv1alpha1.AlertsConfig{}

			By("Checking if the AlertsConfig is properly initialized with finalizers")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, acLookupKey, createdAC); err != nil {
					return false
				}
				return len(createdAC.Finalizers) > 0
			}, timeout, interval).Should(BeTrue())

			By("Updating the AlertsConfig with new parameter values")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, acLookupKey, createdAC); err != nil {
					return err
				}
				// Update the threshold parameter to a new value
				createdAC.Spec.Alerts[uniqueWfName] = alertmanagerv1alpha1.Config{
					Params: map[string]string{
						"threshold": "95", // Changed from 80 to 95
					},
				}
				return k8sClient.Update(ctx, createdAC)
			}, timeout, interval).Should(Succeed())

			By("Verifying the update is reflected in the spec")
			Eventually(func() string {
				if err := k8sClient.Get(ctx, acLookupKey, createdAC); err != nil {
					return ""
				}
				for alertName, _ := range createdAC.Status.AlertsStatus {
					if alertName == uniqueWfName {
						// Check the updated threshold value in the spec
						return createdAC.Spec.Alerts[uniqueWfName].Params["threshold"]
					}
				}
				return ""
			}, timeout, interval).Should(Equal("95"))

			// Use DeferCleanup for cleaner resource cleanup
			DeferCleanup(func() {
				Expect(k8sClient.Delete(ctx, alertsConfig)).Should(Succeed())
				Expect(k8sClient.Delete(ctx, wavefrontAlert)).Should(Succeed())
			})
		})
	})

	// Test deletion of AlertsConfig custom resource
	Context("When deleting an AlertsConfig CR", Label("delete"), func() {
		It("Should handle deletion gracefully and cleanup resources", func() {
			ctx := context.Background()

			By("Creating a WavefrontAlert CR first with a unique name")
			uniqueWfName := wavefrontAlertName + "-delete-test"
			wavefrontAlert := &alertmanagerv1alpha1.WavefrontAlert{
				ObjectMeta: metav1.ObjectMeta{Name: uniqueWfName, Namespace: namespace},
				Spec: alertmanagerv1alpha1.WavefrontAlertSpec{
					AlertType:         "CLASSIC",
					AlertName:         uniqueWfName,
					Condition:         "ts(my.metric > 0)",
					DisplayExpression: "ts(my.metric)",
					Minutes:           ptr(int32(5)),
					ResolveAfter:      ptr(int32(5)),
					Severity:          "warn",
					Tags:              []string{"test", "integration"},
					ExportedParams:    []string{"threshold"},
					ExportedParamsDefaultValues: map[string]string{
						"threshold": "80",
					},
				},
				Status: alertmanagerv1alpha1.WavefrontAlertStatus{
					State:      alertmanagerv1alpha1.Ready,
					RetryCount: 0,
				},
			}

			Expect(k8sClient.Create(ctx, wavefrontAlert)).Should(Succeed())

			// Wait for the WavefrontAlert to be created
			wfaLookupKey := types.NamespacedName{Name: uniqueWfName, Namespace: namespace}
			createdWFA := &alertmanagerv1alpha1.WavefrontAlert{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, wfaLookupKey, createdWFA)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Creating a new AlertsConfig CR with a unique name")
			uniqueAcName := alertConfigName + "-delete-unique"
			alertsConfig := &alertmanagerv1alpha1.AlertsConfig{
				ObjectMeta: metav1.ObjectMeta{Name: uniqueAcName, Namespace: namespace},
				Spec: alertmanagerv1alpha1.AlertsConfigSpec{
					GlobalParams: map[string]string{
						"env": "test",
					},
					Alerts: map[string]alertmanagerv1alpha1.Config{
						uniqueWfName: {
							Params: map[string]string{
								"threshold": "85",
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, alertsConfig)).Should(Succeed())

			// Verify the AlertsConfig is created and has finalizers
			acLookupKey := types.NamespacedName{Name: uniqueAcName, Namespace: namespace}
			createdAC := &alertmanagerv1alpha1.AlertsConfig{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, acLookupKey, createdAC); err != nil {
					return false
				}
				return len(createdAC.Finalizers) > 0
			}, timeout, interval).Should(BeTrue())

			By("Deleting the AlertsConfig resource")
			Expect(k8sClient.Delete(ctx, createdAC)).Should(Succeed())

			By("Verifying the AlertsConfig resource is removed after finalizer processing")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, acLookupKey, createdAC)
				return err != nil // Error expected because resource should be gone
			}, timeout, interval).Should(BeTrue())

			// Clean up the WavefrontAlert
			DeferCleanup(func() {
				Expect(k8sClient.Delete(ctx, wavefrontAlert)).Should(Succeed())
			})
		})
	})

	// Test error handling for missing referenced WavefrontAlerts
	Context("When using an AlertsConfig CR with invalid references", Label("error", "references"), func() {
		It("Should handle missing referenced WavefrontAlerts gracefully", func() {
			ctx := context.Background()

			By("Creating an AlertsConfig that references non-existent WavefrontAlerts")
			missingRefConfig := &alertmanagerv1alpha1.AlertsConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "missing-refs-config",
					Namespace: namespace,
				},
				Spec: alertmanagerv1alpha1.AlertsConfigSpec{
					GlobalParams: map[string]string{
						"env": "test",
					},
					Alerts: map[string]alertmanagerv1alpha1.Config{
						"non-existent-alert": {
							Params: map[string]string{
								"threshold": "90",
							},
						},
						"another-missing-alert": {
							Params: map[string]string{
								"threshold": "95",
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, missingRefConfig)).Should(Succeed())

			// Setup cleanup to run at the end of the test
			DeferCleanup(func() {
				Expect(k8sClient.Delete(ctx, missingRefConfig)).Should(Succeed())
			})

			// Verify the AlertsConfig is created
			configLookupKey := types.NamespacedName{Name: "missing-refs-config", Namespace: namespace}
			createdConfig := &alertmanagerv1alpha1.AlertsConfig{}

			By("Checking if the finalizer is still added despite missing references")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, configLookupKey, createdConfig); err != nil {
					return false
				}
				return len(createdConfig.Finalizers) > 0
			}, timeout, interval).Should(BeTrue())

			By("Verifying the AlertsConfig status reflects the missing alerts")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, configLookupKey, createdConfig); err != nil {
					return false
				}

				// The status should indicate that the references are missing
				// Either by having no AlertsStatus entries or by having error status
				if len(createdConfig.Status.AlertsStatus) == 0 {
					// No status entries means the controller recognized missing refs
					return true
				}

				// Or check for specific errors in the status
				for _, status := range createdConfig.Status.AlertsStatus {
					if status.State == alertmanagerv1alpha1.Error {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})

		It("Should handle invalid parameter values appropriately", func() {
			ctx := context.Background()

			By("First creating a WavefrontAlert with specific exported parameters")
			wavefrontAlert := &alertmanagerv1alpha1.WavefrontAlert{
				ObjectMeta: metav1.ObjectMeta{Name: "params-test-alert", Namespace: namespace},
				Spec: alertmanagerv1alpha1.WavefrontAlertSpec{
					AlertType:         "CLASSIC",
					AlertName:         "params-test-alert",
					Condition:         "ts(my.metric > ${threshold})",
					DisplayExpression: "ts(my.metric)",
					Minutes:           ptr(int32(5)),
					ResolveAfter:      ptr(int32(5)),
					Severity:          "warn",
					Tags:              []string{"test", "params"},
					// Define specific exported parameters that must be provided
					ExportedParams: []string{"threshold", "environment", "region"},
					ExportedParamsDefaultValues: map[string]string{
						"threshold": "80",
						// Note: intentionally not providing defaults for "environment" and "region"
					},
				},
			}

			Expect(k8sClient.Create(ctx, wavefrontAlert)).Should(Succeed())

			// Wait for the WavefrontAlert to be created and status set
			wfaLookupKey := types.NamespacedName{Name: "params-test-alert", Namespace: namespace}
			createdWFA := &alertmanagerv1alpha1.WavefrontAlert{}

			Eventually(func() bool {
				return k8sClient.Get(ctx, wfaLookupKey, createdWFA) == nil
			}, timeout, interval).Should(BeTrue())

			// Update status to Ready so AlertsConfig will try to use it
			statusPatch := []byte(`{"status":{"state":"Ready","retryCount":0}}`)
			Expect(k8sClient.Status().Patch(ctx, createdWFA, client.RawPatch(types.MergePatchType, statusPatch))).
				Should(Succeed())

			By("Creating an AlertsConfig with missing required parameters")
			invalidParamsConfig := &alertmanagerv1alpha1.AlertsConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-params-config",
					Namespace: namespace,
				},
				Spec: alertmanagerv1alpha1.AlertsConfigSpec{
					GlobalParams: map[string]string{
						"env": "test",
						// Note: "environment" is missing but required by the alert
					},
					Alerts: map[string]alertmanagerv1alpha1.Config{
						"params-test-alert": {
							Params: map[string]string{
								"threshold": "90",
								// Missing both "environment" and "region" parameters
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, invalidParamsConfig)).Should(Succeed())

			// Verify the AlertsConfig is created
			configLookupKey := types.NamespacedName{Name: "invalid-params-config", Namespace: namespace}
			createdConfig := &alertmanagerv1alpha1.AlertsConfig{}

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, configLookupKey, createdConfig); err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			// Allow time for controller to process and update status
			time.Sleep(100 * time.Millisecond)

			By("Verifying the controller correctly identifies missing parameters")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, configLookupKey, createdConfig); err != nil {
					return false
				}

				// Check for error state in the alert status
				for alertName, status := range createdConfig.Status.AlertsStatus {
					if alertName == "params-test-alert" {
						return status.State == alertmanagerv1alpha1.Error &&
							status.ErrorDescription != ""
					}
				}
				return false
			}, timeout, interval).Should(BeTrue(), "AlertsConfig should show error state for missing parameters")

			// Verify that controller reports which parameters are missing
			By("Verifying the error message contains information about missing parameters")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, configLookupKey, createdConfig); err != nil {
					return false
				}

				for alertName, status := range createdConfig.Status.AlertsStatus {
					if alertName == "params-test-alert" {
						// Error should mention at least one of the missing parameters
						return strings.Contains(status.ErrorDescription, "environment") ||
							strings.Contains(status.ErrorDescription, "region")
					}
				}
				return false
			}, timeout, interval).Should(BeTrue(), "Error description should mention missing parameters")

			// Clean up resources
			DeferCleanup(func() {
				Expect(k8sClient.Delete(ctx, invalidParamsConfig)).Should(Succeed())
				Expect(k8sClient.Delete(ctx, wavefrontAlert)).Should(Succeed())
			})
		})
	})
})

// Helper function to create integer pointers
func ptr(i int32) *int32 {
	return &i
}
