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
