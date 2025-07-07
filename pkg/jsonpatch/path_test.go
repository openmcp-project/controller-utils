package jsonpatch_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/jsonpatch"
)

var _ = Describe("ConvertPath", func() {

	It("should convert simple paths correctly", func() {
		verifyPathConversion("", "/")
		verifyPathConversion("/", "/")
		verifyPathConversion("a", "/a")
		verifyPathConversion(".a", "/a")
		verifyPathConversion("a.b", "/a/b")
		verifyPathConversion(".a.b", "/a/b")
		verifyPathConversion("a[0]", "/a/0")
		verifyPathConversion("a[0].b", "/a/0/b")
		verifyPathConversion("a[b][c]", "/a/b/c")
		verifyPathConversion("a.b[c]", "/a/b/c")
		verifyPathConversion("a[b.c].d", "/a/b.c/d")
	})

	It("should convert paths with quotes or escapes correctly", func() {
		verifyPathConversion("a['b']", "/a/b")
		verifyPathConversion("a[\"b\"]", "/a/b")
		verifyPathConversion("a['b.c']", "/a/b.c")
		verifyPathConversion("a[\"b.c\"]", "/a/b.c")
		verifyPathConversion("a['b c']", "/a/b c")
		verifyPathConversion("a[\"b c\"]", "/a/b c")
		verifyPathConversion("a['b\\c']", "/a/b\\c")
		verifyPathConversion("a[\"b\\c\"]", "/a/b\\c")
		verifyPathConversion("a\\.b", "/a.b")
		verifyPathConversion("a\\[b\\]", "/a[b]")
		verifyPathConversion("a.\\'b\\'.c", "/a/'b'/c")
	})

	It("should handle paths with ~ and / characters correctly", func() {
		verifyPathConversion(".a~b.c/d", "/a~0b/c~1d")
	})

	It("should throw an error for invalid paths", func() {
		_, err := jsonpatch.ConvertPath("a[")
		Expect(err).To(MatchError(ContainSubstring("unexpected end of input after opening bracket")))

		_, err = jsonpatch.ConvertPath("a[foo")
		Expect(err).To(MatchError(ContainSubstring("unexpected end of input")))

		_, err = jsonpatch.ConvertPath("a['foo]")
		Expect(err).To(MatchError(ContainSubstring("unexpected end of input")))

		_, err = jsonpatch.ConvertPath("a[\"foo]")
		Expect(err).To(MatchError(ContainSubstring("unexpected end of input")))

		_, err = jsonpatch.ConvertPath("a]")
		Expect(err).To(MatchError(ContainSubstring("invalid character")))

		_, err = jsonpatch.ConvertPath("a\"")
		Expect(err).To(MatchError(ContainSubstring("invalid character")))

		_, err = jsonpatch.ConvertPath("a'")
		Expect(err).To(MatchError(ContainSubstring("invalid character")))

		_, err = jsonpatch.ConvertPath("a\\")
		Expect(err).To(MatchError(ContainSubstring("unexpected end of input")))
	})

})

func verifyPathConversion(input, expected string) {
	converted, err := jsonpatch.ConvertPath(input)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	ExpectWithOffset(1, converted).To(Equal(expected))
}
