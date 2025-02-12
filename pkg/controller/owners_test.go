package controller_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	ctrlutils "github.tools.sap/CoLa/controller-utils/pkg/controller"
)

var _ = Describe("Owners", func() {

	Context("HasOwnerReference", func() {

		It("should correctly identify the owner id", func() {
			owner1 := &corev1.Secret{}
			owner1.SetName("owner1")
			owner1.SetUID(types.UID("owner1-uid"))
			owner2 := &corev1.Secret{}
			owner2.SetName("owner2")
			owner2.SetUID(types.UID("owner2-uid"))
			nonOwner := &corev1.Secret{}
			nonOwner.SetName("non-owner")
			nonOwner.SetUID(types.UID("non-owner-uid"))
			owned := &corev1.Secret{}
			owned.SetName("owned")
			owned.SetUID(types.UID("owned-uid"))
			owned.SetOwnerReferences([]metav1.OwnerReference{
				{
					APIVersion: "v1",
					Kind:       "Secret",
					Name:       owner1.Name,
					UID:        owner1.UID,
				},
				{
					APIVersion: "v1",
					Kind:       "Secret",
					Name:       owner2.Name,
					UID:        owner2.UID,
				},
			})

			sc := runtime.NewScheme()
			Expect(clientgoscheme.AddToScheme(sc)).To(Succeed())

			idx, err := ctrlutils.HasOwnerReference(owned, owner1, sc)
			Expect(err).ToNot(HaveOccurred())
			Expect(idx).To(Equal(0))

			idx, err = ctrlutils.HasOwnerReference(owned, owner2, sc)
			Expect(err).ToNot(HaveOccurred())
			Expect(idx).To(Equal(1))

			idx, err = ctrlutils.HasOwnerReference(owned, nonOwner, sc)
			Expect(err).ToNot(HaveOccurred())
			Expect(idx).To(Equal(-1))

			// test with nil scheme
			owner1.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))
			owner2.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))
			nonOwner.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))

			idx, err = ctrlutils.HasOwnerReference(owned, owner1, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(idx).To(Equal(0))

			idx, err = ctrlutils.HasOwnerReference(owned, owner2, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(idx).To(Equal(1))

			idx, err = ctrlutils.HasOwnerReference(owned, nonOwner, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(idx).To(Equal(-1))
		})

	})

})
