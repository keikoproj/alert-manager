package utils_test

import (
	"context"
	"github.com/keikoproj/alert-manager/internal/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Util", func() {
	Describe("Test ExportParamsChecksum", func() {
		Context("empty list", func() {
			flag, resp := utils.ExportParamsChecksum(context.Background(), []string{})
			It("test case", func() {
				Expect(flag).To(BeFalse())
				Expect(resp).To(BeEmpty())
			})
		})
		Context("NON empty list", func() {
			flag, resp := utils.ExportParamsChecksum(context.Background(), []string{"cluster", "namespace"})
			It("test case", func() {
				Expect(flag).To(BeTrue())
				Expect(resp).To(Not(BeEmpty()))
			})
		})
		Context("A test to compare the difference in checksum with just one extra space", func() {
			flag, resp := utils.ExportParamsChecksum(context.Background(), []string{"cluster", "namespace"})
			It("test case -original", func() {
				Expect(flag).To(BeTrue())
				Expect(resp).To(Equal("b00f24f25192cf30021987655d7a9df8"))
			})
		})
		Context("A test to compare the difference in checksum with just one extra space", func() {
			flag, resp := utils.ExportParamsChecksum(context.Background(), []string{"cluster", "namespace "})
			It("test case -extra space", func() {
				Expect(flag).To(BeTrue())
				Expect(resp).To(Not(Equal("b00f24f25192cf30021987655d7a9df8")))
			})
		})
	})

	Describe("ContainsString() test cases", func() {
		Context("valid comparision", func() {
			It("should be true", func() {
				Expect(utils.ContainsString([]string{"iamrole.finalizers.iammanager.keikoproj.io", "iamrole.finalizers2.iammanager.keikoproj.io"}, "iamrole.finalizers.iammanager.keikoproj.io")).To(BeTrue())
			})
		})
		Context("different string comparision", func() {
			It("should return false", func() {
				Expect(utils.ContainsString([]string{"iamrole.finalizers.iammanager.keikoproj.io", "iamrole.finalizers2.iammanager.keikoproj.io"}, "iamrole-iammanager.keikoproj.io")).To(BeFalse())
			})
		})
	})

	Describe("RemoveString() test cases", func() {
		var emptySlice []string
		Context("should remove one value", func() {
			It("should be equal to the remaining string", func() {
				Expect(utils.RemoveString([]string{"iamrole.finalizers.iammanager.keikoproj.io", "iamrole.finalizers2.iammanager.keikoproj.io"}, "iamrole.finalizers.iammanager.keikoproj.io")).To(Equal([]string{"iamrole.finalizers2.iammanager.keikoproj.io"}))
			})
		})
		Context("empty slice with remove usecase", func() {
			It("should just return the empty slice", func() {
				Expect(utils.RemoveString([]string{}, "iamrole.finalizers.iammanager.keikoproj.io")).To(Equal(emptySlice))
			})
		})
		Context("empty slice with remove usecase", func() {
			It("should just return the empty slice", func() {
				Expect(utils.RemoveString([]string{}, "iamrole.finalizers.iammanager.keikoproj.io")).To(Equal(emptySlice))
			})
		})
		Context("empty the slice by removing one string", func() {
			It("should just return the empty slice", func() {
				Expect(utils.RemoveString([]string{"iamrole.finalizers.iammanager.keikoproj.io"}, "iamrole.finalizers.iammanager.keikoproj.io")).To(Equal(emptySlice))
			})
		})

		Context("trying to remove the value which doesn't exists", func() {
			It("should just return the original slice", func() {
				Expect(utils.RemoveString([]string{"iamrole.finalizers.iammanager.keikoproj.io"}, "iamrole.finalizers2.iammanager.keikoproj.io")).To(Equal([]string{"iamrole.finalizers.iammanager.keikoproj.io"}))
			})
		})
	})

})
