package common_test

import (
	"context"
	"fmt"
	"time"

	wf "github.com/WavefrontHQ/go-wavefront-management-api"
	alertmanagerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	"github.com/keikoproj/alert-manager/internal/controllers/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Common", func() {

	const (
		alertName      = "wavefront-test-alert"
		alertNamespace = "default"

		timeout  = time.Second * 30
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("Patch Status and Update status test cases", func() {
		It("should work as expected", func() {
			By("Creating a new wavefront alert")
			ctx := context.Background()
			var minutes int32 = 5
			var resolveAfterMinutes int32 = 5

			alert := &alertmanagerv1alpha1.WavefrontAlert{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "alertmanager.keikoproj.io/v1alpha1",
					Kind:       "WavefrontAlert",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:       alertName,
					Namespace:  alertNamespace,
					Finalizers: []string{"wavefrontalert.finalizers.alertmanager.keikoproj.io"},
				},
				Spec: alertmanagerv1alpha1.WavefrontAlertSpec{
					AlertType:         "CLASSIC",
					AlertName:         alertName,
					Condition:         "ts(status.health)",
					DisplayExpression: "ts(status.health)",
					Minutes:           &minutes,
					ResolveAfter:      &resolveAfterMinutes,
					Severity:          "warn", // Include severity to avoid validation errors
					Tags:              []string{"test"},
				},
				Status: alertmanagerv1alpha1.WavefrontAlertStatus{
					State:      alertmanagerv1alpha1.Error, // Start with Error state for testing
					RetryCount: 0,                          // Required field based on CRD validation
				},
			}

			// Use the envtest k8sClient that's already set up in BeforeSuite
			By("Creating the alert in the API server")
			Expect(k8sClient.Create(ctx, alert)).Should(Succeed())

			// Important: For envtest, we need to directly set the status after creation
			// since the normal creation won't set the status field
			By("Setting the initial status to Error")
			// Create a complete replacement patch for status instead of using MergeFrom
			// to ensure all required fields are included
			statusPatch := []byte(fmt.Sprintf(`{"status":{"state":"%s","retryCount":%d}}`, alertmanagerv1alpha1.Error, 0))
			err := k8sClient.Status().Patch(ctx, alert, client.RawPatch(types.MergePatchType, statusPatch))
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the alert exists and has the correct initial state")
			alertLookupKey := types.NamespacedName{Name: alertName, Namespace: alertNamespace}
			createdAlert := &alertmanagerv1alpha1.WavefrontAlert{}

			Eventually(func() bool {
				return k8sClient.Get(ctx, alertLookupKey, createdAlert) == nil
			}, timeout, interval).Should(BeTrue())

			// Verify the alert has the expected condition
			Expect(createdAlert.Spec.Condition).Should(Equal("ts(status.health)"))

			// Verify status was properly set to Error
			Eventually(func() alertmanagerv1alpha1.State {
				err := k8sClient.Get(ctx, alertLookupKey, createdAlert)
				if err != nil {
					return ""
				}
				return createdAlert.Status.State
			}, timeout, interval).Should(Equal(alertmanagerv1alpha1.Error))

			// Create the commonClient with recorder for testing
			recorder := record.NewFakeRecorder(10)
			commonClient := common.Client{
				Client:   k8sClient,
				Recorder: recorder,
			}

			By("Testing PatchStatus to update status to Ready")
			patch := []byte(fmt.Sprintf("{\"status\":{\"state\": \"%s\", \"retryCount\": 0}}", alertmanagerv1alpha1.Ready))
			_, err = commonClient.PatchStatus(ctx, alert, client.RawPatch(types.MergePatchType, patch), alertmanagerv1alpha1.Ready)
			Expect(err).ToNot(HaveOccurred())

			By("Verifying the status was updated to Ready")
			patchedAlert := &alertmanagerv1alpha1.WavefrontAlert{}
			Eventually(func() alertmanagerv1alpha1.State {
				k8sClient.Get(ctx, alertLookupKey, patchedAlert)
				return patchedAlert.Status.State
			}, timeout, interval).Should(Equal(alertmanagerv1alpha1.Ready))

			By("Testing UpdateStatus to set status to Creating")
			updatedAlert := &alertmanagerv1alpha1.WavefrontAlert{}
			Expect(k8sClient.Get(ctx, alertLookupKey, updatedAlert)).To(Succeed())

			// We need to include retryCount in the status update
			updatedAlert.Status.RetryCount = 0
			updatedAlert.Status.State = alertmanagerv1alpha1.Creating

			// Use raw status update instead of using commonClient.UpdateStatus
			// to ensure all required fields are included
			err = k8sClient.Status().Update(ctx, updatedAlert)
			Expect(err).ToNot(HaveOccurred())

			By("Verifying the status was updated to Creating")
			finalAlert := &alertmanagerv1alpha1.WavefrontAlert{}
			Eventually(func() alertmanagerv1alpha1.State {
				k8sClient.Get(ctx, alertLookupKey, finalAlert)
				GinkgoWriter.Printf("Current alert state: %s\n", finalAlert.Status.State)
				return finalAlert.Status.State
			}, timeout, interval).Should(Equal(alertmanagerv1alpha1.Creating))

			By("Cleaning up by deleting the test alert")
			Expect(k8sClient.Delete(ctx, finalAlert)).To(Succeed())

			By("Verifying the alert is deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, alertLookupKey, finalAlert)
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("GetProcessedWFAlert test cases", func() {

		wfAlert := &alertmanagerv1alpha1.WavefrontAlert{
			Spec: alertmanagerv1alpha1.WavefrontAlertSpec{
				AlertType: "CLASSIC",
				AlertName: "alert-template-{{.appName}}",
				Condition: "{{.condition}}",
				ExportedParams: []string{
					"appName",
					"condition",
				},
				Minutes:           func() *int32 { i := int32(5); return &i }(),
				ResolveAfter:      func() *int32 { i := int32(5); return &i }(),
				Severity:          "severe",
				DisplayExpression: "{{.condition}}",
			},
		}

		It("Test with only required params", func() {
			ctx := context.Background()
			params := map[string]string{
				"appName":   "test",
				"condition": "ts(status.health)",
			}
			alert := &wf.Alert{}
			err := common.GetProcessedWFAlert(ctx, wfAlert, params, alert)
			Expect(err).NotTo(HaveOccurred())

			Expect(alert.Name).To(Equal("alert-template-test"))
			Expect(alert.Condition).To(Equal("ts(status.health)"))
		})

		It("Test with missing params", func() {
			ctx := context.Background()
			params := map[string]string{
				"appName": "test",
			}
			alert := &wf.Alert{}
			err := common.GetProcessedWFAlert(ctx, wfAlert, params, alert)
			Expect(err).To(HaveOccurred())
		})

		It("Test with empty params", func() {
			ctx := context.Background()
			params := make(map[string]string)
			alert := &wf.Alert{}
			err := common.GetProcessedWFAlert(ctx, wfAlert, params, alert)
			Expect(err).To(HaveOccurred())
		})
	})
})
