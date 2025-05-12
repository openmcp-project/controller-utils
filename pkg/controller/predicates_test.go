package controller_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"

	ctrlutils "github.com/openmcp-project/controller-utils/pkg/controller"
)

var _ = Describe("Predicates", func() {

	var base *corev1.Service
	var changed *corev1.Service

	BeforeEach(func() {
		base = &corev1.Service{}
		base.SetName("foo")
		base.SetNamespace("bar")
		base.SetGeneration(1)
		changed = base.DeepCopy()
	})

	Context("DeletionTimestamp", func() {

		It("should detect changes to the deletion timestamp", func() {
			p := ctrlutils.DeletionTimestampChangedPredicate{}
			Expect(p.Update(updateEvent(base, changed))).To(BeFalse(), "DeletionTimestampChangedPredicate should return false if the deletion timestamp did not change")
			changed.SetDeletionTimestamp(ptr.To(metav1.Now()))
			Expect(p.Update(updateEvent(base, changed))).To(BeTrue(), "DeletionTimestampChangedPredicate should return true if the deletion timestamp did change")
		})

	})

	Context("Annotations", func() {

		It("should detect changes to the annotations", func() {
			pHasFoo := ctrlutils.HasAnnotationPredicate("foo", "")
			pHasBar := ctrlutils.HasAnnotationPredicate("bar", "")
			pHasFooWithFoo := ctrlutils.HasAnnotationPredicate("foo", "foo")
			pHasFooWithBar := ctrlutils.HasAnnotationPredicate("foo", "bar")
			pGotFoo := ctrlutils.GotAnnotationPredicate("foo", "")
			pGotFooWithFoo := ctrlutils.GotAnnotationPredicate("foo", "foo")
			pGotFooWithBar := ctrlutils.GotAnnotationPredicate("foo", "bar")
			pLostFoo := ctrlutils.LostAnnotationPredicate("foo", "")
			pLostFooWithFoo := ctrlutils.LostAnnotationPredicate("foo", "foo")
			pLostFooWithBar := ctrlutils.LostAnnotationPredicate("foo", "bar")
			By("old and new resource are equal")
			e := updateEvent(base, changed)
			Expect(pHasFoo.Update(e)).To(BeFalse(), "HasAnnotationPredicate should return false if there are no annotations")
			Expect(pHasFooWithFoo.Update(e)).To(BeFalse(), "HasAnnotationPredicate should return false if there are no annotations")
			Expect(pHasFooWithBar.Update(e)).To(BeFalse(), "HasAnnotationPredicate should return false if there are no annotations")
			Expect(pGotFoo.Update(e)).To(BeFalse(), "GotAnnotationPredicate should return false if there are no annotations")
			Expect(pGotFooWithFoo.Update(e)).To(BeFalse(), "GotAnnotationPredicate should return false if there are no annotations")
			Expect(pGotFooWithBar.Update(e)).To(BeFalse(), "GotAnnotationPredicate should return false if there are no annotations")
			Expect(pLostFoo.Update(e)).To(BeFalse(), "LostAnnotationPredicate should return false if there never were annotations")
			Expect(pLostFooWithFoo.Update(e)).To(BeFalse(), "LostAnnotationPredicate should return false if there never were annotations")
			Expect(pLostFooWithBar.Update(e)).To(BeFalse(), "LostAnnotationPredicate should return false if there never were annotations")
			By("add annotation foo=foo")
			changed.SetAnnotations(map[string]string{
				"foo": "foo",
			})
			e = updateEvent(base, changed)
			Expect(pHasFoo.Update(e)).To(BeTrue(), "HasAnnotationPredicate should return true if the annotation is there")
			Expect(pHasBar.Update(e)).To(BeFalse(), "HasAnnotationPredicate should return false if the annotation is not there")
			Expect(pHasFooWithFoo.Update(e)).To(BeTrue(), "HasAnnotationPredicate should return true if the annotation is there and has the fitting value")
			Expect(pHasFooWithBar.Update(e)).To(BeFalse(), "HasAnnotationPredicate should return false if the annotation is there but has the wrong value")
			Expect(pGotFoo.Update(e)).To(BeTrue(), "GotAnnotationPredicate should return true if the annotation was added")
			Expect(pGotFooWithFoo.Update(e)).To(BeTrue(), "GotAnnotationPredicate should return true if the annotation was added with the correct value")
			Expect(pGotFooWithBar.Update(e)).To(BeFalse(), "GotAnnotationPredicate should return false if the annotation was added but with the wrong value")
			Expect(pLostFoo.Update(e)).To(BeFalse(), "LostAnnotationPredicate should return false if the annotation wasn't there before")
			Expect(pLostFooWithFoo.Update(e)).To(BeFalse(), "LostAnnotationPredicate should return false if the annotation wasn't there before")
			Expect(pLostFooWithBar.Update(e)).To(BeFalse(), "LostAnnotationPredicate should return false if the annotation wasn't there before")
			By("change annotation to foo=bar")
			base = changed.DeepCopy()
			changed.SetAnnotations(map[string]string{
				"foo": "bar",
			})
			e = updateEvent(base, changed)
			Expect(pHasFoo.Update(e)).To(BeTrue(), "HasAnnotationPredicate should return true if the annotation is there")
			Expect(pHasBar.Update(e)).To(BeFalse(), "HasAnnotationPredicate should return false if the annotation is not there")
			Expect(pHasFooWithFoo.Update(e)).To(BeFalse(), "HasAnnotationPredicate should return false if the annotation is there but has the wrong value")
			Expect(pHasFooWithBar.Update(e)).To(BeTrue(), "HasAnnotationPredicate should return true if the annotation is there and has the fitting value")
			Expect(pGotFoo.Update(e)).To(BeFalse(), "GotAnnotationPredicate should return false if the annotation was there before")
			Expect(pGotFooWithFoo.Update(e)).To(BeFalse(), "GotAnnotationPredicate should return false if the annotation was changed but to the wrong value")
			Expect(pGotFooWithBar.Update(e)).To(BeTrue(), "GotAnnotationPredicate should return true if the annotation was changed to the correct value")
			Expect(pLostFoo.Update(e)).To(BeFalse(), "LostAnnotationPredicate should return false if the annotation was there before and is still there")
			Expect(pLostFooWithFoo.Update(e)).To(BeTrue(), "LostAnnotationPredicate should return true if the annotation had the correct value before and it was changed")
			Expect(pLostFooWithBar.Update(e)).To(BeFalse(), "LostAnnotationPredicate should return false if the annotation didn't have the correct value before")
			By("remove annotation foo")
			base = changed.DeepCopy()
			changed.SetAnnotations(nil)
			e = updateEvent(base, changed)
			Expect(pHasFoo.Update(e)).To(BeFalse(), "HasAnnotationPredicate should return false if there are no annotations")
			Expect(pHasFooWithFoo.Update(e)).To(BeFalse(), "HasAnnotationPredicate should return false if there are no annotations")
			Expect(pHasFooWithBar.Update(e)).To(BeFalse(), "HasAnnotationPredicate should return false if there are no annotations")
			Expect(pGotFoo.Update(e)).To(BeFalse(), "GotAnnotationPredicate should return false if there are no annotations")
			Expect(pGotFooWithFoo.Update(e)).To(BeFalse(), "GotAnnotationPredicate should return false if there are no annotations")
			Expect(pGotFooWithBar.Update(e)).To(BeFalse(), "GotAnnotationPredicate should return false if there are no annotations")
			Expect(pLostFoo.Update(e)).To(BeTrue(), "LostAnnotationPredicate should return true if the annotation was there before and got removed")
			Expect(pLostFooWithFoo.Update(e)).To(BeFalse(), "LostAnnotationPredicate should return false if the annotation was removed but didn't have the correct value before")
			Expect(pLostFooWithBar.Update(e)).To(BeTrue(), "LostAnnotationPredicate should return true if the annotation had the correct value before and was removed")
		})

	})

	Context("Labels", func() {

		It("should detect changes to the labels", func() {
			pHasFoo := ctrlutils.HasLabelPredicate("foo", "")
			pHasBar := ctrlutils.HasLabelPredicate("bar", "")
			matchesEverything := ctrlutils.LabelSelectorPredicate(labels.Everything())
			matchesNothing := ctrlutils.LabelSelectorPredicate(labels.Nothing())
			pHasFooWithFoo := ctrlutils.HasLabelPredicate("foo", "foo")
			pHasFooWithBar := ctrlutils.HasLabelPredicate("foo", "bar")
			fooWithFooSelector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
				MatchLabels: map[string]string{
					"foo": "foo",
				},
			})
			Expect(err).ToNot(HaveOccurred())
			fooWithBarSelector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
				MatchLabels: map[string]string{
					"foo": "bar",
				},
			})
			Expect(err).ToNot(HaveOccurred())
			pHasFooWithFooViaSelector := ctrlutils.LabelSelectorPredicate(fooWithFooSelector)
			pHasFooWithBarViaSelector := ctrlutils.LabelSelectorPredicate(fooWithBarSelector)
			pGotFoo := ctrlutils.GotLabelPredicate("foo", "")
			pGotFooWithFoo := ctrlutils.GotLabelPredicate("foo", "foo")
			pGotFooWithBar := ctrlutils.GotLabelPredicate("foo", "bar")
			pLostFoo := ctrlutils.LostLabelPredicate("foo", "")
			pLostFooWithFoo := ctrlutils.LostLabelPredicate("foo", "foo")
			pLostFooWithBar := ctrlutils.LostLabelPredicate("foo", "bar")
			By("old and new resource are equal")
			e := updateEvent(base, changed)
			Expect(matchesEverything.Update(e)).To(BeTrue(), "'everything' LabelSelector should always match")
			Expect(matchesNothing.Update(e)).To(BeFalse(), "'nothing' LabelSelector should never match")
			Expect(pHasFoo.Update(e)).To(BeFalse(), "HasLabelPredicate should return false if there are no labels")
			Expect(pHasFooWithFoo.Update(e)).To(BeFalse(), "HasLabelPredicate should return false if there are no labels")
			Expect(pHasFooWithFooViaSelector.Update(e)).To(BeFalse(), "LabelSelectorPredicate should return false if the labels are not matched")
			Expect(pHasFooWithBar.Update(e)).To(BeFalse(), "HasLabelPredicate should return false if there are no labels")
			Expect(pHasFooWithBarViaSelector.Update(e)).To(BeFalse(), "LabelSelectorPredicate should return falseif the labels are not matched")
			Expect(pGotFoo.Update(e)).To(BeFalse(), "GotLabelPredicate should return false if there are no labels")
			Expect(pGotFooWithFoo.Update(e)).To(BeFalse(), "GotLabelPredicate should return false if there are no labels")
			Expect(pGotFooWithBar.Update(e)).To(BeFalse(), "GotLabelPredicate should return false if there are no labels")
			Expect(pLostFoo.Update(e)).To(BeFalse(), "LostLabelPredicate should return false if there never were labels")
			Expect(pLostFooWithFoo.Update(e)).To(BeFalse(), "LostLabelPredicate should return false if there never were labels")
			Expect(pLostFooWithBar.Update(e)).To(BeFalse(), "LostLabelPredicate should return false if there never were labels")
			By("add label foo=foo")
			changed.SetLabels(map[string]string{
				"foo": "foo",
			})
			e = updateEvent(base, changed)
			Expect(matchesEverything.Update(e)).To(BeTrue(), "'everything' LabelSelector should always match")
			Expect(matchesNothing.Update(e)).To(BeFalse(), "'nothing' LabelSelector should never match")
			Expect(pHasFoo.Update(e)).To(BeTrue(), "HasLabelPredicate should return true if the label is there")
			Expect(pHasBar.Update(e)).To(BeFalse(), "HasLabelPredicate should return false if the label is not there")
			Expect(pHasFooWithFoo.Update(e)).To(BeTrue(), "HasLabelPredicate should return true if the label is there and has the fitting value")
			Expect(pHasFooWithFooViaSelector.Update(e)).To(BeTrue(), "LabelSelectorPredicate should return true if the labels are matched")
			Expect(pHasFooWithBar.Update(e)).To(BeFalse(), "HasLabelPredicate should return false if the label is there but has the wrong value")
			Expect(pHasFooWithBarViaSelector.Update(e)).To(BeFalse(), "LabelSelectorPredicate should return false if the labels are not matched")
			Expect(pGotFoo.Update(e)).To(BeTrue(), "GotLabelPredicate should return true if the label was added")
			Expect(pGotFooWithFoo.Update(e)).To(BeTrue(), "GotLabelPredicate should return true if the label was added with the correct value")
			Expect(pGotFooWithBar.Update(e)).To(BeFalse(), "GotLabelPredicate should return false if the label was added but with the wrong value")
			Expect(pLostFoo.Update(e)).To(BeFalse(), "LostLabelPredicate should return false if the label wasn't there before")
			Expect(pLostFooWithFoo.Update(e)).To(BeFalse(), "LostLabelPredicate should return false if the label wasn't there before")
			Expect(pLostFooWithBar.Update(e)).To(BeFalse(), "LostLabelPredicate should return false if the label wasn't there before")
			By("change label to foo=bar")
			base = changed.DeepCopy()
			changed.SetLabels(map[string]string{
				"foo": "bar",
			})
			e = updateEvent(base, changed)
			Expect(matchesEverything.Update(e)).To(BeTrue(), "'everything' LabelSelector should always match")
			Expect(matchesNothing.Update(e)).To(BeFalse(), "'nothing' LabelSelector should never match")
			Expect(pHasFoo.Update(e)).To(BeTrue(), "HasLabelPredicate should return true if the label is there")
			Expect(pHasBar.Update(e)).To(BeFalse(), "HasLabelPredicate should return false if the label is not there")
			Expect(pHasFooWithFoo.Update(e)).To(BeFalse(), "HasLabelPredicate should return false if the label is there but has the wrong value")
			Expect(pHasFooWithFooViaSelector.Update(e)).To(BeFalse(), "LabelSelectorPredicate should return false if the labels are not matched")
			Expect(pHasFooWithBar.Update(e)).To(BeTrue(), "HasLabelPredicate should return true if the label is there and has the fitting value")
			Expect(pHasFooWithBarViaSelector.Update(e)).To(BeTrue(), "LabelSelectorPredicate should return true if the labels are matched")
			Expect(pGotFoo.Update(e)).To(BeFalse(), "GotLabelPredicate should return false if the label was there before")
			Expect(pGotFooWithFoo.Update(e)).To(BeFalse(), "GotLabelPredicate should return false if the label was changed but to the wrong value")
			Expect(pGotFooWithBar.Update(e)).To(BeTrue(), "GotLabelPredicate should return true if the label was changed to the correct value")
			Expect(pLostFoo.Update(e)).To(BeFalse(), "LostLabelPredicate should return false if the label was there before and is still there")
			Expect(pLostFooWithFoo.Update(e)).To(BeTrue(), "LostLabelPredicate should return true if the label had the correct value before and it was changed")
			Expect(pLostFooWithBar.Update(e)).To(BeFalse(), "LostLabelPredicate should return false if the label didn't have the correct value before")
			By("remove label foo")
			base = changed.DeepCopy()
			changed.SetLabels(nil)
			e = updateEvent(base, changed)
			Expect(matchesEverything.Update(e)).To(BeTrue(), "'everything' LabelSelector should always match")
			Expect(matchesNothing.Update(e)).To(BeFalse(), "'nothing' LabelSelector should never match")
			Expect(pHasFoo.Update(e)).To(BeFalse(), "HasLabelPredicate should return false if there are no labels")
			Expect(pHasFooWithFoo.Update(e)).To(BeFalse(), "HasLabelPredicate should return false if there are no labels")
			Expect(pHasFooWithFooViaSelector.Update(e)).To(BeFalse(), "LabelSelectorPredicate should return false if the labels are not matched")
			Expect(pHasFooWithBar.Update(e)).To(BeFalse(), "HasLabelPredicate should return false if there are no labels")
			Expect(pHasFooWithBarViaSelector.Update(e)).To(BeFalse(), "LabelSelectorPredicate should return false if the labels are not matched")
			Expect(pGotFoo.Update(e)).To(BeFalse(), "GotLabelPredicate should return false if there are no labels")
			Expect(pGotFooWithFoo.Update(e)).To(BeFalse(), "GotLabelPredicate should return false if there are no labels")
			Expect(pGotFooWithBar.Update(e)).To(BeFalse(), "GotLabelPredicate should return false if there are no labels")
			Expect(pLostFoo.Update(e)).To(BeTrue(), "LostLabelPredicate should return true if the label was there before and got removed")
			Expect(pLostFooWithFoo.Update(e)).To(BeFalse(), "LostLabelPredicate should return false if the label was removed but didn't have the correct value before")
			Expect(pLostFooWithBar.Update(e)).To(BeTrue(), "LostLabelPredicate should return true if the label had the correct value before and was removed")
		})

	})

	Context("Status", func() {

		It("should detect changes to the status", func() {
			p := ctrlutils.StatusChangedPredicate{}
			Expect(p.Update(updateEvent(base, changed))).To(BeFalse(), "StatusChangedPredicate should return false if the status did not change")
			By("change status")
			changed.Status = corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{
							IP: "127.0.0.1",
						},
					},
				},
			}
			Expect(p.Update(updateEvent(base, changed))).To(BeTrue(), "StatusChangedPredicate should return true if the status changed")
		})

	})

})

func updateEvent(old, new client.Object) event.UpdateEvent {
	return event.UpdateEvent{
		ObjectOld: old,
		ObjectNew: new,
	}
}
