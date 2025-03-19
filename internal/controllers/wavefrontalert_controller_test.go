package controllers_test

import (
	"context"
	"time"

	"github.com/keikoproj/alert-manager/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("WavefrontalertController", func() {
	const (
		alertName      = "wavefront-test-alert"
		alertNamespace = "default"

		// Increase timeout to give controller more time to update status
		timeout  = time.Minute * 5
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("Single Alert creation", func() {

		It("It should be able to create an alert", func() {
			By("create a new wavefront alert (similar to kubectl create")
			ctx := context.Background()
			var minutes int32
			var resolveAfterMinutes int32
			minutes = 5
			resolveAfterMinutes = 5

			alert := &v1alpha1.WavefrontAlert{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "alertmanager.keikoproj.io/v1alpha1",
					Kind:       "WavefrontAlert",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:       alertName,
					Namespace:  alertNamespace,
					Finalizers: []string{"wavefrontalert.finalizers.alertmanager.keikoproj.io"},
				},
				Spec: v1alpha1.WavefrontAlertSpec{
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
			createdAlert := &v1alpha1.WavefrontAlert{}

			// We'll need to retry getting this newly created Alert, given that creation may not immediately happen.
			Eventually(func() bool {
				return k8sClient.Get(ctx, alertLookupKey, createdAlert) == nil
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our alert value is there by verifying the condition
			Expect(createdAlert.Spec.Condition).Should(Equal("ts(status.health)"))
			// Should throw error since severity is missing in the request
			f := &v1alpha1.WavefrontAlert{}
			Eventually(func() v1alpha1.State {
				err := k8sClient.Get(context.Background(), alertLookupKey, f)
				if err != nil {
					GinkgoT().Logf("Error getting WavefrontAlert: %v", err)
					return ""
				}
				return f.Status.State
			}, timeout, interval).Should(Equal(v1alpha1.Error))

			By("updating the alert by adding severity")

			f.Spec.Severity = "warn"
			f.Spec.ExportedParams = []string{"foo", "bar"}

			Expect(k8sClient.Update(context.Background(), f)).Should(Succeed())

			// We'll need to retry getting this newly created Alert, given that creation may not immediately happen.
			Eventually(func() bool {
				return k8sClient.Get(ctx, alertLookupKey, createdAlert) == nil
			}, timeout, interval).Should(BeTrue())
			fetchedUpdated := &v1alpha1.WavefrontAlert{}
			Eventually(func() v1alpha1.State {
				err := k8sClient.Get(context.Background(), alertLookupKey, fetchedUpdated)
				if err != nil {
					GinkgoT().Logf("Error getting updated WavefrontAlert: %v", err)
					return ""
				}
				return fetchedUpdated.Status.State
			}, timeout, interval).Should(Equal(v1alpha1.ReadyToBeUsed))

			By("Deleting the alert")
			Eventually(func() error {
				f := &v1alpha1.WavefrontAlert{}
				if err := k8sClient.Get(context.Background(), alertLookupKey, f); err != nil {
					return err
				}
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				k := &v1alpha1.WavefrontAlert{}
				return k8sClient.Get(context.Background(), alertLookupKey, k)
			}, timeout, interval).ShouldNot(Succeed())

		})
	})

})
