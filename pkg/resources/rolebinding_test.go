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

var _ = Describe("RoleBindingMutator", func() {
	var (
		ctx         context.Context
		fakeClient  client.WithWatch
		scheme      *runtime.Scheme
		subjects    []v1.Subject
		roleRef     v1.RoleRef
		labels      map[string]string
		annotations map[string]string
		mutator     resources.Mutator[*v1.RoleBinding]
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

		// Define subjects, roleRef, labels, and annotations
		subjects = []v1.Subject{
			{
				Kind:      "User",
				Name:      "test-user",
				Namespace: "test-namespace",
			},
		}
		roleRef = resources.NewRoleRef("test-role")
		labels = map[string]string{"key1": "value1"}
		annotations = map[string]string{"annotation1": "value1"}

		// Create a role binding mutator
		mutator = resources.NewRoleBindingMutator("test-rolebinding", "test-namespace", subjects, roleRef)
		mutator.MetadataMutator().WithLabels(labels).WithAnnotations(annotations)
	})

	It("should create an empty role binding with correct metadata", func() {
		roleBinding := mutator.Empty()

		Expect(roleBinding.Name).To(Equal("test-rolebinding"))
		Expect(roleBinding.Namespace).To(Equal("test-namespace"))
		Expect(roleBinding.APIVersion).To(Equal("rbac.authorization.k8s.io/v1"))
		Expect(roleBinding.Kind).To(Equal("RoleBinding"))
	})

	It("should apply subjects and roleRef using Mutate", func() {
		roleBinding := mutator.Empty()

		// Apply the mutator's Mutate method
		Expect(mutator.Mutate(roleBinding)).To(Succeed())

		// Verify that the subjects and roleRef are applied
		Expect(roleBinding.Subjects).To(Equal(subjects))
		Expect(roleBinding.RoleRef).To(Equal(roleRef))
	})

	It("should create and retrieve the role binding using the fake client", func() {
		roleBinding := mutator.Empty()
		Expect(mutator.Mutate(roleBinding)).To(Succeed())

		// Create the role binding in the fake client
		Expect(fakeClient.Create(ctx, roleBinding)).To(Succeed())

		// Retrieve the role binding from the fake client and verify it
		retrievedRoleBinding := &v1.RoleBinding{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "test-rolebinding", Namespace: "test-namespace"}, retrievedRoleBinding)).To(Succeed())
		Expect(retrievedRoleBinding).To(Equal(roleBinding))
	})
})
