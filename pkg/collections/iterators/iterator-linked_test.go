package iterators_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/collections/iterators"
)

type linkedDummy struct {
	value int
	next  *linkedDummy
}

var _ iterators.LinkedIteratorElement[int] = &linkedDummy{}

func (d *linkedDummy) Value() int {
	return d.value
}

func (d *linkedDummy) Next() iterators.LinkedIteratorElement[int] {
	return d.next
}

func testDummy(vs ...int) *linkedDummy {
	if len(vs) == 0 {
		return nil
	}
	return &linkedDummy{vs[0], testDummy(vs[1:]...)}
}

var testDummyIsValidFunc = func(lie iterators.LinkedIteratorElement[int]) bool {
	return !reflect.ValueOf(lie).IsNil()
}

var _ = Describe("LinkedIterator Tests", func() {

	It("should iterate over the linked struct", func() {
		example := testDummy(2, 1, 0)
		it := iterators.NewLinkedIterator(example, testDummyIsValidFunc)
		Expect(it.HasNext()).To(BeTrue())
		Expect(it.Next()).To(Equal(2))
		Expect(it.HasNext()).To(BeTrue())
		Expect(it.Next()).To(Equal(1))
		Expect(it.HasNext()).To(BeTrue())
		Expect(it.Next()).To(Equal(0))
		Expect(it.HasNext()).To(BeFalse())
	})

	It("should return an 'empty' iterator for an empty linked struct", func() {
		example := testDummy()
		it := iterators.NewLinkedIterator(example, testDummyIsValidFunc)
		Expect(it.HasNext()).To(BeFalse())
	})

})
