package collections_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/collections"
	"github.com/openmcp-project/controller-utils/pkg/pairs"
)

var _ = Describe("Utils Tests", func() {

	Context("ProjectSlice", func() {

		projectFunc := func(i int) int {
			return i * 2
		}

		It("should use the projection function on each element of the slice", func() {
			src := []int{1, 2, 3, 4}
			projected := collections.ProjectSlice(src, projectFunc)
			Expect(projected).To(Equal([]int{2, 4, 6, 8}))
			Expect(src).To(Equal([]int{1, 2, 3, 4}), "original slice should not be modified")
		})

		It("should return an empty slice for an empty or nil input slice", func() {
			Expect(collections.ProjectSlice(nil, projectFunc)).To(BeEmpty())
			Expect(collections.ProjectSlice([]int{}, projectFunc)).To(BeEmpty())
		})

		It("should return nil for a nil projection function", func() {
			src := []int{1, 2, 3, 4}
			projected := collections.ProjectSlice[int, int](src, nil)
			Expect(projected).To(BeNil())
			Expect(src).To(Equal([]int{1, 2, 3, 4}), "original slice should not be modified")
		})

	})

	Context("ProjectMapToSlice", func() {

		projectFunc := func(k string, v string) string {
			return k + ":" + v
		}

		It("should use the projection function on each key-value pair of the map", func() {
			src := map[string]string{"a": "1", "b": "2", "c": "3"}
			projected := collections.ProjectMapToSlice(src, projectFunc)
			Expect(projected).To(ConsistOf("a:1", "b:2", "c:3"))
			Expect(src).To(Equal(map[string]string{"a": "1", "b": "2", "c": "3"}), "original map should not be modified")
		})

		It("should return an empty slice for an empty or nil input map", func() {
			Expect(collections.ProjectMapToSlice(nil, projectFunc)).To(BeEmpty())
			Expect(collections.ProjectMapToSlice(map[string]string{}, projectFunc)).To(BeEmpty())
		})

		It("should return nil for a nil projection function", func() {
			src := map[string]string{"a": "1", "b": "2", "c": "3"}
			projected := collections.ProjectMapToSlice[string, string, string](src, nil)
			Expect(projected).To(BeNil())
			Expect(src).To(Equal(map[string]string{"a": "1", "b": "2", "c": "3"}), "original map should not be modified")
		})

	})

	Context("ProjectMapToMap", func() {

		projectFunc := func(k string, v string) (string, int) {
			return k, len(v)
		}

		It("should use the projection function on each key-value pair of the map", func() {
			src := map[string]string{"a": "1", "b": "22", "c": "333"}
			projected := collections.ProjectMapToMap(src, projectFunc)
			Expect(projected).To(Equal(map[string]int{"a": 1, "b": 2, "c": 3}))
			Expect(src).To(Equal(map[string]string{"a": "1", "b": "22", "c": "333"}), "original map should not be modified")
		})

		It("should return an empty map for an empty or nil input map", func() {
			Expect(collections.ProjectMapToMap(nil, projectFunc)).To(BeEmpty())
			Expect(collections.ProjectMapToMap(map[string]string{}, projectFunc)).To(BeEmpty())
		})

		It("should return nil for a nil projection function", func() {
			src := map[string]string{"a": "1", "b": "22", "c": "333"}
			projected := collections.ProjectMapToMap[string, string, string, int](src, nil)
			Expect(projected).To(BeNil())
			Expect(src).To(Equal(map[string]string{"a": "1", "b": "22", "c": "333"}), "original map should not be modified")
		})

	})

	Context("AggregateSlice", func() {

		sum := func(val, s int) int {
			return val + s
		}
		stradd := func(val int, s string) string {
			return fmt.Sprintf("%s%d", s, val)
		}

		It("should return the initial value if the aggregation function is nil", func() {
			src := []int{1, 2, 3, 4}
			result := collections.AggregateSlice(src, nil, 0)
			Expect(result).To(Equal(0))
			Expect(src).To(Equal([]int{1, 2, 3, 4}))
		})

		It("should correctly aggregate the slice using the provided function", func() {
			src := []int{1, 2, 3, 4}
			result := collections.AggregateSlice(src, sum, 0)
			Expect(result).To(Equal(10))
			Expect(src).To(Equal([]int{1, 2, 3, 4}))

			result2 := collections.AggregateSlice(src, stradd, "test")
			Expect(result2).To(Equal("test1234"))
			Expect(src).To(Equal([]int{1, 2, 3, 4}))
		})

		It("should handle a nil input slice", func() {
			result := collections.AggregateSlice[int, int](nil, sum, 100)
			Expect(result).To(Equal(100))
		})

	})

	Context("AggregateMap", func() {

		aggregate := func(k string, v int, agg pairs.Pair[string, int]) pairs.Pair[string, int] {
			return pairs.New(agg.Key+k, agg.Value+v)
		}

		It("should return the initial value if the aggregation function is nil", func() {
			src := map[string]int{"a": 1, "b": 2, "c": 3}
			result := collections.AggregateMap(src, nil, 0)
			Expect(result).To(Equal(0))
			Expect(src).To(Equal(map[string]int{"a": 1, "b": 2, "c": 3}))
		})

		It("should correctly aggregate the map using the provided function", func() {
			src := map[string]int{"a": 1, "b": 2, "c": 3}
			result := collections.AggregateMap(src, aggregate, pairs.New("", 0))
			Expect(result.Key).To(HaveLen(3))
			Expect(result.Key).To(ContainSubstring("a"))
			Expect(result.Key).To(ContainSubstring("b"))
			Expect(result.Key).To(ContainSubstring("c"))
			Expect(result.Value).To(Equal(6))
			Expect(src).To(Equal(map[string]int{"a": 1, "b": 2, "c": 3}))
		})

		It("should handle a nil input map", func() {
			result := collections.AggregateMap(nil, aggregate, pairs.New("", 0))
			Expect(result.Key).To(BeEmpty())
			Expect(result.Value).To(Equal(0))
		})

	})

})
