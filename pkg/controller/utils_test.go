package controller

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/pairs"

	"k8s.io/apimachinery/pkg/util/validation"
)

var _ = Describe("Predicates", func() {

	Context("K8sNameHash", func() {

		testData := []pairs.Pair[*[]string, string]{
			{
				Key:   &[]string{"test1"},
				Value: "dnhq5gcrs4mzrzzsa6cujsllg3b5ahhn67fkgmrvtvxr3a2woaka",
			},
			{
				Key:   &[]string{"bla"},
				Value: "jxz4h5upzsb3e7u5ileqimnhesm7c6dvzanftg2wnsmitoljm4bq",
			},
			{
				Key:   &[]string{"some other test", "this is a very, very long string"},
				Value: "rjphpfjbmwn6qqydv6xhtmj3kxrlzepn2tpwy4okw2ypoc3nlffq",
			},
		}

		It("should generate the same hash for the same input value", func() {
			for _, p := range testData {
				for range 5 {
					res := K8sNameHash(*p.Key...)
					Expect(res).To(Equal(p.Value))
				}
			}
		})

		It("should generate different hashes for different input values", func() {
			res1 := K8sNameHash(*testData[0].Key...)
			res2 := K8sNameHash(*testData[1].Key...)
			res3 := K8sNameHash(*testData[2].Key...)
			Expect(res1).NotTo(Equal(res2))
			Expect(res1).NotTo(Equal(res3))
			Expect(res2).NotTo(Equal(res3))
		})

		It("should generate a valid DNS1123Subdomain", func() {
			for _, p := range testData {
				res := K8sNameHash(*p.Key...)
				errs := validation.IsDNS1123Subdomain(res)
				Expect(errs).To(BeEmpty(), fmt.Sprintf("value %q is invalid: %v", res, errs))
			}
		})

	})

})
