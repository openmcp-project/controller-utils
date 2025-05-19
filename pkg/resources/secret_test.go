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

var _ = Describe("SecretMutator", func() {
	var (
		ctx         context.Context
		fakeClient  client.WithWatch
		scheme      *runtime.Scheme
		data        map[string][]byte
		stringData  map[string]string
		labels      map[string]string
		annotations map[string]string
		secretType  core.SecretType
		mutator     resources.Mutator[*core.Secret]
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

		// Define data, labels, annotations, and secret type
		data = map[string][]byte{"key1": []byte("value1"), "key2": []byte("value2")}
		stringData = map[string]string{"key3": "value3", "key4": "value4"}
		labels = map[string]string{"label1": "value1"}
		annotations = map[string]string{"annotation1": "value1"}
		secretType = core.SecretTypeOpaque
	})

	It("should create an empty Secret with correct metadata", func() {
		mutator = resources.NewSecretMutator("test-secret", "test-namespace", data, secretType, labels, annotations)
		secret := mutator.Empty()

		Expect(secret.Name).To(Equal("test-secret"))
		Expect(secret.Namespace).To(Equal("test-namespace"))
		Expect(secret.APIVersion).To(Equal("v1"))
		Expect(secret.Kind).To(Equal("Secret"))
		Expect(secret.Type).To(Equal(secretType))
		Expect(secret.Data).To(Equal(data))
		Expect(secret.StringData).To(BeEmpty())
	})

	It("should create an empty Secret with correct metadata (string data)", func() {
		mutator = resources.NewSecretMutatorWithStringData("test-secret", "test-namespace", stringData, secretType, labels, annotations)
		secret := mutator.Empty()

		Expect(secret.Name).To(Equal("test-secret"))
		Expect(secret.Namespace).To(Equal("test-namespace"))
		Expect(secret.APIVersion).To(Equal("v1"))
		Expect(secret.Kind).To(Equal("Secret"))
		Expect(secret.Type).To(Equal(secretType))
		Expect(secret.Data).To(BeEmpty())
		Expect(secret.StringData).To(Equal(stringData))
	})

	It("should apply data, labels, and annotations using Mutate", func() {
		mutator = resources.NewSecretMutator("test-secret", "test-namespace", data, secretType, labels, annotations)
		secret := mutator.Empty()

		// Apply the mutator's Mutate method
		Expect(mutator.Mutate(secret)).To(Succeed())

		// Verify that the data, labels, and annotations are applied
		Expect(secret.Data).To(Equal(data))
		Expect(secret.StringData).To(BeEmpty())
		Expect(secret.Labels).To(Equal(labels))
		Expect(secret.Annotations).To(Equal(annotations))
	})

	It("should apply data, labels, and annotations using Mutate (string data)", func() {
		mutator = resources.NewSecretMutatorWithStringData("test-secret", "test-namespace", stringData, secretType, labels, annotations)
		secret := mutator.Empty()

		// Apply the mutator's Mutate method
		Expect(mutator.Mutate(secret)).To(Succeed())

		// Verify that the data, labels, and annotations are applied
		Expect(secret.Data).To(BeEmpty())
		Expect(secret.StringData).To(Equal(stringData))
		Expect(secret.Labels).To(Equal(labels))
		Expect(secret.Annotations).To(Equal(annotations))
	})

	It("should create and retrieve the Secret using the fake client", func() {
		mutator = resources.NewSecretMutator("test-secret", "test-namespace", data, secretType, labels, annotations)
		secret := mutator.Empty()
		Expect(mutator.Mutate(secret)).To(Succeed())

		// Create the Secret in the fake client
		Expect(fakeClient.Create(ctx, secret)).To(Succeed())

		// Retrieve the Secret from the fake client and verify it
		retrievedSecret := &core.Secret{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "test-secret", Namespace: "test-namespace"}, retrievedSecret)).To(Succeed())
		Expect(retrievedSecret).To(Equal(secret))
	})
})
