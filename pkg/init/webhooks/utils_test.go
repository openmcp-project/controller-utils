package webhooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	testGVK = schema.GroupVersionKind{
		Group:   "example.com",
		Version: "v2",
		Kind:    "Project",
	}
)

func Test_generate_funcs(t *testing.T) {
	testCases := []struct {
		desc       string
		funcToTest func(gvk schema.GroupVersionKind) string
		expected   string
	}{
		{
			desc:       "generateMutatePath should return expected string",
			funcToTest: generateMutatePath,
			expected:   "/mutate-example-com-v2-project",
		},
		{
			desc:       "generateMutateName should return expected string",
			funcToTest: generateMutateName,
			expected:   "mutate-example-com-v2-project",
		},
		{
			desc:       "generateValidatePath should return expected string",
			funcToTest: generateValidatePath,
			expected:   "/validate-example-com-v2-project",
		},
		{
			desc:       "generateValidateName should return expected string",
			funcToTest: generateValidateName,
			expected:   "validate-example-com-v2-project",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			actual := tC.funcToTest(testGVK)
			assert.Equal(t, tC.expected, actual)
		})
	}
}
