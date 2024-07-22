package wavefront_test

import (
	"context"

	wf "github.com/WavefrontHQ/go-wavefront-management-api"
	alertmanagerv1alpha1 "github.com/keikoproj/alert-manager/api/v1alpha1"
	"github.com/keikoproj/alert-manager/pkg/wavefront"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conversion", func() {
	Context("Alert conversion to wavefront request", func() {
		wfAlert := &alertmanagerv1alpha1.WavefrontAlert{
			Spec: alertmanagerv1alpha1.WavefrontAlertSpec{
				AlertType: "CLASSIC",
				Severity:  "{{ .foo }}",
				ExportedParams: []string{
					"foo",
				},
			},
		}
		It("Minutes in nil", func() {
			var alert wf.Alert
			err := wavefront.ConvertAlertCRToWavefrontRequest(context.Background(), wfAlert.Spec, &alert)
			Expect(err).NotTo(BeNil())
		})
		It("resolveMinutes in nil", func() {
			mins := int32(5)
			wfAlert.Spec.Minutes = &mins
			var alert wf.Alert
			err := wavefront.ConvertAlertCRToWavefrontRequest(context.Background(), wfAlert.Spec, &alert)
			Expect(err).NotTo(BeNil())
		})

		It("resolveMinutes in nil", func() {
			mins := int32(5)
			wfAlert.Spec.Minutes = &mins
			wfAlert.Spec.ResolveAfter = &mins
			var alert wf.Alert
			err := wavefront.ConvertAlertCRToWavefrontRequest(context.Background(), wfAlert.Spec, &alert)
			Expect(err).To(BeNil())
		})

	})

})
