package template_test

import (
	"context"
	"encoding/json"

	"github.com/keikoproj/alert-manager/api/v1alpha1"
	"github.com/keikoproj/alert-manager/internal/template"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Template", func() {
	Describe("Test ProcessTemplate", func() {
		wfAlert := v1alpha1.WavefrontAlert{
			Spec: v1alpha1.WavefrontAlertSpec{
				AlertName:   "TestAlert1",
				Condition:   "Something",
				Description: "description",
			},
		}
		Context("No golang template variables", func() {
			It("Should not error out", func() {
				wfAlert0 := wfAlert
				tempBytes, _ := json.Marshal(wfAlert0.Spec)
				resp, _ := template.ProcessTemplate(context.Background(), string(tempBytes), make(map[string]string))
				Expect(resp).To(Equal(string(tempBytes)))
			})
		})
		Context("With one golang template variable", func() {
			wfAlertX := wfAlert
			wfAlertX.Spec.Condition = "some {{ .exportedParam1 }} variable"
			wfAlertX.Spec.ExportedParams = []string{"exportedParam1"}
			It("with variable passed and Should not error out", func() {
				wfAlert1 := wfAlertX
				tempBytes, _ := json.Marshal(wfAlert1.Spec)
				resp, _ := template.ProcessTemplate(context.Background(), string(tempBytes), map[string]string{
					"exportedParam1": "golang",
				})
				wfAlert1.Spec.Condition = "some golang variable"
				expBytes, _ := json.Marshal(wfAlert1.Spec)
				Expect(resp).To(Equal(string(expBytes)))
			})
			It("with extra variables passed and Should NOT error out", func() {
				wfAlert2 := wfAlertX
				tempBytes, _ := json.Marshal(wfAlert2.Spec)
				resp, _ := template.ProcessTemplate(context.Background(), string(tempBytes), map[string]string{
					"exportedParam1": "golang",
					"exportedParam2": "golang234234",
				})
				wfAlert2.Spec.Condition = "some golang variable"
				expBytes, _ := json.Marshal(wfAlert2.Spec)
				Expect(resp).To(Equal(string(expBytes)))
			})
		})

	})
})
