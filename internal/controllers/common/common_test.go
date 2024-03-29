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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Common", func() {

	const (
		alertName      = "wavefront-test-alert"
		alertNamespace = "default"

		timeout  = time.Second * 60
		duration = time.Second * 10
		interval = time.Millisecond * 250
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
					Tags:              []string{"foo", "bar"},
				},
			}
			Expect(k8sClient.Create(ctx, alert)).Should(Succeed())

			// Lets wait until we get the above alert into informer cache
			alertLookupKey := types.NamespacedName{Name: alertName, Namespace: alertNamespace}
			createdAlert := &alertmanagerv1alpha1.WavefrontAlert{}

			// We'll need to retry getting this newly created Alert, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, alertLookupKey, createdAlert)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our alert value is there by verifying the condition
			Expect(createdAlert.Spec.Condition).Should(Equal("ts(status.health)"))
			// Should throw error since severity is missing in the request
			f := &alertmanagerv1alpha1.WavefrontAlert{}
			Eventually(func() alertmanagerv1alpha1.State {
				k8sClient.Get(context.Background(), alertLookupKey, f)
				return f.Status.State
			}, timeout, interval).Should(Equal(alertmanagerv1alpha1.Error))

			commonClient := common.Client{
				Client:   k8sClient,
				Recorder: k8sCl.SetUpEventHandler(context.Background()),
			}
			By("testing update patch status on wavefront alert")
			patch := []byte(fmt.Sprintf("{\"status\":{\"state\": \"%s\"}}", alertmanagerv1alpha1.Ready))
			_, err := commonClient.PatchStatus(ctx, alert, client.RawPatch(types.MergePatchType, patch), alertmanagerv1alpha1.Ready)
			Expect(err).To(BeNil())
			// Should be in Ready state since it is hard coded patched
			f2 := &alertmanagerv1alpha1.WavefrontAlert{}
			Eventually(func() alertmanagerv1alpha1.State {
				k8sClient.Get(context.Background(), alertLookupKey, f2)
				return f2.Status.State
			}, timeout, interval).Should(Equal(alertmanagerv1alpha1.Ready))
			f2.Status.State = alertmanagerv1alpha1.Error
			f2.Status.ErrorDescription = "something"
			f2.Status.RetryCount = 2

			By("Testing update status function on wavefront alert")
			_, err = commonClient.UpdateStatus(ctx, f2, alertmanagerv1alpha1.Error, 90000)
			Expect(err).To(BeNil())
			f3 := &alertmanagerv1alpha1.WavefrontAlert{}
			Eventually(func() alertmanagerv1alpha1.State {
				k8sClient.Get(context.Background(), alertLookupKey, f3)
				return f3.Status.State
			}, timeout, interval).Should(Equal(alertmanagerv1alpha1.Error))
			Expect(f3.Status.ErrorDescription).Should(Equal("something"))
		})
	})

	Context("PatchWfAlertAndAlertsConfigStatus test cases", func() {
		It("create alerts config", func() {

		})
	})
	Context("GetProcessedWFAlert test cases", func() {

		wfAlert := &alertmanagerv1alpha1.WavefrontAlert{
			Spec: alertmanagerv1alpha1.WavefrontAlertSpec{
				AlertType: "CLASSIC",
				Severity:  "{{ .foo }}",
				ExportedParams: []string{
					"foo",
				},
			},
		}

		config := &alertmanagerv1alpha1.Config{
			Params: map[string]string{
				"foobar": "barfoo",
			},
		}

		It("invalid template params", func() {
			var alert wf.Alert
			err := common.GetProcessedWFAlert(context.Background(), wfAlert, config.Params, &alert)
			Expect(err).NotTo(BeNil())
		})

		It("invalid spec to convert to wavefront request", func() {
			config.Params["foo"] = "$foo"
			var alert wf.Alert
			err := common.GetProcessedWFAlert(context.Background(), wfAlert, config.Params, &alert)
			Expect(err).NotTo(BeNil())
		})

		It("no severity found in wavefront request", func() {
			config.Params["foo"] = "bar"
			min := int32(5)
			wfAlert.Spec.Minutes = &min
			wfAlert.Spec.ResolveAfter = &min
			var alert wf.Alert
			err := common.GetProcessedWFAlert(context.Background(), wfAlert, config.Params, &alert)
			Expect(err).NotTo(BeNil())
		})

		It("no condition in the request", func() {
			config.Params["foo"] = "warn"
			min := int32(5)
			wfAlert.Spec.Minutes = &min
			wfAlert.Spec.ResolveAfter = &min
			var alert wf.Alert
			err := common.GetProcessedWFAlert(context.Background(), wfAlert, config.Params, &alert)
			Expect(err).NotTo(BeNil())
		})
	})
})
