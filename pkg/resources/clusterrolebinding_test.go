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

var _ = Describe("ClusterRoleBindingMutator", func() {
	var (
		ctx         context.Context
		fakeClient  client.WithWatch
		scheme      *runtime.Scheme
		subjects    []v1.Subject
		roleRef     v1.RoleRef
		labels      map[string]string
		annotations map[string]string
		mutator     resources.Mutator[*v1.ClusterRoleBinding]
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
		roleRef = resources.NewClusterRoleRef("test-role")
		labels = map[string]string{"key1": "value1"}
		annotations = map[string]string{"annotation1": "value1"}

		// Create a cluster role binding mutator
		mutator = resources.NewClusterRoleBindingMutator("test-clusterrolebinding", subjects, roleRef)
		mutator.MetadataMutator().WithLabels(labels).WithAnnotations(annotations)
	})

	It("should create an empty cluster role binding with correct metadata", func() {
		clusterRoleBinding := mutator.Empty()

		Expect(clusterRoleBinding.Name).To(Equal("test-clusterrolebinding"))
		Expect(clusterRoleBinding.APIVersion).To(Equal("rbac.authorization.k8s.io/v1"))
		Expect(clusterRoleBinding.Kind).To(Equal("ClusterRoleBinding"))
	})

	It("should apply subjects and roleRef using Mutate", func() {
		clusterRoleBinding := mutator.Empty()

		// Apply the mutator's Mutate method
		Expect(mutator.Mutate(clusterRoleBinding)).To(Succeed())

		// Verify that the subjects and roleRef are applied
		Expect(clusterRoleBinding.Subjects).To(Equal(subjects))
		Expect(clusterRoleBinding.RoleRef).To(Equal(roleRef))
	})

	It("should create and retrieve the cluster role binding using the fake client", func() {
		clusterRoleBinding := mutator.Empty()
		Expect(mutator.Mutate(clusterRoleBinding)).To(Succeed())

		// Create the cluster role binding in the fake client
		Expect(fakeClient.Create(ctx, clusterRoleBinding)).To(Succeed())

		// Retrieve the cluster role binding from the fake client and verify it
		retrievedClusterRoleBinding := &v1.ClusterRoleBinding{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "test-clusterrolebinding"}, retrievedClusterRoleBinding)).To(Succeed())
		Expect(retrievedClusterRoleBinding).To(Equal(clusterRoleBinding))
	})
})
