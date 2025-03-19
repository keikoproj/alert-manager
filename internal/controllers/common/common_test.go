package common_test

import (
	"context"
	"fmt"

	wf "github.com/WavefrontHQ/go-wavefront-management-api"
	alertmanagerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	"github.com/keikoproj/alert-manager/internal/controllers/common"
	"github.com/keikoproj/alert-manager/pkg/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Common", func() {

	const (
		alertName      = "wavefront-test-alert"
		alertNamespace = "default"
	)

	Context("Patch Status  and Update status test cases", func() {
		It("should work as expected", func() {
			By("create a new wavefront alert (similar to kubectl create")
			ctx := context.Background()
			var minutes int32
			var resolveAfterMinutes int32
			minutes = 5
			resolveAfterMinutes = 5

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
					Severity:          "severe",
					Tags:              []string{"test"},
				},
				Status: alertmanagerv1alpha1.WavefrontAlertStatus{
					RetryCount: 0,
					State:      alertmanagerv1alpha1.Creating,
				},
			}

			k8sClientObj := &k8s.Client{
				Cl: fake.NewSimpleClientset(),
			}

			By("testing update status on wavefront alert")
			alert.Status.ObservedGeneration = alert.Generation
			alert.Status.State = alertmanagerv1alpha1.Ready

			fakeClient := k8sClient

			commonClient := common.Client{
				Client:   fakeClient,
				Recorder: k8sClientObj.SetUpEventHandler(context.Background()),
			}
			By("testing update patch status on wavefront alert")
			patch := []byte(fmt.Sprintf("{\"status\":{\"state\": \"%s\"}}", alertmanagerv1alpha1.Ready))
			_, err := commonClient.PatchStatus(ctx, alert, client.RawPatch(types.MergePatchType, patch), alertmanagerv1alpha1.Ready)
			Expect(err).To(BeNil())
			// Should be in Ready state since it is hard coded patched
			f2 := &alertmanagerv1alpha1.WavefrontAlert{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: alertName, Namespace: alertNamespace}, f2)
			Expect(err).ToNot(BeNil())
			// We expect error since we never created the object. We just used the fake client to test the interface

			var requeueTime float64 = 5
			_, err = commonClient.UpdateStatus(ctx, alert, alertmanagerv1alpha1.Ready, requeueTime)
			Expect(err).To(BeNil())

			By("testing convert alertcr function")
			wavefrontAlert := &wf.Alert{}
			commonClient.ConvertAlertCR(ctx, alert, wavefrontAlert)

			Expect(wavefrontAlert.Name).To(Equal(alert.Spec.AlertName))
			Expect(int32(wavefrontAlert.Minutes)).To(Equal(*alert.Spec.Minutes))
			Expect(int32(wavefrontAlert.ResolveAfterMinutes)).To(Equal(*alert.Spec.ResolveAfter))
			Expect(wavefrontAlert.Target).To(Equal(alert.Spec.Target))
			Expect(wavefrontAlert.Condition).To(Equal(alert.Spec.Condition))
			Expect(wavefrontAlert.DisplayExpression).To(Equal(alert.Spec.DisplayExpression))
			Expect(wavefrontAlert.Severity).To(Equal(alert.Spec.Severity))
			Expect(wavefrontAlert.Tags).To(Equal(alert.Spec.Tags))

		})
	})

	Context("GetProcessedWFAlert test cases", func() {

		wfAlert := &alertmanagerv1alpha1.WavefrontAlert{
			Spec: alertmanagerv1alpha1.WavefrontAlertSpec{
				AlertType: "CLASSIC",
				AlertName: "alert-template-{{ .appName }}",
				Condition: "{{.condition }}",
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
