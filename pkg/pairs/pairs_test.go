package pairs_test

import (
	"cmp"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/openmcp-project/controller-utils/pkg/pairs"
)

func TestConditions(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "ClusterAccess Test Suite")
}

type comparableIntAlias int

var _ pairs.Comparable[comparableIntAlias] = comparableIntAlias(0)

// Compare implements pairs.Comparable.
// This inverses the usual order.
func (c comparableIntAlias) CompareTo(other comparableIntAlias) int {
	return cmp.Compare(int(c)*(-1), int(other)*(-1))
}

type comparableStruct struct {
	val int
}

func (c *comparableStruct) CompareTo(other *comparableStruct) int {
	return cmp.Compare(c.val, other.val)
}

var _ = Describe("Pairs", func() {

	Context("Compare", func() {

		It("should be able to compare int-based keys", func() {
			type intAlias int
			a := pairs.New(intAlias(1), "foo")
			b := pairs.New(intAlias(2), "bar")
			Expect(a.CompareTo(b)).To(Equal(-1))
			Expect(b.CompareTo(a)).To(Equal(1))
			Expect(a.CompareTo(a)).To(Equal(0))
		})

		It("should be able to compare float-based keys", func() {
			type floatAlias float64
			a := pairs.New(floatAlias(1.1), "foo")
			b := pairs.New(floatAlias(1.2), "bar")
			Expect(a.CompareTo(b)).To(Equal(-1))
			Expect(b.CompareTo(a)).To(Equal(1))
			Expect(a.CompareTo(a)).To(Equal(0))
		})

		It("should be able to compare string-based keys", func() {
			type stringAlias string
			a := pairs.New(stringAlias("a"), "foo")
			b := pairs.New(stringAlias("b"), "bar")
			Expect(a.CompareTo(b)).To(Equal(-1))
			Expect(b.CompareTo(a)).To(Equal(1))
			Expect(a.CompareTo(a)).To(Equal(0))
		})

		It("should prefer Comparable implementation over default comparison", func() {
			a := pairs.New(comparableIntAlias(1), "foo")
			b := pairs.New(comparableIntAlias(2), "bar")
			// inversed order, because of the Compare() implementation
			Expect(a.CompareTo(b)).To(Equal(1))
			Expect(b.CompareTo(a)).To(Equal(-1))
			Expect(a.CompareTo(a)).To(Equal(0))
		})

		It("should detect if *K does implement Comparable", func() {
			a := pairs.New(comparableStruct{val: 1}, "foo")
			b := pairs.New(comparableStruct{val: 2}, "bar")
			Expect(a.CompareTo(b)).To(Equal(-1))
			Expect(b.CompareTo(a)).To(Equal(1))
			Expect(a.CompareTo(a)).To(Equal(0))
		})

		It("should panic if K does not implement Comparable and is not based on a comparable type", func() {
			type notComparable struct{}
			a := pairs.New(notComparable{}, "foo")
			b := pairs.New(notComparable{}, "bar")
			Expect(func() { a.CompareTo(b) }).To(Panic())
			Expect(func() { b.CompareTo(a) }).To(Panic())
			Expect(func() { a.CompareTo(a) }).To(Panic())
		})

	})

	Context("Conversion", func() {

		It("should convert a map to a list of pairs", func() {
			src := map[string]string{
				"foo": "bar",
				"baz": "asdf",
			}
			pairs := pairs.MapToPairs(src)
			Expect(pairs).To(ConsistOf(
				MatchFields(0, Fields{
					"Key":   Equal("foo"),
					"Value": Equal("bar"),
				}),
				MatchFields(0, Fields{
					"Key":   Equal("baz"),
					"Value": Equal("asdf"),
				}),
			))
		})

		It("should convert a list of pairs to a map", func() {
			src := []pairs.Pair[string, string]{
				pairs.New("foo", "foo"),
				pairs.New("foo", "bar"),
				pairs.New("baz", "asdf"),
			}
			m := pairs.PairsToMap(src)
			Expect(m).To(HaveLen(2))
			Expect(m).To(HaveKeyWithValue("foo", "bar"))
			Expect(m).To(HaveKeyWithValue("baz", "asdf"))
		})

	})

})
