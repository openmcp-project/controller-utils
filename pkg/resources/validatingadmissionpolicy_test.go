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

var _ = Describe("ValidatingAdmissionPolicyMutator", func() {
	var (
		ctx         context.Context
		fakeClient  client.WithWatch
		scheme      *runtime.Scheme
		labels      map[string]string
		annotations map[string]string
		mutator     resources.Mutator[*admissionv1.ValidatingAdmissionPolicy]
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

		// Create a ValidatingAdmissionPolicy mutator
		mutator = resources.NewValidatingAdmissionPolicyMutator("test-vap", admissionv1.ValidatingAdmissionPolicySpec{
			ParamKind: &admissionv1.ParamKind{
				APIVersion: "v1",
				Kind:       "TestParam",
			},
			MatchConstraints: &admissionv1.MatchResources{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
			},
			Validations: []admissionv1.Validation{
				{
					Expression:        "asdf",
					Message:           "doesnotmatter",
					MessageExpression: "qwer",
				},
			},
			FailurePolicy: ptr.To(admissionv1.Fail),
			AuditAnnotations: []admissionv1.AuditAnnotation{
				{
					Key:             "example.com/audit",
					ValueExpression: "example.com/audit-value",
				},
			},
			MatchConditions: []admissionv1.MatchCondition{
				{
					Name:       "example-condition",
					Expression: "example.com/condition-expression",
				},
			},
			Variables: []admissionv1.Variable{
				{
					Name:       "example-variable",
					Expression: "example.com/variable-expression",
				},
			},
		})
		mutator.MetadataMutator().WithLabels(labels).WithAnnotations(annotations)
	})

	It("should create an empty ValidatingAdmissionPolicy with correct metadata", func() {
		vap := mutator.Empty()

		Expect(vap.Name).To(Equal("test-vap"))
		Expect(vap.APIVersion).To(Equal("admissionregistration.k8s.io/v1"))
		Expect(vap.Kind).To(Equal("ValidatingAdmissionPolicy"))
	})

	It("should apply labels and annotations using Mutate", func() {
		vap := mutator.Empty()

		// Apply the mutator's Mutate method
		Expect(mutator.Mutate(vap)).To(Succeed())

		// Verify that the labels and annotations are applied
		Expect(vap.Labels).To(Equal(labels))
		Expect(vap.Annotations).To(Equal(annotations))
	})

	It("should create and retrieve the ValidatingAdmissionPolicy using the fake client", func() {
		vap := mutator.Empty()
		Expect(mutator.Mutate(vap)).To(Succeed())

		// Create the ValidatingAdmissionPolicy in the fake client
		Expect(fakeClient.Create(ctx, vap)).To(Succeed())

		// Retrieve the ValidatingAdmissionPolicy from the fake client and verify it
		retrievedValidatingAdmissionPolicy := &admissionv1.ValidatingAdmissionPolicy{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: "test-vap"}, retrievedValidatingAdmissionPolicy)).To(Succeed())
		Expect(retrievedValidatingAdmissionPolicy).To(Equal(vap))
	})
})
