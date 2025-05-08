package resources_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmcp-project/controller-utils/pkg/resources"
	"github.com/openmcp-project/controller-utils/pkg/testing"
)

var _ = Describe("ServiceAccountMutator", func() {
	var (
		ctx         context.Context
		fakeClient  client.WithWatch
		scheme      *runtime.Scheme
		labels      map[string]string
		annotations map[string]string
		mutator     resources.Mutator[*core.ServiceAccount]
	)

	BeforeEach(func() {
		ctx = context.TODO()

		// Create a scheme and register the core/v1 API
		scheme = runtime.NewScheme()
		Expect(core.AddToScheme(scheme)).To(Succeed())

		// Initialize the fake client
		var err error
		fakeClient, err = testing.GetFakeClient(scheme)
		Expect(err).ToNot(HaveOccurred())

		// Define labels and annotations
		labels = map[string]string{"label1": "value1"}
		annotations = map[string]string{"annotation1": "value1"}

		// Create a ServiceAccount mutator
		mutator = resources.NewServiceAccountMutator("test-serviceaccount", "test-namespace", labels, annotations)
	})

	It("should create an empty ServiceAccount with correct metadata", func() {
		serviceAccount := mutator.Empty()

		Expect(serviceAccount.Name).To(Equal("test-serviceaccount"))
		Expect(serviceAccount.Namespace).To(Equal("test-namespace"))
		Expect(serviceAccount.APIVersion).To(Equal("v1"))
		Expect(serviceAccount.Kind).To(Equal("ServiceAccount"))
	})

	It("should apply labels and annotations using Mutate", func() {
		serviceAccount := mutator.Empty()

		// Apply the mutator's Mutate method
		Expect(mutator.Mutate(serviceAccount)).To(Succeed())

		// Verify that the labels and annotations are applied
		Expect(serviceAccount.Labels).To(Equal(labels))
		Expect(serviceAccount.Annotations).To(Equal(annotations))
	})

	It("should create and retrieve the ServiceAccount using the fake client", func() {
		serviceAccount := mutator.Empty()
		Expect(mutator.Mutate(serviceAccount)).To(Succeed())

		// Create the ServiceAccount in the fake client
		Expect(fakeClient.Create(ctx, serviceAccount)).To(Succeed())

		// Retrieve the ServiceAccount from the fake client and verify it
		retrievedServiceAccount := &core.ServiceAccount{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "test-serviceaccount", Namespace: "test-namespace"}, retrievedServiceAccount)).To(Succeed())
		Expect(retrievedServiceAccount).To(Equal(serviceAccount))
	})
})
