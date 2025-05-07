package resources_test

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmcp-project/controller-utils/pkg/resources"
	"github.com/openmcp-project/controller-utils/pkg/testing"
)

var _ = Describe("CRDMutator", func() {
	var (
		ctx         context.Context
		fakeClient  client.WithWatch
		scheme      *runtime.Scheme
		labels      map[string]string
		annotations map[string]string
		crd         *apiextensionsv1.CustomResourceDefinition
		mutator     resources.Mutator[*apiextensionsv1.CustomResourceDefinition]
	)

	BeforeEach(func() {
		ctx = context.TODO()

		// Create a scheme and register the apiextensions/v1 API
		scheme = runtime.NewScheme()
		Expect(apiextensionsv1.AddToScheme(scheme)).To(Succeed())

		// Initialize the fake client
		var err error
		fakeClient, err = testing.GetFakeClient(scheme)
		Expect(err).ToNot(HaveOccurred())

		// Define labels and annotations
		labels = map[string]string{"label1": "value1"}
		annotations = map[string]string{"annotation1": "value1"}

		// Define a CRD object
		crd = &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "test-crd",
				Labels:      labels,
				Annotations: annotations,
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: "example.com",
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural:   "tests",
					Singular: "test",
					Kind:     "Test",
				},
				Scope: apiextensionsv1.NamespaceScoped,
			},
		}

		// Create a CRD mutator
		mutator = resources.NewCRDMutator(crd, labels, annotations)
	})

	It("should create an empty CRD with correct metadata", func() {
		crd := mutator.Empty()

		Expect(crd.Name).To(Equal("test-crd"))
		Expect(crd.APIVersion).To(Equal("apiextensions.k8s.io/v1"))
		Expect(crd.Kind).To(Equal("CustomResourceDefinition"))
	})

	It("should apply labels and annotations using Mutate", func() {
		crd := mutator.Empty()

		// Apply the mutator's Mutate method
		Expect(mutator.Mutate(crd)).To(Succeed())

		// Verify that the labels and annotations are applied
		Expect(crd.Labels).To(Equal(labels))
		Expect(crd.Annotations).To(Equal(annotations))
	})

	It("should create and retrieve the CRD using the fake client", func() {
		crd := mutator.Empty()
		Expect(mutator.Mutate(crd)).To(Succeed())

		// Create the CRD in the fake client
		Expect(fakeClient.Create(ctx, crd)).To(Succeed())

		// Retrieve the CRD from the fake client and verify it
		retrievedCRD := &apiextensionsv1.CustomResourceDefinition{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "test-crd"}, retrievedCRD)).To(Succeed())
		Expect(retrievedCRD).To(Equal(crd))
	})
})
