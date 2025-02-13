package filters_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.tools.sap/CoLa/controller-utils/pkg/collections"
	"github.tools.sap/CoLa/controller-utils/pkg/collections/filters"
)

var _ = Describe("Filter Tests", func() {

	Context("FilterSlice", func() {

		origin := []int{-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5} // comparison value for example
		example := []int{-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5}

		It("should filter the slice", func() {
			Expect(filters.FilterSlice(example, filters.NumericallyGreaterThan(0))).To(Equal([]int{1, 2, 3, 4, 5}))
			Expect(example).To(Equal(origin), "original slice should not be modified")
		})

		It("should return an empty slice if no elements match the filter", func() {
			Expect(filters.FilterSlice(example, filters.NumericallyGreaterThan(10))).To(BeEmpty())
			Expect(example).To(Equal(origin), "original slice should not be modified")
		})

	})

	Context("FilterMap", func() {

		origin := map[int]int{-5: 5, -4: 4, -3: 3, -2: 2, -1: 1, 0: 0, 1: -1, 2: -2, 3: -3, 4: -4, 5: -5} // comparison value for example
		example := map[int]int{-5: 5, -4: 4, -3: 3, -2: 2, -1: 1, 0: 0, 1: -1, 2: -2, 3: -3, 4: -4, 5: -5}

		It("should filter the map", func() {
			Expect(filters.FilterMap(example, func(args ...any) bool {
				return args[0] == args[1]
			})).To(Equal(map[int]int{0: 0}))
			Expect(example).To(Equal(origin), "original map should not be modified")
		})

		It("should filter the map based on the key", func() {
			Expect(filters.FilterMap(example, filters.ApplyToNthArgument(0, filters.NumericallyGreaterThan(0)))).To(Equal(map[int]int{1: -1, 2: -2, 3: -3, 4: -4, 5: -5}))
			Expect(example).To(Equal(origin), "original map should not be modified")
		})

		It("should filter the map based on the value", func() {
			Expect(filters.FilterMap(example, filters.ApplyToNthArgument(1, filters.NumericallyGreaterThan(0)))).To(Equal(map[int]int{-5: 5, -4: 4, -3: 3, -2: 2, -1: 1}))
			Expect(example).To(Equal(origin), "original map should not be modified")
		})

	})

	Context("FilterCollection", func() {

		origin := collections.NewLinkedList(-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5) // comparison value for example
		example := collections.NewLinkedList(-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5)

		It("should filter the collection", func() {
			filtered := filters.FilterCollection(example, filters.NumericallyGreaterThan(0))
			Expect(filtered.ToSlice()).To(Equal([]int{1, 2, 3, 4, 5}))
			Expect(reflect.TypeOf(filtered)).To(Equal(reflect.TypeOf(example)))
			Expect(example.ToSlice()).To(Equal(origin.ToSlice()), "original collection should not be modified")
		})

		It("should return an empty collection if no elements match the filter", func() {
			filtered := filters.FilterCollection(example, filters.NumericallyGreaterThan(10))
			Expect(filtered.ToSlice()).To(BeEmpty())
			Expect(reflect.TypeOf(filtered)).To(Equal(reflect.TypeOf(example)))
			Expect(example.ToSlice()).To(Equal(origin.ToSlice()), "original collection should not be modified")
		})

	})

})
