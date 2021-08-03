package controllers_test

import (
	"context"
	"github.com/keikoproj/alert-manager/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("WavefrontalertController", func() {
	const (
		alertName      = "wavefront-test-alert"
		alertNamespace = "default"

		timeout  = time.Second * 60
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
					ExportedParams:    []string{"foo", "bar"},
				},
			}
			Expect(k8sClient.Create(ctx, alert)).Should(Succeed())

			// Lets wait until we get the above alert into informer cache
			alertLookupKey := types.NamespacedName{Name: alertName, Namespace: alertNamespace}
			createdAlert := &v1alpha1.WavefrontAlert{}

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
			f := &v1alpha1.WavefrontAlert{}
			Eventually(func() v1alpha1.State {
				k8sClient.Get(context.Background(), alertLookupKey, f)
				return f.Status.State
			}, timeout, interval).Should(Equal(v1alpha1.Error))

			By("updating the alert by adding severity")

			f.Spec.Severity = "warn"

			Expect(k8sClient.Update(context.Background(), f)).Should(Succeed())
			fetchedUpdated := &v1alpha1.WavefrontAlert{}
			Eventually(func() v1alpha1.State {
				k8sClient.Get(context.Background(), alertLookupKey, fetchedUpdated)
				return fetchedUpdated.Status.State
			}, timeout, interval).Should(Equal(v1alpha1.ReadyToBeUsed))

			//By("updating the alert by removing exported checksum- This time it should execute the request and put it in Ready state")
			//
			//fetchedUpdated.Spec.ExportedParams = []string{}
			////Wavefront call mock
			//
			//Expect(k8sClient.Update(context.Background(), fetchedUpdated)).Should(Succeed())
			//mockWavefront.EXPECT().CreateAlert(gomock.Any(), gomock.Any()).Return(nil)
			//
			//fetchedUpdated2 := &v1alpha1.WavefrontAlert{}
			//Eventually(func() v1alpha1.State {
			//	k8sClient.Get(context.Background(), alertLookupKey, fetchedUpdated2)
			//	return fetchedUpdated2.Status.State
			//}, timeout, interval).Should(Equal(v1alpha1.Ready))

			By("Deleting the alert")
			Eventually(func() error {
				f := &v1alpha1.WavefrontAlert{}
				k8sClient.Get(context.Background(), alertLookupKey, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				k := &v1alpha1.WavefrontAlert{}
				return k8sClient.Get(context.Background(), alertLookupKey, k)
			}, timeout, interval).ShouldNot(Succeed())

		})
	})

})
