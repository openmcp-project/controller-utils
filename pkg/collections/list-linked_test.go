package collections

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LinkedList Integrity Validation", func() {

	It("should not violate the list's integrity", func() {
		li := NewLinkedList[int](1, 2, 3, 4)
		li.validateIntegrity("after creation")
		li.Add(5)
		li.validateIntegrity("after Add")
		li.Remove(2)
		li.validateIntegrity("after Remove")
		Expect(li.RemoveIndex(0)).To(Succeed())
		li.validateIntegrity("after RemoveIndex(0)")
		Expect(li.RemoveIndex(li.size - 1)).To(Succeed())
		li.validateIntegrity("after RemoveIndex(size-1)")
		Expect(li.AddIndex(0, 0)).To(Succeed())
		li.validateIntegrity("after AddIndex to beginning of list")
		Expect(li.AddIndex(8, li.size)).To(Succeed())
		li.validateIntegrity("after AddIndex to end of list")
		Expect(li.Push(5, 6)).To(Succeed())
		li.validateIntegrity("after Push")
		li.Poll()
		li.validateIntegrity("after Poll")
		_, err := li.Fetch()
		Expect(err).ToNot(HaveOccurred())
		li.validateIntegrity("after Fetch")
	})

})

// validateIntegrity checks that for each element e
// - e.next.prev == e
// - e.prev.next == e
func (l *LinkedList[T]) validateIntegrity(msg string) {
	elem := l.dummy
	Expect(elem.next).ToNot(BeNil(), msg)
	Expect(elem.prev).ToNot(BeNil(), msg)
	Expect(elem.next.prev).To(Equal(elem), msg)
	Expect(elem.prev.next).To(Equal(elem), msg)
	elem = elem.next
	for elem != l.dummy {
		Expect(elem.next).ToNot(BeNil(), msg)
		Expect(elem.prev).ToNot(BeNil(), msg)
		Expect(elem.next.prev).To(Equal(elem), msg)
		Expect(elem.prev.next).To(Equal(elem), msg)
		elem = elem.next
	}
}
