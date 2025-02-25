package collections_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/collections"
	cerr "github.com/openmcp-project/controller-utils/pkg/collections/errors"
)

type queueImplementation struct {
	Name             string
	Constructor      func(elems ...int) collections.Queue[int]
	AllowsDuplicates bool
}

var _ = Describe("Queue Tests", func() {

	for _, impl := range []*queueImplementation{
		{
			Name:             "LinkedList",
			Constructor:      func(elems ...int) collections.Queue[int] { return collections.NewLinkedList[int](elems...) },
			AllowsDuplicates: true,
		},
	} {
		runQueueTests(impl)
	}

})

func runQueueTests(impl *queueImplementation) {
	var q collections.Queue[int]

	BeforeEach(func() {
		q = impl.Constructor(baseData...)
	})

	AfterEach(func() {
		Expect(baseData).To(Equal([]int{1, 3, 2, 4}), "array passed into constructor should not be modified")
	})

	Context(impl.Name, func() {

		Context("Queue-specific functionality", func() {

			It("should add elements to the queue", func() {
				Expect(q.Push(5)).To(Succeed())
				Expect(q.ToSlice()).To(Equal([]int{1, 3, 2, 4, 5}))
				Expect(q.Push(6, 7)).To(Succeed())
				Expect(q.ToSlice()).To(Equal([]int{1, 3, 2, 4, 5, 6, 7}))
			})

			It("should retrieve the first element without modifying it for peek/element", func() {
				Expect(q.Peek()).To(Equal(1))
				Expect(q.ToSlice()).To(Equal([]int{1, 3, 2, 4}))
				val, err := q.Element()
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(1))
			})

			It("should show correct error behavior for peek/element", func() {
				var eq collections.Queue[int] = collections.NewLinkedList[int]()
				Expect(eq.Peek()).To(Equal(0))
				_, err := eq.Element()
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeEquivalentTo(cerr.CollectionEmptyError{}))
			})

			It("should remove and return the first element for poll/fetch", func() {
				s := q.Size()
				Expect(q.Poll()).To(Equal(1))
				Expect(q.Size()).To(Equal(s - 1))
				s = q.Size()
				Expect(q.Poll()).To(Equal(3))
				Expect(q.Size()).To(Equal(s - 1))
				s = q.Size()
				val, err := q.Fetch()
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(2))
				Expect(q.Size()).To(Equal(s - 1))
				s = q.Size()
				val, err = q.Fetch()
				Expect(err).ToNot(HaveOccurred())
				Expect(val).To(Equal(4))
				Expect(q.Size()).To(Equal(s - 1))
			})

			It("should show correct error behavior for poll/fetch", func() {
				var eq collections.Queue[int] = collections.NewLinkedList[int]()
				Expect(eq.Poll()).To(Equal(0))
				_, err := eq.Fetch()
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeEquivalentTo(cerr.CollectionEmptyError{}))
			})

		})

	})

}
