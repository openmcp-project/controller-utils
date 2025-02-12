package collections_test

import (
	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.tools.sap/CoLa/controller-utils/pkg/collections"
)

var baseDataJSON []byte

type listImplementation struct {
	Name             string
	Constructor      func(elems ...int) collections.List[int]
	AllowsDuplicates bool
}

var _ = Describe("List Tests", func() {

	var err error
	baseDataJSON, err = json.Marshal(baseData)
	Expect(err).ToNot(HaveOccurred(), "unable to create baseDataJSON")

	for _, impl := range []*listImplementation{
		{
			Name:             "LinkedList",
			Constructor:      func(elems ...int) collections.List[int] { return collections.NewLinkedList[int](elems...) },
			AllowsDuplicates: true,
		},
	} {
		runListTests(impl)
	}

})

func runListTests(impl *listImplementation) {
	var li collections.List[int]

	BeforeEach(func() {
		li = impl.Constructor(baseData...)
	})

	AfterEach(func() {
		Expect(baseData).To(Equal([]int{1, 3, 2, 4}), "array passed into constructor should not be modified")
	})

	Context(impl.Name, func() {

		Context("List-specific functionality", func() {

			It("should add an element at a specific index", func() {
				Expect(li.AddIndex(0, 0)).To(Succeed())
				Expect(li.ToSlice()).To(Equal([]int{0, 1, 3, 2, 4}))
				Expect(li.AddIndex(5, li.Size())).To(Succeed())
				Expect(li.ToSlice()).To(Equal([]int{0, 1, 3, 2, 4, 5}))
				Expect(li.AddIndex(8, 3)).To(Succeed())
				Expect(li.ToSlice()).To(Equal([]int{0, 1, 3, 8, 2, 4, 5}))
				Expect(li.AddIndex(0, -1)).ToNot(Succeed())
				Expect(li.ToSlice()).To(Equal([]int{0, 1, 3, 8, 2, 4, 5}))
				Expect(li.AddIndex(0, li.Size()+1)).ToNot(Succeed())
				Expect(li.ToSlice()).To(Equal([]int{0, 1, 3, 8, 2, 4, 5}))
			})

			It("should remove an element from a specific index", func() {
				Expect(li.RemoveIndex(0)).To(Succeed())
				Expect(li.ToSlice()).To(Equal([]int{3, 2, 4}))
				Expect(li.RemoveIndex(1)).To(Succeed())
				Expect(li.ToSlice()).To(Equal([]int{3, 4}))
				Expect(li.RemoveIndex(1)).To(Succeed())
				Expect(li.ToSlice()).To(Equal([]int{3}))
				Expect(li.RemoveIndex(-1)).ToNot(Succeed())
				Expect(li.ToSlice()).To(Equal([]int{3}))
				Expect(li.RemoveIndex(li.Size())).ToNot(Succeed())
				Expect(li.ToSlice()).To(Equal([]int{3}))
			})

			It("should get the element at a specific index", func() {
				elem, err := li.Get(0)
				Expect(err).ToNot(HaveOccurred())
				Expect(elem).To(Equal(1))
				elem, err = li.Get(2)
				Expect(err).ToNot(HaveOccurred())
				Expect(elem).To(Equal(2))
				elem, err = li.Get(li.Size() - 1)
				Expect(err).ToNot(HaveOccurred())
				Expect(elem).To(Equal(baseData[len(baseData)-1]))
				_, err = li.Get(-1)
				Expect(err).To(HaveOccurred())
				_, err = li.Get(li.Size())
				Expect(err).To(HaveOccurred())
			})

		})

		Context("JSON conversion", func() {

			It("should marshal a list into a JSON list", func() {
				jsonContent, err := json.Marshal(li)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes.Equal(jsonContent, baseDataJSON)).To(BeTrue())
			})

			It("should unmarshal a JSON list in a list", func() {
				li2 := impl.Constructor(8, 4, 0, 10)
				Expect(json.Unmarshal(baseDataJSON, li2)).To(Succeed())
				Expect(li2.ToSlice()).To(Equal(baseData))
			})

		})

	})

}
