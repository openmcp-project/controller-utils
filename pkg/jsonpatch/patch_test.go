package jsonpatch_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	jpapi "github.com/openmcp-project/controller-utils/api/jsonpatch"
	"github.com/openmcp-project/controller-utils/pkg/jsonpatch"
)

const (
	docBase = `{"foo":"bar","baz":{"foobar":"asdf"},"abc":[{"a":1},{"b":2},{"c":3}]}`
)

var _ = Describe("JSONPatch", func() {

	var doc []byte

	BeforeEach(func() {
		doc = []byte(docBase)
	})

	Context("Untyped", func() {

		It("should not do anything if the patch is empty", func() {
			patch := jsonpatch.New(newPatches()...)
			result, err := patch.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(doc))
			Expect(doc).To(Equal([]byte(docBase)))
		})

		It("should apply a simple patch", func() {
			patch := jsonpatch.New(newPatches(newPatch(jpapi.ADD, "/foo", "baz", ""))...)
			result, err := patch.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]byte(`{"foo":"baz","baz":{"foobar":"asdf"},"abc":[{"a":1},{"b":2},{"c":3}]}`)))
			Expect(doc).To(Equal([]byte(docBase)))
		})

		It("should add an element to a list", func() {
			patch := jsonpatch.New(newPatches(newPatch(jpapi.ADD, "/abc/-1", map[string]any{"d": 4}, ""))...)
			result, err := patch.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]byte(`{"foo":"bar","baz":{"foobar":"asdf"},"abc":[{"a":1},{"b":2},{"c":3},{"d":4}]}`)))
			Expect(doc).To(Equal([]byte(docBase)))
		})

		It("should apply multiple patches in the correct order", func() {
			patch := jsonpatch.New(newPatches(
				newPatch(jpapi.ADD, "/foo", "baz", ""),
				newPatch(jpapi.COPY, "/baz/foobar", nil, "/foo"),
				newPatch(jpapi.REPLACE, "/abc/2/c", 6, ""),
				newPatch(jpapi.REMOVE, "/abc/1", nil, ""),
			)...)
			result, err := patch.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]byte(`{"foo":"baz","baz":{"foobar":"baz"},"abc":[{"a":1},{"c":6}]}`)))
			Expect(doc).To(Equal([]byte(docBase)))
		})

		It("should handle paths that need conversion correctly", func() {
			patch := jsonpatch.New(newPatches(
				newPatch(jpapi.ADD, ".foo", "baz", ""),
				newPatch(jpapi.COPY, "baz.foobar", nil, ".foo"),
				newPatch(jpapi.REPLACE, "abc[2].c", 6, ""),
				newPatch(jpapi.REMOVE, ".abc[1]", nil, ""),
			)...)
			result, err := patch.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]byte(`{"foo":"baz","baz":{"foobar":"baz"},"abc":[{"a":1},{"c":6}]}`)))
			Expect(doc).To(Equal([]byte(docBase)))
		})

		It("should apply options correctly", func() {
			patch := jsonpatch.New(newPatches()...)
			result, err := patch.Apply(doc, jsonpatch.Indent("  "))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(result)).To(Equal(`{
  "foo": "bar",
  "baz": {
    "foobar": "asdf"
  },
  "abc": [
    {
      "a": 1
    },
    {
      "b": 2
    },
    {
      "c": 3
    }
  ]
}`))
			Expect(doc).To(Equal([]byte(docBase)))

			patch = jsonpatch.New(newPatches(
				newPatch(jpapi.REPLACE, "/abc/-1", map[string]any{"d": 4}, ""),
			)...)
			result, err = patch.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]byte(`{"foo":"bar","baz":{"foobar":"asdf"},"abc":[{"a":1},{"b":2},{"d":4}]}`)))
			Expect(doc).To(Equal([]byte(docBase)))

			_, err = patch.Apply(doc, jsonpatch.SupportNegativeIndices(false))
			Expect(err).To(HaveOccurred())
		})

	})

	Context("Typed", func() {

		type abc struct {
			A int `json:"a,omitempty"`
			B int `json:"b,omitempty"`
			C int `json:"c,omitempty"`
		}

		type baz struct {
			Foobar string `json:"foobar"`
		}

		type testDoc struct {
			Foo string `json:"foo"`
			Baz baz    `json:"baz"`
			ABC []abc  `json:"abc"`
		}

		var typedDoc *testDoc
		var typedDocCompare *testDoc

		BeforeEach(func() {
			typedDoc = &testDoc{}
			err := json.Unmarshal([]byte(docBase), typedDoc)
			Expect(err).ToNot(HaveOccurred())
			typedDocCompare = &testDoc{}
			err = json.Unmarshal([]byte(docBase), typedDocCompare)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not do anything if the patch is empty", func() {
			patch := jsonpatch.NewTyped[*testDoc](newPatches()...)
			result, err := patch.Apply(typedDoc)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result).To(Equal(typedDoc))
			Expect(result == typedDoc).To(BeFalse(), "result should not be the same pointer as the input document")
			Expect(typedDoc).To(Equal(typedDocCompare))
		})

		It("should apply a simple patch", func() {
			patch := jsonpatch.NewTyped[*testDoc](newPatches(newPatch(jpapi.ADD, "/foo", "baz", ""))...)
			result, err := patch.Apply(typedDoc)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result).To(Equal(&testDoc{
				Foo: "baz",
				Baz: baz{Foobar: "asdf"},
				ABC: []abc{{A: 1}, {B: 2}, {C: 3}},
			}))
			Expect(result == typedDoc).To(BeFalse(), "result should not be the same pointer as the input document")
			Expect(typedDoc).To(Equal(typedDocCompare))
		})

		It("should apply multiple patches in the correct order", func() {
			patch := jsonpatch.NewTyped[*testDoc](newPatches(
				newPatch(jpapi.ADD, "/foo", "baz", ""),
				newPatch(jpapi.COPY, "/baz/foobar", nil, "/foo"),
				newPatch(jpapi.REPLACE, "/abc/2/c", 6, ""),
				newPatch(jpapi.REMOVE, "/abc/1", nil, ""),
			)...)
			result, err := patch.Apply(typedDoc)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result).To(Equal(&testDoc{
				Foo: "baz",
				Baz: baz{Foobar: "baz"},
				ABC: []abc{{A: 1}, {C: 6}},
			}))
			Expect(result == typedDoc).To(BeFalse(), "result should not be the same pointer as the input document")
			Expect(typedDoc).To(Equal(typedDocCompare))
		})

		It("should handle paths that need conversion correctly", func() {
			patch := jsonpatch.NewTyped[*testDoc](newPatches(
				newPatch(jpapi.ADD, ".foo", "baz", ""),
				newPatch(jpapi.COPY, "baz.foobar", nil, ".foo"),
				newPatch(jpapi.REPLACE, "abc[2].c", 6, ""),
				newPatch(jpapi.REMOVE, ".abc[1]", nil, ""),
			)...)
			result, err := patch.Apply(typedDoc)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result).To(Equal(&testDoc{
				Foo: "baz",
				Baz: baz{Foobar: "baz"},
				ABC: []abc{{A: 1}, {C: 6}},
			}))
			Expect(result == typedDoc).To(BeFalse(), "result should not be the same pointer as the input document")
			Expect(typedDoc).To(Equal(typedDocCompare))
		})

	})

	Context("API", func() {

		It("should be able to marshal and unmarshal JSONPatches", func() {
			rawAPIPatches := []byte(`[{"op":"add","path":"/foo","from":"/bar","value":{"foobar":"foobaz"}}]`)
			var apiPatches jpapi.JSONPatches
			err := json.Unmarshal(rawAPIPatches, &apiPatches)
			Expect(err).ToNot(HaveOccurred())
			Expect(apiPatches).To(ConsistOf(newPatch(jpapi.ADD, "/foo", map[string]any{"foobar": "foobaz"}, "/bar")))
			marshalled, err := json.Marshal(apiPatches)
			Expect(err).ToNot(HaveOccurred())
			Expect(marshalled).To(Equal(rawAPIPatches))
		})

	})

})

func newPatch(op, path string, value any, from string) jpapi.JSONPatch {
	var valueData *jsonpatch.PatchValueData
	if value != nil {
		valueJSON, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		valueData = &jsonpatch.PatchValueData{
			Raw: valueJSON,
		}
	}
	return jpapi.JSONPatch{
		Op:    op,
		Path:  path,
		Value: valueData,
		From:  from,
	}
}

func newPatches(patches ...jpapi.JSONPatch) jpapi.JSONPatches {
	return patches
}
