package collections_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/collections"
)

type collectionImplementation struct {
	Name             string
	Constructor      func(elems ...int) collections.Collection[int]
	AllowsDuplicates bool
}

var _ = Describe("Collection Tests", func() {

	for _, impl := range []*collectionImplementation{
		{
			Name:             "LinkedList",
			Constructor:      func(elems ...int) collections.Collection[int] { return collections.NewLinkedList[int](elems...) },
			AllowsDuplicates: true,
		},
	} {
		runCollectionTests(impl)
	}

})

func runCollectionTests(impl *collectionImplementation) {
	var col collections.Collection[int]

	BeforeEach(func() {
		col = impl.Constructor(baseData...)
	})

	AfterEach(func() {
		Expect(baseData).To(Equal([]int{1, 3, 2, 4}), "array passed into constructor should not be modified")
	})

	Context(impl.Name, func() {

		Context("Converting to slice", func() {

			It("should convert an empty collection into an empty slice", func() {
				Expect(impl.Constructor().ToSlice()).To(Equal([]int{}))
			})

			It("should convert the collection into a slice", func() {
				Expect(col.ToSlice()).To(Equal(baseData))
			})

		})

		Context("Iterator", func() {

			It("should return a working iterator", func() {
				it := col.Iterator()
				Expect(it.HasNext()).To(BeTrue())
				Expect(it.Next()).To(Equal(1))
				Expect(it.HasNext()).To(BeTrue())
				Expect(it.Next()).To(Equal(3))
				Expect(it.HasNext()).To(BeTrue())
				Expect(it.Next()).To(Equal(2))
				Expect(it.HasNext()).To(BeTrue())
				Expect(it.Next()).To(Equal(4))
				Expect(it.HasNext()).To(BeFalse())
			})

		})

		Context("Adding elements", func() {

			It("should add a single element", func() {
				oldSize := col.Size()
				Expect(col.Add(5)).To(BeTrue())
				Expect(col.Size()).To(Equal(oldSize + 1))
				Expect(col.ToSlice()).To(Equal(append(baseData, 5)))
			})

			It("should add multiple elements", func() {
				oldSize := col.Size()
				Expect(col.Add(5, 6, 7)).To(BeTrue())
				Expect(col.Size()).To(Equal(oldSize + 3))
				Expect(col.ToSlice()).To(Equal(append(baseData, 5, 6, 7)))
			})

			It("should construct an empty list and add multiple elements to it", func() {
				col2 := impl.Constructor()
				Expect(col2.Size()).To(Equal(0))
				Expect(col2.AddAll(col)).To(BeTrue())
				Expect(col2.ToSlice()).To(Equal(baseData))
			})

			if !impl.AllowsDuplicates {

				It("should not add duplicates to the collection [no duplicates]", func() {
					Expect(col.Add(2, 3)).To(BeFalse())
					Expect(col.ToSlice()).To(Equal(baseData))
					Expect(col.Add(2, 3, -3)).To(BeTrue())
					Expect(col.ToSlice()).To(Equal(append(baseData, -3)))
				})

			}

		})

		Context("Comparing collections and elements", func() {

			It("should check if element is contained in collection", func() {
				Expect(col.Contains(4)).To(BeTrue())
				Expect(col.Contains(8)).To(BeFalse())
			})

			It("should check if multiple elements are contained in collection", func() {
				Expect(col.ContainsAll(col)).To(BeTrue())
				Expect(col.ContainsAll(collections.NewLinkedList(1, 2, 3, 4, -3))).To(BeFalse())
			})

			It("should compare two collections", func() {
				Expect(col.Equals(col)).To(BeTrue())
				Expect(col.Equals(collections.NewLinkedList(baseData...))).To(BeTrue())
				Expect(col.Equals(collections.NewLinkedList[int]())).To(BeFalse())
			})

		})

		Context("Size", func() {

			It("should return the size of the collection", func() {
				Expect(col.Size()).To(Equal(len(baseData)))
				Expect(impl.Constructor().Size()).To(BeZero())
			})

			It("should check if the collection is empty", func() {
				Expect(col.IsEmpty()).To(BeFalse())
				Expect(impl.Constructor().IsEmpty()).To(BeTrue())
			})

		})

		Context("Removing elements", func() {

			It("should clear the collection", func() {
				col.Clear()
				Expect(col.Size()).To(BeZero())
				Expect(col.IsEmpty()).To(BeTrue())
				Expect(col.ToSlice()).To(Equal([]int{}))
			})

			It("should do nothing if the to-be-removed element does not exist", func() {
				Expect(col.Remove(-3)).To(BeFalse())
				Expect(col.ToSlice()).To(Equal(baseData))
			})

			It("should remove a single element", func() {
				oldSize := col.Size()
				Expect(col.Remove(3)).To(BeTrue())
				Expect(col.Size()).To(Equal(oldSize - 1))
				Expect(col.ToSlice()).To(Equal([]int{1, 2, 4}))
			})

			It("should remove all elements of another collection", func() {
				col2 := impl.Constructor(2, 3)
				Expect(col.RemoveAll(col2)).To(BeTrue())
				Expect(col.ToSlice()).To(Equal([]int{1, 4}))
			})

			It("should retain all elements of another collection", func() {
				col2 := impl.Constructor(2, 3)
				Expect(col.RetainAll(col2)).To(BeTrue())
				Expect(col.ToSlice()).To(Equal([]int{3, 2}))
			})

			It("should remove all elements that match the given predicate", func() {
				col.Add(3)
				Expect(col.RemoveIf(func(i int) bool { return i > 2 })).To(BeTrue())
				Expect(col.ToSlice()).To(Equal([]int{1, 2}))
			})

			if impl.AllowsDuplicates {

				It("should remove a single element [duplicates]", func() {
					Expect(col.Add(3)).To(BeTrue())
					Expect(col.Remove(3)).To(BeTrue())
					Expect(col.ToSlice()).To(Equal([]int{1, 2, 4, 3}))
				})

				It("should remove all elements of another collection [duplicates]", func() {
					col2 := impl.Constructor(2, 3)
					Expect(col.Add(3)).To(BeTrue())
					Expect(col.RemoveAll(col2)).To(BeTrue())
					Expect(col.ToSlice()).To(Equal([]int{1, 4, 3}))
				})

				It("should remove all instances of the element [duplicates]", func() {
					Expect(col.Add(3)).To(BeTrue())
					Expect(col.RemoveAllOf(3)).To(BeTrue())
					Expect(col.ToSlice()).To(Equal([]int{1, 2, 4}))
				})

				It("should retain all elements of another collection [duplicates]", func() {
					col2 := impl.Constructor(2, 3)
					Expect(col.Add(3)).To(BeTrue())
					Expect(col.RetainAll(col2)).To(BeTrue())
					Expect(col.ToSlice()).To(Equal([]int{3, 2, 3}))
				})

			}

		})

		Context("New", func() {

			It("should return a new collection of the same type", func() {
				n := col.New()
				Expect(reflect.TypeOf(n)).To(Equal(reflect.TypeOf(col)))
			})

		})

	})

}
