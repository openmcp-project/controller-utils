package controller

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation"
)

func TestK8sNameHash(t *testing.T) {
	tt := []struct {
		input   []string
		expHash string
	}{
		{
			[]string{"test1"},
			"dnhq5gcrs4mzrzzsa6cujsllg3b5ahhn67fkgmrvtvxr3a2woaka",
		},
		{
			// check that the same string produces the same hash
			[]string{"test1"},
			"dnhq5gcrs4mzrzzsa6cujsllg3b5ahhn67fkgmrvtvxr3a2woaka",
		},
		{
			[]string{"bla"},
			"jxz4h5upzsb3e7u5ileqimnhesm7c6dvzanftg2wnsmitoljm4bq",
		},
		{
			[]string{"some other test", "this is a very, very long string"},
			"rjphpfjbmwn6qqydv6xhtmj3kxrlzepn2tpwy4okw2ypoc3nlffq",
		},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprint(tc.input), func(t *testing.T) {
			res := K8sNameHash(tc.input...)

			if res != tc.expHash {
				t.Errorf("exp hash %q, got %q", tc.expHash, res)
			}

			// ensure the result is a valid DNS1123Subdomain
			if errs := validation.IsDNS1123Subdomain(res); errs != nil {
				t.Errorf("value %q is invalid: %v", res, errs)
			}

		})
	}

}
