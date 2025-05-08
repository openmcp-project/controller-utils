package resources_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmcp-project/controller-utils/pkg/resources"
	"github.com/openmcp-project/controller-utils/pkg/testing"
)

var _ = Describe("Resource Functions", func() {
	var (
		ctx         context.Context
		fakeClient  client.WithWatch
		scheme      *runtime.Scheme
		labels      map[string]string
		annotations map[string]string
		data        map[string]string
		mutator     resources.Mutator[*corev1.ConfigMap]
	)

	BeforeEach(func() {
		ctx = context.TODO()

		// Create a scheme and register the core/v1 API
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())

		// Initialize the fake client
		var err error
		fakeClient, err = testing.GetFakeClient(scheme)
		Expect(err).ToNot(HaveOccurred())

		// Define labels, annotations, and data
		labels = map[string]string{"label1": "value1"}
		annotations = map[string]string{"annotation1": "value1"}
		data = map[string]string{"key1": "value1", "key2": "value2"}

		// Create a ConfigMap mutator
		mutator = resources.NewConfigMapMutator("test-configmap", "test-namespace", data, labels, annotations)
	})

	It("should get, create or update, and delete a resource", func() {
		// Test CreateOrUpdateResource
		Expect(resources.CreateOrUpdateResource(ctx, fakeClient, mutator)).To(Succeed())

		// Test GetResource
		retrievedConfigMap, err := resources.GetResource(ctx, fakeClient, mutator)
		Expect(err).ToNot(HaveOccurred())
		Expect(retrievedConfigMap.Data).To(Equal(data))
		Expect(retrievedConfigMap.Labels).To(Equal(labels))
		Expect(retrievedConfigMap.Annotations).To(Equal(annotations))

		// Test DeleteResource
		Expect(resources.DeleteResource(ctx, fakeClient, mutator)).To(Succeed())

		// Verify the resource is deleted
		_, err = resources.GetResource(ctx, fakeClient, mutator)
		Expect(err).To(HaveOccurred())
	})
})
