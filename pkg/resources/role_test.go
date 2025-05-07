package resources_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmcp-project/controller-utils/pkg/resources"
	"github.com/openmcp-project/controller-utils/pkg/testing"
)

var _ = Describe("RoleMutator", func() {
	var (
		ctx         context.Context
		fakeClient  client.WithWatch
		scheme      *runtime.Scheme
		rules       []v1.PolicyRule
		labels      map[string]string
		annotations map[string]string
		mutator     resources.Mutator[*v1.Role]
	)

	BeforeEach(func() {
		ctx = context.TODO()

		// Create a scheme and register the rbac/v1 API
		scheme = runtime.NewScheme()
		Expect(v1.AddToScheme(scheme)).To(Succeed())

		// Initialize the fake client
		var err error
		fakeClient, err = testing.GetFakeClient(scheme)
		Expect(err).ToNot(HaveOccurred())

		// Define rules, labels, and annotations
		rules = []v1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list"},
			},
		}
		labels = map[string]string{"key1": "value1"}
		annotations = map[string]string{"annotation1": "value1"}

		// Create a role mutator
		mutator = resources.NewRoleMutator("test-role", "test-namespace", rules, labels, annotations)
	})

	It("should create an empty role with correct metadata", func() {
		role := mutator.Empty()

		Expect(role.Name).To(Equal("test-role"))
		Expect(role.Namespace).To(Equal("test-namespace"))
		Expect(role.APIVersion).To(Equal("rbac.authorization.k8s.io/v1"))
		Expect(role.Kind).To(Equal("Role"))
	})

	It("should apply rules using Mutate", func() {
		role := mutator.Empty()

		// Apply the mutator's Mutate method
		Expect(mutator.Mutate(role)).To(Succeed())

		// Verify that the rules are applied
		Expect(role.Rules).To(Equal(rules))
	})

	It("should create and retrieve the role using the fake client", func() {
		role := mutator.Empty()
		Expect(mutator.Mutate(role)).To(Succeed())

		// Create the role in the fake client
		Expect(fakeClient.Create(ctx, role)).To(Succeed())

		// Retrieve the role from the fake client and verify it
		retrievedRole := &v1.Role{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "test-role", Namespace: "test-namespace"}, retrievedRole)).To(Succeed())
		Expect(retrievedRole).To(Equal(role))
	})
})
