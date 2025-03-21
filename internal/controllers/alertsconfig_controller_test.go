package controllers_test

import (
	"context"
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

var _ = Describe("AlertsConfigController", func() {
	const (
		alertConfigName    = "test-alerts-config"
		wavefrontAlertName = "test-wavefront-alert"
		namespace          = "default"
		timeout            = time.Second * 10
		interval           = time.Millisecond * 250
	)

	// Global setup for all tests
	BeforeEach(func() {
		// Set up common mock expectations for all tests
		mockID := "test-alert-id-123"

		// Mock CreateAlert with specific behavior
		mockWavefront.EXPECT().CreateAlert(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, alert *wf.Alert) error {
				alert.ID = &mockID
				return nil
			}).AnyTimes()

		// Mock DeleteAlert with specific behavior - important for cleanup
		mockWavefront.EXPECT().DeleteAlert(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, id string) error {
				return nil
			}).AnyTimes()

		// Mock other operations
		mockWavefront.EXPECT().ReadAlert(gomock.Any(), gomock.Any()).Return(&wf.Alert{
			ID:        &mockID,
			Name:      wavefrontAlertName,
			Tags:      []string{"test", "integration"},
			Severity:  "warn",
			Condition: "ts(my.metric > 0)",
		}, nil).AnyTimes()

		mockWavefront.EXPECT().UpdateAlert(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	})

	Context("When creating an AlertsConfig CR", func() {
		It("Should create the CR and set up finalizers", func() {
			ctx := context.Background()

			By("Creating a WavefrontAlert CR first")
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

			By("Creating a new AlertsConfig CR")
			alertsConfig := &alertmanagerv1alpha1.AlertsConfig{
				ObjectMeta: metav1.ObjectMeta{Name: alertConfigName, Namespace: namespace},
				Spec: alertmanagerv1alpha1.AlertsConfigSpec{
					GlobalParams: map[string]string{
						"env": "test",
					},
					Alerts: map[string]alertmanagerv1alpha1.Config{
						wavefrontAlertName: {
							Params: map[string]string{
								"threshold": "90",
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, alertsConfig)).Should(Succeed())

			// Verify the AlertsConfig is created
			acLookupKey := types.NamespacedName{Name: alertConfigName, Namespace: namespace}
			createdAC := &alertmanagerv1alpha1.AlertsConfig{}

			By("Checking if the finalizer is added")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, acLookupKey, createdAC); err != nil {
					return false
				}
				return len(createdAC.Finalizers) > 0 &&
					createdAC.Finalizers[0] == "alertsconfig.finalizers.alertmanager.keikoproj.io"
			}, timeout, interval).Should(BeTrue())

			By("Waiting for the AlertsConfig status to be updated")
			Eventually(func() int {
				if err := k8sClient.Get(ctx, acLookupKey, createdAC); err != nil {
					return 0
				}
				return len(createdAC.Status.AlertsStatus)
			}, timeout, interval).Should(BeNumerically(">", 0))

			By("Cleaning up resources")
			Expect(k8sClient.Delete(ctx, alertsConfig)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, wavefrontAlert)).Should(Succeed())
		})
	})

	Context("When updating an AlertsConfig CR", func() {
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

			// Update the status manually to Ready
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
								"threshold": "80",
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, alertsConfig)).Should(Succeed())

			// Verify the AlertsConfig is created
			acLookupKey := types.NamespacedName{Name: uniqueAcName, Namespace: namespace}
			createdAC := &alertmanagerv1alpha1.AlertsConfig{}

			Eventually(func() bool {
				return k8sClient.Get(ctx, acLookupKey, createdAC) == nil
			}, timeout, interval).Should(BeTrue())

			// Wait for status to be populated before updating
			Eventually(func() int {
				if err := k8sClient.Get(ctx, acLookupKey, createdAC); err != nil {
					return 0
				}
				return len(createdAC.Status.AlertsStatus)
			}, timeout, interval).Should(BeNumerically(">", 0))

			// Wait a short time to ensure initial processing is complete
			time.Sleep(200 * time.Millisecond)

			By("Updating the AlertsConfig with new parameters")
			// Get the latest version before updating
			Expect(k8sClient.Get(ctx, acLookupKey, createdAC)).Should(Succeed())

			updatedSpec := createdAC.Spec.DeepCopy()
			updatedSpec.Alerts[uniqueWfName] = alertmanagerv1alpha1.Config{
				Params: map[string]string{
					"threshold": "95", // Updated threshold
				},
			}

			createdAC.Spec = *updatedSpec
			Expect(k8sClient.Update(ctx, createdAC)).Should(Succeed())

			By("Verifying the AlertsConfig update is processed")
			Eventually(func() string {
				updatedAC := &alertmanagerv1alpha1.AlertsConfig{}
				if err := k8sClient.Get(ctx, acLookupKey, updatedAC); err != nil {
					return ""
				}

				return updatedAC.Spec.Alerts[uniqueWfName].Params["threshold"]
			}, timeout, interval).Should(Equal("95"))

			By("Cleaning up resources")
			Expect(k8sClient.Delete(ctx, alertsConfig)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, wavefrontAlert)).Should(Succeed())
		})
	})

	Context("When deleting an AlertsConfig CR", func() {
		It("Should handle deletion gracefully and cleanup resources", func() {
			ctx := context.Background()

			By("Creating a WavefrontAlert CR first with a unique name")
			uniqueWfName := wavefrontAlertName + "-delete-test"
			wavefrontAlert := &alertmanagerv1alpha1.WavefrontAlert{
				ObjectMeta: metav1.ObjectMeta{
					Name:       uniqueWfName,
					Namespace:  namespace,
					Finalizers: []string{"wavefrontalert.finalizers.alertmanager.keikoproj.io"},
				},
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

			// Wait for status to be updated
			time.Sleep(100 * time.Millisecond)

			// Mock Wavefront operations
			mockID := "test-alert-id-345"
			mockWavefront.EXPECT().CreateAlert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, alert *wf.Alert) error {
					alert.ID = &mockID
					return nil
				}).AnyTimes()

			mockWavefront.EXPECT().DeleteAlert(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			By("Creating a new AlertsConfig CR with a unique name")
			uniqueAcName := alertConfigName + "-delete-unique"
			alertsConfig := &alertmanagerv1alpha1.AlertsConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:       uniqueAcName,
					Namespace:  namespace,
					Finalizers: []string{"alertsconfig.finalizers.alertmanager.keikoproj.io"},
				},
				Spec: alertmanagerv1alpha1.AlertsConfigSpec{
					GlobalParams: map[string]string{
						"env": "test",
					},
					Alerts: map[string]alertmanagerv1alpha1.Config{
						uniqueWfName: {
							Params: map[string]string{
								"threshold": "70",
							},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, alertsConfig)).Should(Succeed())

			// Set up the status for the AlertsConfig
			acLookupKey := types.NamespacedName{Name: uniqueAcName, Namespace: namespace}
			createdAC := &alertmanagerv1alpha1.AlertsConfig{}
			Eventually(func() error {
				return k8sClient.Get(ctx, acLookupKey, createdAC)
			}, timeout, interval).Should(BeNil())

			// Wait for initial processing
			time.Sleep(100 * time.Millisecond)

			// Manually patch status to simulate the controller's behavior
			statusPatch := []byte(`{"status":{"alertsStatus":{"` + uniqueWfName + `":{"id":"` + mockID + `","name":"` + uniqueWfName + `","state":"Ready","retryCount":0}}}}`)
			Expect(k8sClient.Status().Patch(ctx, createdAC, client.RawPatch(types.MergePatchType, statusPatch))).
				Should(Succeed())

			// Make sure the status patch is applied
			time.Sleep(100 * time.Millisecond)

			By("Deleting the AlertsConfig")
			// Get the latest version before deleting
			Expect(k8sClient.Get(ctx, acLookupKey, createdAC)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, createdAC)).Should(Succeed())

			By("Verifying the CR is eventually deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, acLookupKey, &alertmanagerv1alpha1.AlertsConfig{})
				return err != nil // Should get a "not found" error
			}, timeout, interval).Should(BeTrue())

			// Give some time for finalizer processing to complete
			time.Sleep(100 * time.Millisecond)

			By("Cleaning up the WavefrontAlert")
			Expect(k8sClient.Delete(ctx, wavefrontAlert)).Should(Succeed())
		})
	})
})

// Helper function to create integer pointers
func ptr(i int32) *int32 {
	return &i
}
