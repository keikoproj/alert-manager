package wavefront_test

import (
	"context"
	"github.com/WavefrontHQ/go-wavefront-management-api"
	wf "github.com/keikoproj/alert-manager/pkg/wavefront"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validate", func() {
	Describe(" Test ValidateAlertInput", func() {
		Context("CLASSIC type test cases ", func() {
			input := &wavefront.Alert{
				Name: "test-alert",
			}
			It("AlertType is empty", func() {
				err := wf.ValidateAlertInput(context.Background(), input)
				Expect(err).NotTo(BeNil())
			})

			It("Invalid AlertType", func() {
				input.AlertType = "foo"
				err := wf.ValidateAlertInput(context.Background(), input)
				Expect(err).NotTo(BeNil())
			})
			It("Condition empty test case", func() {
				input.AlertType = "CLASSIC"
				err := wf.ValidateAlertInput(context.Background(), input)
				Expect(err).NotTo(BeNil())
			})
			It("Severity is empty", func() {
				input.Condition = "ts(status.health)"
				err := wf.ValidateAlertInput(context.Background(), input)
				Expect(err).NotTo(BeNil())
			})
			It("Invalid Severity", func() {
				input.Severity = "foo"
				err := wf.ValidateAlertInput(context.Background(), input)
				Expect(err).NotTo(BeNil())
			})
			It("Successful input use case", func() {
				input.Severity = "warn"
				err := wf.ValidateAlertInput(context.Background(), input)
				Expect(err).To(BeNil())
			})
		})

		Context("THRESHOLD type test cases ", func() {
			input := &wavefront.Alert{
				Name: "test-alert2",
			}
			It("AlertType is empty", func() {
				err := wf.ValidateAlertInput(context.Background(), input)
				Expect(err).NotTo(BeNil())
			})

			It("Invalid AlertType", func() {
				input.AlertType = "foo"
				err := wf.ValidateAlertInput(context.Background(), input)
				Expect(err).NotTo(BeNil())
			})
			It("Conditions empty test case", func() {
				input.AlertType = "THRESHOLD"
				err := wf.ValidateAlertInput(context.Background(), input)
				Expect(err).NotTo(BeNil())
			})
			It("invalid severity", func() {
				input.Conditions = map[string]string{
					"foo": "bar",
				}
				err := wf.ValidateAlertInput(context.Background(), input)
				Expect(err).NotTo(BeNil())
			})
			It("Successful input use case", func() {
				input.Conditions = map[string]string{
					"warn": "bar",
				}
				err := wf.ValidateAlertInput(context.Background(), input)
				Expect(err).To(BeNil())
			})
		})

		Context("ExportParam with config Param comparision test", func() {
			It("Send 1 value instead of 2 required values", func() {
				err := wf.ValidateTemplateParams(context.Background(), []string{"foo", "bar"}, map[string]string{
					"foo": "bar",
				})
				Expect(err).NotTo(BeNil())
			})
			It("Send wrong param name", func() {
				err := wf.ValidateTemplateParams(context.Background(), []string{"foo"}, map[string]string{
					"bar": "bar",
				})
				Expect(err).NotTo(BeNil())
			})
			It("Successful usecase", func() {
				err := wf.ValidateTemplateParams(context.Background(), []string{"foo"}, map[string]string{
					"foo": "bar",
				})
				Expect(err).To(BeNil())
			})
		})
	})

})
