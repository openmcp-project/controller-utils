package filters_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/collections/filters"
)

var _ = Describe("Predefined Filters Tests", func() {

	Context("And", func() {

		It("should return true if all filters return true", func() {
			Expect(filters.And(filters.True, filters.True)(nil)).To(BeTrue())
		})

		It("should return false if any filter returns false", func() {
			Expect(filters.And(filters.True, filters.False)(nil)).To(BeFalse())
		})

	})

	Context("Or", func() {

		It("should return true if any filter returns true", func() {
			Expect(filters.Or(filters.True, filters.False)(nil)).To(BeTrue())
		})

		It("should return false if all filters return false", func() {
			Expect(filters.Or(filters.False, filters.False)(nil)).To(BeFalse())
		})

	})

	Context("Not", func() {

		It("should return the negation of the filter", func() {
			Expect(filters.Not(filters.True)(nil)).To(BeFalse())
			Expect(filters.Not(filters.False)(nil)).To(BeTrue())
		})

	})

	Context("ApplyToNthArgument", func() {

		equalsMinusThree := func(args ...any) bool {
			return args[0] == -3
		}

		It("should apply the filter to the nth argument", func() {
			Expect(filters.ApplyToNthArgument(2, equalsMinusThree)(-1, -2, -3, -4, -5)).To(BeTrue())
			Expect(filters.ApplyToNthArgument(2, equalsMinusThree)(-3, -2, -1)).To(BeFalse())
		})

	})

	Context("Wrap", func() {

		It("should wrap a function into a filter", func() {
			Expect(filters.Wrap(strings.HasPrefix, map[int]any{1: "x"})("xyz")).To(BeTrue())
			Expect(filters.Wrap(strings.HasPrefix, map[int]any{1: "x"})("abc")).To(BeFalse())
		})

	})

})
