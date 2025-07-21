package resources_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmcp-project/controller-utils/pkg/resources"
	"github.com/openmcp-project/controller-utils/pkg/testing"
)

var _ = Describe("ValidatingAdmissionPolicyBindingMutator", func() {
	var (
		ctx         context.Context
		fakeClient  client.WithWatch
		scheme      *runtime.Scheme
		labels      map[string]string
		annotations map[string]string
		mutator     resources.Mutator[*admissionv1.ValidatingAdmissionPolicyBinding]
	)

	BeforeEach(func() {
		ctx = context.TODO()

		// Create a scheme and register the admissionregistration/v1 API
		scheme = runtime.NewScheme()
		Expect(admissionv1.AddToScheme(scheme)).To(Succeed())

		// Initialize the fake client
		var err error
		fakeClient, err = testing.GetFakeClient(scheme)
		Expect(err).ToNot(HaveOccurred())

		// Define labels and annotations
		labels = map[string]string{"label1": "value1"}
		annotations = map[string]string{"annotation1": "value1"}

		// Create a ValidatingAdmissionPolicyBinding mutator
		mutator = resources.NewValidatingAdmissionPolicyBindingMutator("test-vapb", admissionv1.ValidatingAdmissionPolicyBindingSpec{
			PolicyName: "test-policy",
			ParamRef: &admissionv1.ParamRef{
				Name:      "test-param",
				Namespace: "foo",
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
				ParameterNotFoundAction: ptr.To(admissionv1.DenyAction),
			},
			MatchResources: &admissionv1.MatchResources{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
			},
			ValidationActions: []admissionv1.ValidationAction{
				"asdf",
			},
		})
		mutator.MetadataMutator().WithLabels(labels).WithAnnotations(annotations)
	})

	It("should create an empty ValidatingAdmissionPolicyBinding with correct metadata", func() {
		vapb := mutator.Empty()

		Expect(vapb.Name).To(Equal("test-vapb"))
		Expect(vapb.APIVersion).To(Equal("admissionregistration.k8s.io/v1"))
		Expect(vapb.Kind).To(Equal("ValidatingAdmissionPolicyBinding"))
	})

	It("should apply labels and annotations using Mutate", func() {
		vapb := mutator.Empty()

		// Apply the mutator's Mutate method
		Expect(mutator.Mutate(vapb)).To(Succeed())

		// Verify that the labels and annotations are applied
		Expect(vapb.Labels).To(Equal(labels))
		Expect(vapb.Annotations).To(Equal(annotations))
	})

	It("should create and retrieve the ValidatingAdmissionPolicyBinding using the fake client", func() {
		vapb := mutator.Empty()
		Expect(mutator.Mutate(vapb)).To(Succeed())

		// Create the ValidatingAdmissionPolicyBinding in the fake client
		Expect(fakeClient.Create(ctx, vapb)).To(Succeed())

		// Retrieve the ValidatingAdmissionPolicyBinding from the fake client and verify it
		retrievedValidatingAdmissionPolicyBinding := &admissionv1.ValidatingAdmissionPolicyBinding{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "test-vapb"}, retrievedValidatingAdmissionPolicyBinding)).To(Succeed())
		Expect(retrievedValidatingAdmissionPolicyBinding).To(Equal(vapb))
	})
})
