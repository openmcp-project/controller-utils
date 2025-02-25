package maps_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/collections/maps"
)

var _ = Describe("LinkedIterator Tests", func() {

	Context("Merge", func() {

		It("should merge multiple maps", func() {
			m1 := map[string]string{"foo": "bar"}
			m2 := map[string]string{"bar": "baz", "foobar": "foobaz"}
			m3 := map[string]string{"abc": "def", "xyz": "uvw"}

			merged := maps.Merge(m1, m2, m3)

			Expect(merged).To(Equal(map[string]string{"foo": "bar", "bar": "baz", "foobar": "foobaz", "abc": "def", "xyz": "uvw"}))
		})

		It("should merge multiple maps with overlapping keys", func() {
			m1 := map[string]string{"foo": "bar"}
			m2 := map[string]string{"foo": "baz", "foobar": "foobaz"}
			m3 := map[string]string{"foo": "fooo", "xyz": "uvw"}

			merged := maps.Merge(m1, m2, m3)

			Expect(merged).To(Equal(map[string]string{"foo": "fooo", "foobar": "foobaz", "xyz": "uvw"}))
		})

		It("should ignore nil and empty maps", func() {
			merged := maps.Merge(nil, map[string]string{}, map[string]string{"foo": "bar"})
			Expect(merged).To(Equal(map[string]string{"foo": "bar"}))
		})

	})

	Context("Intersect", func() {

		It("should intersect multiple maps", func() {
			m1 := map[string]string{"foo": "bar", "bar": "baz", "foobar": "foobaz"}
			m2 := map[string]string{"bar": "baz", "foobar": "foobaz", "abc": "def"}
			m3 := map[string]string{"bar": "baz", "foobar": "foobaz", "abc": "def"}

			intersected := maps.Intersect(m1, m2, m3)

			Expect(intersected).To(Equal(map[string]string{"bar": "baz", "foobar": "foobaz"}))
		})

		It("should remove all entries if one map is empty", func() {
			intersected := maps.Intersect(map[string]string{"foo": "bar"}, map[string]string{})
			Expect(intersected).To(BeEmpty())

			intersected = maps.Intersect(map[string]string{}, map[string]string{"foo": "bar"})
			Expect(intersected).To(BeEmpty())
		})

	})

})
