package matchers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/openmcp-project/controller-utils/pkg/testing/matchers"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MatchCondition", func() {

	It("should match if expected and actual are nil", func() {
		Expect(MatchCondition(nil).Match(nil)).To(BeTrue())
	})

	It("should match if expected and actual are identical", func() {
		expected := TestConditionFromValues("test", metav1.ConditionTrue, 0, "reason", "message", metav1.Now())
		Expect(MatchCondition(expected).Match(expected.ToCondition())).To(BeTrue())
	})

	It("should match if expected is a Condition generated from the actual metav1.Condition", func() {
		actual := metav1.Condition{
			Type:               "test",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: 0,
			Reason:             "reason",
			Message:            "message",
			LastTransitionTime: metav1.Now(),
		}
		expected := TestConditionFromCondition(actual)
		Expect(MatchCondition(expected).Match(&actual)).To(BeTrue())
	})

	It("should match if actual is a metav1.Condition generated from the expected Condition", func() {
		expected := TestConditionFromValues("test", metav1.ConditionTrue, 0, "reason", "message", metav1.Now())
		actual := expected.ToCondition()
		Expect(MatchCondition(expected).Match(actual)).To(BeTrue())
	})

	It("should match any unset fields in the expected condition", func() {
		actual := metav1.Condition{
			Type:               "test",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: 0,
			Reason:             "reason",
			Message:            "message",
			LastTransitionTime: metav1.Now(),
		}

		Expect(MatchCondition(TestCondition().WithType(actual.Type).WithStatus(actual.Status).WithObservedGeneration(actual.ObservedGeneration).WithReason(actual.Reason).WithMessage(actual.Message).WithLastTransitionTime(actual.LastTransitionTime)).Match(actual)).To(BeTrue())
		Expect(MatchCondition(TestCondition().WithStatus(actual.Status).WithObservedGeneration(actual.ObservedGeneration).WithReason(actual.Reason).WithMessage(actual.Message).WithLastTransitionTime(actual.LastTransitionTime)).Match(actual)).To(BeTrue())
		Expect(MatchCondition(TestCondition().WithType(actual.Type).WithObservedGeneration(actual.ObservedGeneration).WithReason(actual.Reason).WithMessage(actual.Message).WithLastTransitionTime(actual.LastTransitionTime)).Match(actual)).To(BeTrue())
		Expect(MatchCondition(TestCondition().WithType(actual.Type).WithStatus(actual.Status).WithReason(actual.Reason).WithMessage(actual.Message).WithLastTransitionTime(actual.LastTransitionTime)).Match(actual)).To(BeTrue())
		Expect(MatchCondition(TestCondition().WithType(actual.Type).WithStatus(actual.Status).WithObservedGeneration(actual.ObservedGeneration).WithMessage(actual.Message).WithLastTransitionTime(actual.LastTransitionTime)).Match(actual)).To(BeTrue())
		Expect(MatchCondition(TestCondition().WithType(actual.Type).WithStatus(actual.Status).WithObservedGeneration(actual.ObservedGeneration).WithReason(actual.Reason).WithLastTransitionTime(actual.LastTransitionTime)).Match(actual)).To(BeTrue())
		Expect(MatchCondition(TestCondition().WithType(actual.Type).WithStatus(actual.Status).WithObservedGeneration(actual.ObservedGeneration).WithReason(actual.Reason).WithMessage(actual.Message)).Match(actual)).To(BeTrue())
	})

	It("should match independent of whether actual is a pointer or value", func() {
		actual := metav1.Condition{
			Type:               "test",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: 0,
			Reason:             "reason",
			Message:            "message",
			LastTransitionTime: metav1.Now(),
		}
		expected := TestConditionFromCondition(actual)

		Expect(MatchCondition(expected).Match(&actual)).To(BeTrue())
		Expect(MatchCondition(expected).Match(actual)).To(BeTrue())
		Expect(MatchCondition(expected).Match(expected)).To(BeTrue())
		Expect(MatchCondition(expected).Match(*expected)).To(BeTrue())
	})

})
