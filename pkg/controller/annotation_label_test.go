package controller_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrlutils "github.com/openmcp-project/controller-utils/pkg/controller"
	testutils "github.com/openmcp-project/controller-utils/pkg/testing"
)

var _ = Describe("Annotation/Label Library", func() {

	Context("Annotations", func() {

		Context("IsMetadataEntryAlreadyExistsError", func() {

			It("should return true if the error is of type IsMetadataEntryAlreadyExistsError", func() {
				err := ctrlutils.NewMetadataEntryAlreadyExistsError(ctrlutils.ANNOTATION, "test-annotation", "desired-value", "actual-value")
				Expect(ctrlutils.IsMetadataEntryAlreadyExistsError(err)).To(BeTrue())
			})

			It("should return false if the error is not of type IsMetadataEntryAlreadyExistsError", func() {
				err := fmt.Errorf("test-error")
				Expect(ctrlutils.IsMetadataEntryAlreadyExistsError(err)).To(BeFalse())
			})

		})

		Context("Fetch", func() {

			Context("GetAnnotation", func() {

				It("should return the value of the annotation, if it exists", func() {
					env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
					ns := &corev1.Namespace{}
					Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-annotation"}, ns)).To(Succeed())
					value, ok := ctrlutils.GetAnnotation(ns, "foo.bar.baz/foo")
					Expect(ok).To(BeTrue())
					Expect(value).To(Equal(ns.GetAnnotations()["foo.bar.baz/foo"]))
				})

				It("should return an empty string and false if the annotation does not exist", func() {
					env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
					ns := &corev1.Namespace{}
					Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "no-annotation"}, ns)).To(Succeed())
					value, ok := ctrlutils.GetAnnotation(ns, "foo.bar.baz/foo")
					Expect(ok).To(BeFalse())
					Expect(value).To(Equal(""))
				})

			})

			Context("HasAnnotation", func() {

				It("should return true if the annotation exists", func() {
					env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
					ns := &corev1.Namespace{}
					Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-annotation"}, ns)).To(Succeed())
					Expect(ctrlutils.HasAnnotation(ns, "foo.bar.baz/foo")).To(BeTrue())
				})

				It("should return false if the annotation does not exist", func() {
					env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
					ns := &corev1.Namespace{}
					Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "no-annotation"}, ns)).To(Succeed())
					Expect(ctrlutils.HasAnnotation(ns, "foo.bar.baz/foo")).To(BeFalse())
				})

			})

			Context("HasAnnotationWithValue", func() {

				It("should return true if the annotation exists and has the expected value", func() {
					env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
					ns := &corev1.Namespace{}
					Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-annotation"}, ns)).To(Succeed())
					Expect(ctrlutils.HasAnnotationWithValue(ns, "foo.bar.baz/foo", "bar")).To(BeTrue())
				})

				It("should return false if the annotation exists but does not have the expected value", func() {
					env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
					ns := &corev1.Namespace{}
					Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-annotation"}, ns)).To(Succeed())
					Expect(ctrlutils.HasAnnotationWithValue(ns, "foo.bar.baz/foo", "asdf")).To(BeFalse())
				})

				It("should return false if the annotation does not exist", func() {
					env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
					ns := &corev1.Namespace{}
					Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "no-annotation"}, ns)).To(Succeed())
					Expect(ctrlutils.HasAnnotationWithValue(ns, "foo.bar.baz/foo", "asdf")).To(BeFalse())
				})

			})

		})

		Context("Modify in memory only", func() {

			It("should patch the annotation on the object, if it does not exist", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "no-annotation"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureAnnotation(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "bar", false)).To(Succeed())
				Expect(ns.GetAnnotations()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
				Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
				Expect(ns.GetAnnotations()).ToNot(HaveKey("foo.bar.baz/foo"))
			})

			It("should not fail if the annotation already exists with the desired value", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-annotation"}, ns)).To(Succeed())
				oldNs := ns.DeepCopy()
				Expect(ctrlutils.EnsureAnnotation(env.Ctx, nil, ns, "foo.bar.baz/foo", "bar", false)).To(Succeed())
				Expect(ns.GetAnnotations()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
				Expect(ns).To(Equal(oldNs))
				// client is nil, so trying to patch in cluster will cause a panic
			})

			It("should return a MetadataEntryAlreadyExistsError if the annotation already exists with a different value", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-annotation"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureAnnotation(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "baz", false)).To(MatchError(ctrlutils.NewMetadataEntryAlreadyExistsError(ctrlutils.ANNOTATION, "foo.bar.baz/foo", "baz", "bar")))
			})

			It("should overwrite the annotation if the mode is set to OVERWRITE", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-annotation"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureAnnotation(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "baz", false, ctrlutils.OVERWRITE)).To(Succeed())
				Expect(ns.GetAnnotations()).To(HaveKeyWithValue("foo.bar.baz/foo", "baz"))
				Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
				Expect(ns.GetAnnotations()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
			})

			It("should delete the annotation if the mode is set to DELETE", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-annotation"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureAnnotation(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "", false, ctrlutils.DELETE)).To(Succeed())
				Expect(ns.GetAnnotations()).NotTo(HaveKey("foo.bar.baz/foo"))
				Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
				Expect(ns.GetAnnotations()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
			})

		})

		Context("Modify in memory and in cluster", func() {

			It("should patch the annotation on the object, if it does not exist", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "no-annotation"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureAnnotation(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "bar", true)).To(Succeed())
				Expect(ns.GetAnnotations()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
				Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
				Expect(ns.GetAnnotations()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
			})

			It("should not fail if the annotation already exists with the desired value", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-annotation"}, ns)).To(Succeed())
				oldNs := ns.DeepCopy()
				Expect(ctrlutils.EnsureAnnotation(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "bar", true)).To(Succeed())
				Expect(ns.GetAnnotations()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
				Expect(ns).To(Equal(oldNs))
				Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
				Expect(ns.GetAnnotations()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
				Expect(ns).To(Equal(oldNs))
			})

			It("should return a MetadataEntryAlreadyExistsError if the annotation already exists with a different value", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-annotation"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureAnnotation(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "baz", true)).To(MatchError(ctrlutils.NewMetadataEntryAlreadyExistsError(ctrlutils.ANNOTATION, "foo.bar.baz/foo", "baz", "bar")))
			})

			It("should overwrite the annotation if the mode is set to OVERWRITE", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-annotation"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureAnnotation(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "baz", true, ctrlutils.OVERWRITE)).To(Succeed())
				Expect(ns.GetAnnotations()).To(HaveKeyWithValue("foo.bar.baz/foo", "baz"))
				Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
				Expect(ns.GetAnnotations()).To(HaveKeyWithValue("foo.bar.baz/foo", "baz"))
			})

			It("should delete the annotation if the mode is set to DELETE", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-annotation"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureAnnotation(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "", true, ctrlutils.DELETE)).To(Succeed())
				Expect(ns.GetAnnotations()).NotTo(HaveKey("foo.bar.baz/foo"))
				Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
				Expect(ns.GetAnnotations()).NotTo(HaveKey("foo.bar.baz/foo"))
			})

		})

	})

	Context("Labels", func() {

		Context("IsMetadataEntryAlreadyExistsError", func() {

			It("should return true if the error is of type IsMetadataEntryAlreadyExistsError", func() {
				err := ctrlutils.NewMetadataEntryAlreadyExistsError(ctrlutils.LABEL, "test-label", "desired-value", "actual-value")
				Expect(ctrlutils.IsMetadataEntryAlreadyExistsError(err)).To(BeTrue())
			})

			It("should return false if the error is not of type IsMetadataEntryAlreadyExistsError", func() {
				err := fmt.Errorf("test-error")
				Expect(ctrlutils.IsMetadataEntryAlreadyExistsError(err)).To(BeFalse())
			})

		})

		Context("Fetch", func() {

			Context("GetLabel", func() {

				It("should return the value of the label, if it exists", func() {
					env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
					ns := &corev1.Namespace{}
					Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-label"}, ns)).To(Succeed())
					value, ok := ctrlutils.GetLabel(ns, "foo.bar.baz/foo")
					Expect(ok).To(BeTrue())
					Expect(value).To(Equal(ns.GetLabels()["foo.bar.baz/foo"]))
				})

				It("should return an empty string and false if the label does not exist", func() {
					env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
					ns := &corev1.Namespace{}
					Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "no-label"}, ns)).To(Succeed())
					value, ok := ctrlutils.GetLabel(ns, "foo.bar.baz/foo")
					Expect(ok).To(BeFalse())
					Expect(value).To(Equal(""))
				})

			})

			Context("HasLabel", func() {

				It("should return true if the label exists", func() {
					env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
					ns := &corev1.Namespace{}
					Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-label"}, ns)).To(Succeed())
					Expect(ctrlutils.HasLabel(ns, "foo.bar.baz/foo")).To(BeTrue())
				})

				It("should return false if the label does not exist", func() {
					env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
					ns := &corev1.Namespace{}
					Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "no-label"}, ns)).To(Succeed())
					Expect(ctrlutils.HasLabel(ns, "foo.bar.baz/foo")).To(BeFalse())
				})

			})

			Context("HasLabelWithValue", func() {

				It("should return true if the label exists and has the expected value", func() {
					env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
					ns := &corev1.Namespace{}
					Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-label"}, ns)).To(Succeed())
					Expect(ctrlutils.HasLabelWithValue(ns, "foo.bar.baz/foo", "bar")).To(BeTrue())
				})

				It("should return false if the label exists but does not have the expected value", func() {
					env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
					ns := &corev1.Namespace{}
					Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-label"}, ns)).To(Succeed())
					Expect(ctrlutils.HasLabelWithValue(ns, "foo.bar.baz/foo", "asdf")).To(BeFalse())
				})

				It("should return false if the label does not exist", func() {
					env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
					ns := &corev1.Namespace{}
					Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "no-label"}, ns)).To(Succeed())
					Expect(ctrlutils.HasLabelWithValue(ns, "foo.bar.baz/foo", "asdf")).To(BeFalse())
				})

			})

		})

		Context("Modify in memory only", func() {

			It("should patch the label on the object, if it does not exist", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "no-label"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureLabel(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "bar", false)).To(Succeed())
				Expect(ns.GetLabels()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
				Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
				Expect(ns.GetLabels()).ToNot(HaveKey("foo.bar.baz/foo"))
			})

			It("should not fail if the label already exists with the desired value", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-label"}, ns)).To(Succeed())
				oldNs := ns.DeepCopy()
				Expect(ctrlutils.EnsureLabel(env.Ctx, nil, ns, "foo.bar.baz/foo", "bar", false)).To(Succeed())
				Expect(ns.GetLabels()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
				Expect(ns).To(Equal(oldNs))
				// client is nil, so trying to patch in cluster will cause a panic
			})

			It("should return a MetadataEntryAlreadyExistsError if the label already exists with a different value", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-label"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureLabel(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "baz", false)).To(MatchError(ctrlutils.NewMetadataEntryAlreadyExistsError(ctrlutils.LABEL, "foo.bar.baz/foo", "baz", "bar")))
			})

			It("should overwrite the label if the mode is set to OVERWRITE", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-label"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureLabel(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "baz", false, ctrlutils.OVERWRITE)).To(Succeed())
				Expect(ns.GetLabels()).To(HaveKeyWithValue("foo.bar.baz/foo", "baz"))
				Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
				Expect(ns.GetLabels()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
			})

			It("should delete the label if the mode is set to DELETE", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-label"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureLabel(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "", false, ctrlutils.DELETE)).To(Succeed())
				Expect(ns.GetLabels()).NotTo(HaveKey("foo.bar.baz/foo"))
				Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
				Expect(ns.GetLabels()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
			})

		})

		Context("Modify in memory and in cluster", func() {

			It("should patch the label on the object, if it does not exist", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "no-label"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureLabel(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "bar", true)).To(Succeed())
				Expect(ns.GetLabels()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
				Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
				Expect(ns.GetLabels()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
			})

			It("should not fail if the label already exists with the desired value", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-label"}, ns)).To(Succeed())
				oldNs := ns.DeepCopy()
				Expect(ctrlutils.EnsureLabel(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "bar", true)).To(Succeed())
				Expect(ns.GetLabels()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
				Expect(ns).To(Equal(oldNs))
				Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
				Expect(ns.GetLabels()).To(HaveKeyWithValue("foo.bar.baz/foo", "bar"))
				Expect(ns).To(Equal(oldNs))
			})

			It("should return a MetadataEntryAlreadyExistsError if the label already exists with a different value", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-label"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureLabel(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "baz", true)).To(MatchError(ctrlutils.NewMetadataEntryAlreadyExistsError(ctrlutils.LABEL, "foo.bar.baz/foo", "baz", "bar")))
			})

			It("should overwrite the label if the mode is set to OVERWRITE", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-label"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureLabel(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "baz", true, ctrlutils.OVERWRITE)).To(Succeed())
				Expect(ns.GetLabels()).To(HaveKeyWithValue("foo.bar.baz/foo", "baz"))
				Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
				Expect(ns.GetLabels()).To(HaveKeyWithValue("foo.bar.baz/foo", "baz"))
			})

			It("should delete the label if the mode is set to DELETE", func() {
				env := testutils.NewEnvironmentBuilder().WithInitObjectPath("testdata", "test-01").Build()
				ns := &corev1.Namespace{}
				Expect(env.Client().Get(env.Ctx, client.ObjectKey{Name: "foo-label"}, ns)).To(Succeed())
				Expect(ctrlutils.EnsureLabel(env.Ctx, env.Client(), ns, "foo.bar.baz/foo", "", true, ctrlutils.DELETE)).To(Succeed())
				Expect(ns.GetLabels()).NotTo(HaveKey("foo.bar.baz/foo"))
				Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
				Expect(ns.GetLabels()).NotTo(HaveKey("foo.bar.baz/foo"))
			})

		})

	})

})
