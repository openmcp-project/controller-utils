package resources_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmcp-project/controller-utils/pkg/resources"
	"github.com/openmcp-project/controller-utils/pkg/testing"
)

var _ = Describe("NamespaceMutator", func() {
	var (
		ctx         context.Context
		fakeClient  client.WithWatch
		scheme      *runtime.Scheme
		labels      map[string]string
		annotations map[string]string
		mutator     resources.Mutator[*v1.Namespace]
	)

	BeforeEach(func() {
		ctx = context.TODO()

		// Create a scheme and register the core/v1 API
		scheme = runtime.NewScheme()
		Expect(v1.AddToScheme(scheme)).To(Succeed())

		// Initialize the fake client
		var err error
		fakeClient, err = testing.GetFakeClient(scheme)
		Expect(err).ToNot(HaveOccurred())

		// Define labels and annotations
		labels = map[string]string{"key1": "value1"}
		annotations = map[string]string{"annotation1": "value1"}

		// Create a namespace mutator
		mutator = resources.NewNamespaceMutator("test-namespace", labels, annotations)
	})

	It("should create an empty namespace with correct metadata", func() {
		namespace := mutator.Empty()

		Expect(namespace.Name).To(Equal("test-namespace"))
		Expect(namespace.APIVersion).To(Equal("v1"))
		Expect(namespace.Kind).To(Equal("Namespace"))
	})

	It("should apply labels and annotations using Mutate", func() {
		namespace := mutator.Empty()

		// Apply the mutator's Mutate method
		Expect(mutator.Mutate(namespace)).To(Succeed())

		// Verify that the labels and annotations are applied
		Expect(namespace.Labels).To(Equal(labels))
		Expect(namespace.Annotations).To(Equal(annotations))

		// Add additional labels and annotations
		additionalLabels := map[string]string{"key2": "value2"}
		additionalAnnotations := map[string]string{"annotation2": "value2"}
		namespace.SetLabels(additionalLabels)
		namespace.SetAnnotations(additionalAnnotations)

		// Apply the mutator's Mutate method again
		Expect(mutator.Mutate(namespace)).To(Succeed())
		// Verify that the original and additional labels and annotations are merged
		Expect(namespace.Labels).To(Equal(map[string]string{"key1": "value1", "key2": "value2"}))
		Expect(namespace.Annotations).To(Equal(map[string]string{"annotation1": "value1", "annotation2": "value2"}))
	})

	It("should create and retrieve the namespace using the fake client", func() {
		namespace := mutator.Empty()
		Expect(mutator.Mutate(namespace)).To(Succeed())

		// Create the namespace in the fake client
		Expect(fakeClient.Create(ctx, namespace)).To(Succeed())

		// Retrieve the namespace from the fake client and verify it
		retrievedNamespace := &v1.Namespace{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "test-namespace"}, retrievedNamespace)).To(Succeed())
		Expect(retrievedNamespace).To(Equal(namespace))
	})
})
