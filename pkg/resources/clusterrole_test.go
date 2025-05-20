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

var _ = Describe("ClusterRoleMutator", func() {
	var (
		ctx         context.Context
		fakeClient  client.WithWatch
		scheme      *runtime.Scheme
		rules       []v1.PolicyRule
		labels      map[string]string
		annotations map[string]string
		mutator     resources.Mutator[*v1.ClusterRole]
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

		// Create a cluster role mutator
		mutator = resources.NewClusterRoleMutator("test-clusterrole", rules)
		mutator.MetadataMutator().WithLabels(labels).WithAnnotations(annotations)
	})

	It("should create an empty cluster role with correct metadata", func() {
		clusterRole := mutator.Empty()

		Expect(clusterRole.Name).To(Equal("test-clusterrole"))
		Expect(clusterRole.APIVersion).To(Equal("rbac.authorization.k8s.io/v1"))
		Expect(clusterRole.Kind).To(Equal("ClusterRole"))
	})

	It("should apply rules using Mutate", func() {
		clusterRole := mutator.Empty()

		// Apply the mutator's Mutate method
		Expect(mutator.Mutate(clusterRole)).To(Succeed())

		// Verify that the rules are applied
		Expect(clusterRole.Rules).To(Equal(rules))
	})

	It("should create and retrieve the cluster role using the fake client", func() {
		clusterRole := mutator.Empty()
		Expect(mutator.Mutate(clusterRole)).To(Succeed())

		// Create the cluster role in the fake client
		Expect(fakeClient.Create(ctx, clusterRole)).To(Succeed())

		// Retrieve the cluster role from the fake client and verify it
		retrievedClusterRole := &v1.ClusterRole{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "test-clusterrole"}, retrievedClusterRole)).To(Succeed())
		Expect(retrievedClusterRole).To(Equal(clusterRole))
	})
})
