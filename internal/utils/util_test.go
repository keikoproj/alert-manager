package utils_test

import (
	"context"
	"fmt"

	"github.com/keikoproj/alert-manager/api/v1alpha1"
	"github.com/keikoproj/alert-manager/internal/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Util", func() {
	Describe("Test MergeMaps function", func() {
		Context("empty global map", func() {
			result := utils.MergeMaps(context.Background(), map[string]string{}, map[string]string{
				"foo": "bar",
				"xyz": "xyz",
				"abc": "abc",
			})
			It("should have no issues", func() {
				Expect(result).To(Equal(map[string]string{
					"abc": "abc",
					"foo": "bar",
					"xyz": "xyz",
				}))
			})
		})
		Context("empty overwrite map", func() {
			result := utils.MergeMaps(context.Background(), map[string]string{
				"foo": "bar",
				"xyz": "xyz",
				"abc": "abc",
			}, map[string]string{})
			It("should have no issues", func() {
				Expect(result).To(Equal(map[string]string{
					"abc": "abc",
					"foo": "bar",
					"xyz": "xyz",
				}))
			})
		})
		Context(" overwrite map", func() {
			result := utils.MergeMaps(context.Background(), map[string]string{
				"foo": "bar",
				"xyz": "xyz",
				"abc": "abc",
			}, map[string]string{
				"foo": "foo",
			})
			It("should have no issues", func() {
				Expect(result).To(Equal(map[string]string{
					"abc": "abc",
					"foo": "foo",
					"xyz": "xyz",
				}))
			})
		})
	})

	Describe("Test CalculateAlertConfigChecksum ", func() {
		Context("Valid case", func() {
			flag, resp := utils.CalculateAlertConfigChecksum(context.Background(), v1alpha1.Config{
				Params: map[string]string{
					"foo": "bar",
				},
			}, map[string]string{})
			It("resp should not be empty", func() {
				Expect(flag).To(BeTrue())
				Expect(resp).To(Equal("cc57a3bd14055f7abafa726652854821175ae72e09a32cf9baad513d2084f493"))
			})
		})
		Context("Overwriting global param", func() {
			flag, resp := utils.CalculateAlertConfigChecksum(context.Background(), v1alpha1.Config{
				Params: map[string]string{
					"foo": "bar",
				},
			}, map[string]string{
				"foo": "foo",
			})
			It("resp should not be empty", func() {
				Expect(flag).To(BeTrue())
				Expect(resp).To(Equal("cc57a3bd14055f7abafa726652854821175ae72e09a32cf9baad513d2084f493"))
			})
		})
		Context("only global param", func() {
			flag, resp := utils.CalculateAlertConfigChecksum(context.Background(), v1alpha1.Config{
				Params: map[string]string{},
			}, map[string]string{
				"foo": "bar",
			})
			It("resp should not be empty", func() {
				Expect(flag).To(BeTrue())
				Expect(resp).To(Equal("cc57a3bd14055f7abafa726652854821175ae72e09a32cf9baad513d2084f493"))
			})
		})
		Context("order shouldn't matter- test case1", func() {
			flag1, resp1 := utils.CalculateAlertConfigChecksum(context.Background(), v1alpha1.Config{
				Params: map[string]string{
					"bar": "bar",
				},
			}, map[string]string{
				"foo": "bar",
				"abc": "abc",
			})
			flag2, resp2 := utils.CalculateAlertConfigChecksum(context.Background(), v1alpha1.Config{
				Params: map[string]string{
					"bar": "bar",
					"foo": "bar",
				},
			}, map[string]string{
				"abc": "abc",
			})
			It("hash values should match regardless of order", func() {
				Expect(flag1).To(BeTrue())
				Expect(flag2).To(BeTrue())
				Expect(resp1).To(Equal("29e1aa772b11ab2825b0225c72cd8da8ce70b42cf6ff5f4b92de6ef4dae79f39"))
				Expect(resp2).To(Equal("29e1aa772b11ab2825b0225c72cd8da8ce70b42cf6ff5f4b92de6ef4dae79f39"))
				Expect(resp1).To(Equal(resp2))
			})
		})
		Context("order shouldn't matter- test case2", func() {
			flag1, resp1 := utils.CalculateAlertConfigChecksum(context.Background(), v1alpha1.Config{
				Params: map[string]string{
					"bar": "bar",
				},
			}, map[string]string{
				"foo": "bar",
				"abc": "abc",
			})
			flag2, resp2 := utils.CalculateAlertConfigChecksum(context.Background(), v1alpha1.Config{
				Params: map[string]string{
					"bar": "bar",
					"foo": "bar",
				},
			}, map[string]string{
				"abc": "abc",
			})
			It("hash values should match regardless of order", func() {
				Expect(flag1).To(BeTrue())
				Expect(flag2).To(BeTrue())
				Expect(resp1).To(Equal("29e1aa772b11ab2825b0225c72cd8da8ce70b42cf6ff5f4b92de6ef4dae79f39"))
				Expect(resp2).To(Equal("29e1aa772b11ab2825b0225c72cd8da8ce70b42cf6ff5f4b92de6ef4dae79f39"))
				Expect(resp1).To(Equal(resp2))
			})
		})
	})

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
		Context("valid test case", func() {
			flag, resp := utils.ExportParamsChecksum(context.Background(), []string{"cluster", "namespace"})
			It("test case -original", func() {
				Expect(flag).To(BeTrue())
				Expect(resp).To(Equal("b86a15f46b81ab2b1e203f0a1c3ac48e1568582493718dff21fa11d55967b108"))
			})
		})
		Context("A test to compare the difference in checksum with just one extra space", func() {
			flag, resp := utils.ExportParamsChecksum(context.Background(), []string{"cluster", "namespace "})
			It("test case -extra space", func() {
				Expect(flag).To(BeTrue())
				Expect(resp).To(Equal("b393526f3634932caf1bf3becebab8c6d15daa0caf6699cb96b2d78a8f41da67"))
				Expect(resp).ToNot(Equal("b86a15f46b81ab2b1e203f0a1c3ac48e1568582493718dff21fa11d55967b108"))
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
				Expect(utils.RemoveString(emptySlice, "iamrole.finalizers.iammanager.keikoproj.io")).To(Equal(emptySlice))
			})
		})
		Context("empty slice with remove usecase", func() {
			It("should just return the empty slice", func() {
				Expect(utils.RemoveString(emptySlice, "iamrole.finalizers.iammanager.keikoproj.io")).To(Equal(emptySlice))
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

	Describe("CalculateChecksum() test cases", func() {
		Context("with simple string input", func() {
			It("should return consistent hash value", func() {
				checksum := utils.CalculateChecksum(context.Background(), "test-input")
				fmt.Printf("DEBUG: input: %q, checksum: %q\n", "test-input", checksum)
				Expect(checksum).To(Equal("ae1608896372720b6ebb58261e0c0092c608324b0804bc99267c1753990faaa8"))

				// Test idempotence - should return same hash for same input
				checksum2 := utils.CalculateChecksum(context.Background(), "test-input")
				Expect(checksum).To(Equal(checksum2))
			})
		})

		Context("with empty string input", func() {
			It("should return hash of empty string", func() {
				checksum := utils.CalculateChecksum(context.Background(), "")
				Expect(checksum).To(Equal("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"))
			})
		})

		Context("with different inputs", func() {
			It("should return different hashes", func() {
				checksum1 := utils.CalculateChecksum(context.Background(), "input1")
				checksum2 := utils.CalculateChecksum(context.Background(), "input2")
				Expect(checksum1).NotTo(Equal(checksum2))
			})
		})
	})

	Describe("TrimSpaces() test cases", func() {
		Context("with string input containing spaces", func() {
			It("should trim leading and trailing spaces", func() {
				result := utils.TrimSpaces("  test string with spaces  ")
				Expect(result).To(Equal("test string with spaces"))
			})
		})

		Context("with non-string input", func() {
			It("should return empty string", func() {
				result := utils.TrimSpaces(123)
				Expect(result).To(Equal(""))
			})
		})

		Context("with empty string", func() {
			It("should return empty string", func() {
				result := utils.TrimSpaces("")
				Expect(result).To(Equal(""))
			})
		})
	})

	Describe("TrimSpacesMap() test cases", func() {
		Context("with map containing strings with spaces", func() {
			It("should trim all values", func() {
				input := map[string]string{
					"key1": "  value1  ",
					"key2": "value2",
					"key3": " value3 with spaces ",
				}

				result := utils.TrimSpacesMap(input)

				Expect(result["key1"]).To(Equal("value1"))
				Expect(result["key2"]).To(Equal("value2"))
				Expect(result["key3"]).To(Equal("value3 with spaces"))
			})
		})

		Context("with empty map", func() {
			It("should return empty map", func() {
				input := map[string]string{}
				result := utils.TrimSpacesMap(input)
				Expect(result).To(Equal(map[string]string{}))
			})
		})
	})
})
