package resources_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openmcp-project/controller-utils/pkg/resources"
)

var _ = Describe("Metadata Mutator", func() {

	It("should correctly add labels, annotations, owner references and finalizers", func() {
		ns := &corev1.Namespace{}
		ns.ObjectMeta = metav1.ObjectMeta{
			Name: "test-namespace",
			Labels: map[string]string{
				"oldLabel": "old",
			},
			Annotations: map[string]string{
				"oldAnnotation": "old",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "v1",
					Kind:       "Owner",
					Name:       "owner",
					UID:        "12345",
				},
			},
			Finalizers: []string{
				"oldFinalizer",
			},
		}

		m := resources.NewMetadataMutator().
			WithLabels(map[string]string{
				"newLabel": "new",
			}).
			WithAnnotations(map[string]string{
				"newAnnotation": "new",
			}).
			WithOwnerReferences([]metav1.OwnerReference{
				{
					APIVersion: "v1",
					Kind:       "NewOwner",
					Name:       "newOwner",
					UID:        "67890",
				},
			}).
			WithFinalizers([]string{"newFinalizer"})
		Expect(m.Mutate(ns)).To(Succeed())
		Expect(ns.Labels).To(HaveKeyWithValue("oldLabel", "old"))
		Expect(ns.Labels).To(HaveKeyWithValue("newLabel", "new"))
		Expect(ns.Annotations).To(HaveKeyWithValue("oldAnnotation", "old"))
		Expect(ns.Annotations).To(HaveKeyWithValue("newAnnotation", "new"))
		Expect(ns.OwnerReferences).To(ConsistOf(
			MatchFields(IgnoreExtras, Fields{
				"APIVersion": Equal("v1"),
				"Kind":       Equal("Owner"),
				"Name":       Equal("owner"),
				"UID":        BeEquivalentTo("12345"),
			}),
			MatchFields(IgnoreExtras, Fields{
				"APIVersion": Equal("v1"),
				"Kind":       Equal("NewOwner"),
				"Name":       Equal("newOwner"),
				"UID":        BeEquivalentTo("67890"),
			}),
		))
		Expect(ns.Finalizers).To(ConsistOf(
			"oldFinalizer",
			"newFinalizer",
		))
	})

})
