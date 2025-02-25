package filters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/collections/filters"
)

var _ = Describe("Number Filters Tests", func() {

	Context("NumericallyGreaterThan", func() {

		It("should return true if the number is greater than the comparison value", func() {
			Expect(filters.NumericallyGreaterThan(0)(1)).To(BeTrue())
			Expect(filters.NumericallyGreaterThan(0)(0)).To(BeFalse())
			Expect(filters.NumericallyGreaterThan(0)(-1)).To(BeFalse())
		})

	})

	Context("NumericallyLessThan", func() {

		It("should return true if the number is less than the comparison value", func() {
			Expect(filters.NumericallyLessThan(0)(-1)).To(BeTrue())
			Expect(filters.NumericallyLessThan(0)(0)).To(BeFalse())
			Expect(filters.NumericallyLessThan(0)(1)).To(BeFalse())
		})

	})

	Context("NumericallyEqualTo", func() {

		It("should return true if the number is equal to the comparison value", func() {
			Expect(filters.NumericallyEqualTo(0)(0)).To(BeTrue())
			Expect(filters.NumericallyEqualTo(0)(1)).To(BeFalse())
			Expect(filters.NumericallyEqualTo(0)(-1)).To(BeFalse())
		})

	})

	Context("NumericallyGreaterThanOrEqualTo", func() {

		It("should return true if the number is greater than or equal to the comparison value", func() {
			Expect(filters.NumericallyGreaterThanOrEqualTo(0)(1)).To(BeTrue())
			Expect(filters.NumericallyGreaterThanOrEqualTo(0)(0)).To(BeTrue())
			Expect(filters.NumericallyGreaterThanOrEqualTo(0)(-1)).To(BeFalse())
		})

	})

	Context("NumericallyLessThanOrEqualTo", func() {

		It("should return true if the number is less than or equal to the comparison value", func() {
			Expect(filters.NumericallyLessThanOrEqualTo(0)(-1)).To(BeTrue())
			Expect(filters.NumericallyLessThanOrEqualTo(0)(0)).To(BeTrue())
			Expect(filters.NumericallyLessThanOrEqualTo(0)(1)).To(BeFalse())
		})

	})

})
