package controller

import (
	"crypto/sha3"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Test_Version8UUID(t *testing.T) {
	testCases := []struct {
		desc        string
		data        []byte
		expectedErr *string
	}{
		{
			desc: "Text1",
			data: sha3.SumSHAKE128([]byte("hello world"), 16),
		},
		{
			desc: "Text2",
			data: sha3.SumSHAKE128([]byte("The quick brown fox jumps over the lazy dog"), 16),
		},
		{
			desc: "Text3",
			data: sha3.SumSHAKE128([]byte("Lorem ipsum dolor sit amet"), 16),
		},
		{
			desc:        "TooShort",
			data:        sha3.SumSHAKE128([]byte("Lorem ipsum dolor sit amet"), 15),
			expectedErr: ptr.To("invalid data (got 15 bytes)"),
		},
		{
			desc:        "TooLong",
			data:        sha3.SumSHAKE128([]byte("Lorem ipsum dolor sit amet"), 17),
			expectedErr: ptr.To("invalid data (got 17 bytes)"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			u, err := Version8UUID(tC.data)
			if tC.expectedErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, uuid.Version(8), u.Version(), "unexpected version")
				assert.Equal(t, uuid.RFC4122, u.Variant(), "unexpected variant")
			} else {
				assert.EqualError(t, err, *tC.expectedErr)
				assert.Equal(t, uuid.Nil, u)
			}
		})
	}
}

func Test_K8sNameUUID(t *testing.T) {
	testCases := []struct {
		desc         string
		input        []string
		expectedUUID string
		expectedErr  error
	}{
		{
			desc:         "should generate ID from one name",
			input:        []string{"example"},
			expectedUUID: "23cc2129-a257-8644-95e0-289d55c69704",
		},
		{
			desc:         "should generate ID from two names",
			input:        []string{corev1.NamespaceDefault, "example"},
			expectedUUID: "2bcf790e-815e-8ea9-857b-15be429583a5",
		},
		{
			desc:        "should fail because of empty slice",
			input:       []string{},
			expectedErr: ErrInvalidNames,
		},
		{
			desc:        "should fail because of slice with empty string",
			input:       []string{""},
			expectedErr: ErrInvalidNames,
		},
		{
			desc:        "should fail because of slice with empty strings",
			input:       []string{"", ""},
			expectedErr: ErrInvalidNames,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			actual, err := K8sNameUUID(tC.input...)
			assert.Equal(t, tC.expectedUUID, actual)
			assert.Equal(t, tC.expectedErr, err)
		})
	}
}

func Test_K8sObjectUUID(t *testing.T) {
	testCases := []struct {
		desc         string
		obj          client.Object
		expectedUUID string
	}{
		{
			desc: "should work with config map",
			obj: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			},
			expectedUUID: "2bcf790e-815e-8ea9-857b-15be429583a5", // same as in Test_K8sNameUUID
		},
		{
			desc: "should work with config map and empty namespace",
			obj: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example",
				},
			},
			expectedUUID: "2bcf790e-815e-8ea9-857b-15be429583a5", // same as above
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			actual, err := K8sObjectUUID(tC.obj)
			assert.NoError(t, err)
			assert.Equal(t, tC.expectedUUID, actual)
		})
	}
}

func Test_NameHashSHAKE128Base32(t *testing.T) {
	testCases := []struct {
		input    []string
		expected string
	}{
		{
			input:    []string{"example"},
			expected: "epgccknc",
		},
		{
			input:    []string{corev1.NamespaceDefault, "example"},
			expected: "fphxsdub",
		},
	}
	for _, tC := range testCases {
		t.Run(strings.Join(tC.input, " "), func(t *testing.T) {
			actual := NameHashSHAKE128Base32(tC.input...)
			assert.Equal(t, tC.expected, actual)
		})
	}
}

func Test_ObjectHashSHAKE128Base32(t *testing.T) {
	testCases := []struct {
		desc     string
		obj      client.Object
		expected string
	}{
		{
			desc: "should work with config map",
			obj: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			},
			expected: "fphxsdub", // same as in Test_K8sNameUUID
		},
		{
			desc: "should work with config map and empty namespace",
			obj: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example",
				},
			},
			expected: "fphxsdub", // same as above
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			actual := ObjectHashSHAKE128Base32(tC.obj)
			assert.Equal(t, tC.expected, actual)
		})
	}
}

func Test_ShortenToXCharacters(t *testing.T) {
	testCases := []struct {
		desc        string
		input       string
		maxLen      int
		expected    string
		expectedErr error
	}{
		{
			desc:     "short string",
			input:    "short",
			expected: "short",
			maxLen:   100,
		},
		{
			desc:     "short string to trim",
			input:    "short1234567",
			expected: "s--j5gore3p",
			maxLen:   11,
		},
		{
			desc:        "maxLen too small",
			input:       "shorter1234",
			maxLen:      10,
			expectedErr: ErrMaxLenTooSmall,
		},
		{
			desc:     "long string",
			input:    "this-is-a-very-a-very-a-very-long-string-that-is-over-63-characters",
			expected: "this-is-a-very-a-very-a-very-long-string-that-is-over--6reoyp5o",
			maxLen:   63,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			actual, err := ShortenToXCharacters(tC.input, tC.maxLen)
			if tC.expectedErr == nil {
				assert.Equal(t, tC.expected, actual)
				assert.LessOrEqual(t, len(actual), tC.maxLen)
			} else {
				assert.Equal(t, tC.expectedErr, err)
			}
		})
	}
}
