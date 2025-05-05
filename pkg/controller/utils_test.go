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
			"wrckybtbh7enmn4vx2nnbpvpkuarsnvm",
		},
		{
			// check that the same string produces the same hash
			[]string{"test1"},
			"wrckybtbh7enmn4vx2nnbpvpkuarsnvm",
		},
		{
			[]string{"bla"},
			"76tha37scj5hjglta4tvn6b4kmxeh3ic",
		},
		{
			[]string{"some other test", "this is a very, very long string"},
			"fkkzqgh27xym6tqbswyql3wy4atsf6pt",
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
