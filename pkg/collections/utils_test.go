package collections_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/collections"
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

})
