package maps_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/collections/maps"
)

var _ = Describe("Map Utils Tests", func() {

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

	Context("GetAny", func() {

		It("should return a key-value pair from the map", func() {
			m1 := map[string]string{"foo": "bar", "bar": "baz", "foobar": "foobaz"}
			pair := maps.GetAny(m1)
			Expect(pair).ToNot(BeNil())
			Expect(pair.Key).To(BeElementOf("foo", "bar", "foobar"))
			Expect(m1[pair.Key]).To(Equal(pair.Value))
		})

		It("should return nil for an empty or nil map", func() {
			var nilMap map[string]string
			Expect(maps.GetAny(nilMap)).To(BeNil())
			Expect(maps.GetAny(map[string]string{})).To(BeNil())
		})

	})

})
