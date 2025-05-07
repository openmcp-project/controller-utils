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

var _ = Describe("ConfigMapMutator", func() {
	var (
		ctx         context.Context
		fakeClient  client.WithWatch
		scheme      *runtime.Scheme
		data        map[string]string
		labels      map[string]string
		annotations map[string]string
		mutator     resources.Mutator[*v1.ConfigMap]
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

		// Define data, labels, and annotations
		data = map[string]string{"key1": "value1", "key2": "value2"}
		labels = map[string]string{"label1": "value1"}
		annotations = map[string]string{"annotation1": "value1"}

		// Create a ConfigMap mutator
		mutator = resources.NewConfigMapMutator("test-configmap", "test-namespace", data, labels, annotations)
	})

	It("should create an empty ConfigMap with correct metadata", func() {
		configMap := mutator.Empty()

		Expect(configMap.Name).To(Equal("test-configmap"))
		Expect(configMap.Namespace).To(Equal("test-namespace"))
		Expect(configMap.APIVersion).To(Equal("v1"))
		Expect(configMap.Kind).To(Equal("ConfigMap"))
	})

	It("should apply data, labels, and annotations using Mutate", func() {
		configMap := mutator.Empty()

		// Apply the mutator's Mutate method
		Expect(mutator.Mutate(configMap)).To(Succeed())

		// Verify that the data, labels, and annotations are applied
		Expect(configMap.Data).To(Equal(data))
		Expect(configMap.Labels).To(Equal(labels))
		Expect(configMap.Annotations).To(Equal(annotations))
	})

	It("should create and retrieve the ConfigMap using the fake client", func() {
		configMap := mutator.Empty()
		Expect(mutator.Mutate(configMap)).To(Succeed())

		// Create the ConfigMap in the fake client
		Expect(fakeClient.Create(ctx, configMap)).To(Succeed())

		// Retrieve the ConfigMap from the fake client and verify it
		retrievedConfigMap := &v1.ConfigMap{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "test-configmap", Namespace: "test-namespace"}, retrievedConfigMap)).To(Succeed())
		Expect(retrievedConfigMap).To(Equal(configMap))
	})
})
