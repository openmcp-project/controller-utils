package conditions_test

import (
	"slices"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/openmcp-project/controller-utils/pkg/testing/matchers"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openmcp-project/controller-utils/pkg/conditions"
)

func testConditionSet() []metav1.Condition {
	now := metav1.NewTime(time.Now().Add((-24) * time.Hour))
	return []metav1.Condition{
		TestConditionFromValues("true", conditions.FromBool(true), 0, "reason", "message", now).ToCondition(),
		TestConditionFromValues("false", conditions.FromBool(false), 0, "reason", "message", now).ToCondition(),
		TestConditionFromValues("alsoTrue", conditions.FromBool(true), 0, "alsoReason", "alsoMessage", now).ToCondition(),
	}
}

func invert(status metav1.ConditionStatus) metav1.ConditionStatus {
	if status == metav1.ConditionTrue {
		return metav1.ConditionFalse
	}
	if status == metav1.ConditionFalse {
		return metav1.ConditionTrue
	}
	return metav1.ConditionUnknown
}

var _ = Describe("Conditions", func() {

	Context("GetCondition", func() {

		It("should return the requested condition", func() {
			cons := testConditionSet()

			con := conditions.GetCondition(cons, "true")
			Expect(con).ToNot(BeNil())
			Expect(con.Type).To(Equal("true"))
			Expect(con.Status).To(Equal(metav1.ConditionTrue))

			con = conditions.GetCondition(cons, "false")
			Expect(con).ToNot(BeNil())
			Expect(con.Type).To(Equal("false"))
			Expect(con.Status).To(Equal(metav1.ConditionFalse))

			con = conditions.GetCondition(cons, "alsoTrue")
			Expect(con).ToNot(BeNil())
			Expect(con.Type).To(Equal("alsoTrue"))
			Expect(con.Status).To(Equal(metav1.ConditionTrue))
		})

		It("should return nil if the condition does not exist", func() {
			cons := testConditionSet()

			con := conditions.GetCondition(cons, "doesNotExist")
			Expect(con).To(BeNil())
		})

		It("should return a pointer to the condition, so that it can be changed", func() {
			// This depends on the actual implementation of the Condition interface.
			// However, most implementations will probably look very similar to the one in this test.
			cons := testConditionSet()

			con := conditions.GetCondition(cons, "true")
			Expect(con).ToNot(BeNil())
			Expect(con.Type).To(Equal("true"))
			Expect(con.Status).To(Equal(metav1.ConditionTrue))

			con.Status = metav1.ConditionFalse
			con = conditions.GetCondition(cons, "true")
			Expect(con).ToNot(BeNil())
			Expect(con.Type).To(Equal("true"))
			Expect(con.Status).To(Equal(metav1.ConditionFalse))
		})

	})

	Context("ConditionUpdater", func() {

		It("should update the condition (same value, keep other cons)", func() {
			cons := testConditionSet()
			oldCon := conditions.GetCondition(cons, "true")
			updated, changed := conditions.ConditionUpdater(cons, false).UpdateCondition(oldCon.Type, oldCon.Status, oldCon.ObservedGeneration, "newReason", "newMessage").Conditions()
			Expect(changed).To(BeTrue())
			newCon := conditions.GetCondition(updated, "true")
			Expect(updated).To(HaveLen(len(cons)))
			Expect(newCon).ToNot(Equal(oldCon))
			Expect(newCon.Type).To(Equal(oldCon.Type))
			Expect(newCon.Status).To(Equal(oldCon.Status))
			Expect(newCon.ObservedGeneration).To(Equal(oldCon.ObservedGeneration))
			Expect(newCon.Reason).To(Equal("newReason"))
			Expect(newCon.Message).To(Equal("newMessage"))
			Expect(oldCon.Reason).To(Equal("reason"))
			Expect(oldCon.Message).To(Equal("message"))
			Expect(newCon.LastTransitionTime).To(Equal(oldCon.LastTransitionTime))
		})

		It("should update the condition (different value, keep other cons)", func() {
			cons := testConditionSet()
			oldCon := conditions.GetCondition(cons, "true")
			updated, changed := conditions.ConditionUpdater(cons, false).UpdateCondition(oldCon.Type, invert(oldCon.Status), oldCon.ObservedGeneration+1, "newReason", "newMessage").Conditions()
			Expect(changed).To(BeTrue())
			newCon := conditions.GetCondition(updated, "true")
			Expect(updated).To(HaveLen(len(cons)))
			Expect(newCon).ToNot(Equal(oldCon))
			Expect(newCon.Type).To(Equal(oldCon.Type))
			Expect(newCon.Status).To(Equal(invert(oldCon.Status)))
			Expect(newCon.ObservedGeneration).To(Equal(oldCon.ObservedGeneration + 1))
			Expect(newCon.Reason).To(Equal("newReason"))
			Expect(newCon.Message).To(Equal("newMessage"))
			Expect(oldCon.Reason).To(Equal("reason"))
			Expect(oldCon.Message).To(Equal("message"))
			Expect(newCon.LastTransitionTime).ToNot(Equal(oldCon.LastTransitionTime))
			Expect(newCon.LastTransitionTime.After(oldCon.LastTransitionTime.Time)).To(BeTrue())
		})

		It("should update the condition (same value, discard other cons)", func() {
			cons := testConditionSet()
			oldCon := conditions.GetCondition(cons, "true")
			updated, changed := conditions.ConditionUpdater(cons, true).UpdateCondition(oldCon.Type, oldCon.Status, oldCon.ObservedGeneration, "newReason", "newMessage").Conditions()
			Expect(changed).To(BeTrue())
			newCon := conditions.GetCondition(updated, "true")
			Expect(updated).To(HaveLen(1))
			Expect(newCon).ToNot(Equal(oldCon))
			Expect(newCon.Type).To(Equal(oldCon.Type))
			Expect(newCon.Status).To(Equal(oldCon.Status))
			Expect(newCon.ObservedGeneration).To(Equal(oldCon.ObservedGeneration))
			Expect(newCon.Reason).To(Equal("newReason"))
			Expect(newCon.Message).To(Equal("newMessage"))
			Expect(oldCon.Reason).To(Equal("reason"))
			Expect(oldCon.Message).To(Equal("message"))
			Expect(newCon.LastTransitionTime).To(Equal(oldCon.LastTransitionTime))
		})

		It("should update the condition (different value, discard other cons)", func() {
			cons := testConditionSet()
			oldCon := conditions.GetCondition(cons, "true")
			updated, changed := conditions.ConditionUpdater(cons, true).UpdateCondition(oldCon.Type, invert(oldCon.Status), oldCon.ObservedGeneration+1, "newReason", "newMessage").Conditions()
			Expect(changed).To(BeTrue())
			newCon := conditions.GetCondition(updated, "true")
			Expect(updated).To(HaveLen(1))
			Expect(newCon).ToNot(Equal(oldCon))
			Expect(newCon.Type).To(Equal(oldCon.Type))
			Expect(newCon.Status).To(Equal(invert(oldCon.Status)))
			Expect(newCon.ObservedGeneration).To(Equal(oldCon.ObservedGeneration + 1))
			Expect(newCon.Reason).To(Equal("newReason"))
			Expect(newCon.Message).To(Equal("newMessage"))
			Expect(oldCon.Reason).To(Equal("reason"))
			Expect(oldCon.Message).To(Equal("message"))
			Expect(newCon.LastTransitionTime).ToNot(Equal(oldCon.LastTransitionTime))
			Expect(newCon.LastTransitionTime.After(oldCon.LastTransitionTime.Time)).To(BeTrue())
		})

		It("should sort the conditions by type", func() {
			cons := []metav1.Condition{
				TestConditionFromValues("c", conditions.FromBool(true), 0, "reason", "message", metav1.Now()).ToCondition(),
				TestConditionFromValues("d", conditions.FromBool(true), 0, "reason", "message", metav1.Now()).ToCondition(),
				TestConditionFromValues("a", conditions.FromBool(true), 0, "reason", "message", metav1.Now()).ToCondition(),
				TestConditionFromValues("b", conditions.FromBool(true), 0, "reason", "message", metav1.Now()).ToCondition(),
			}
			compareConditions := func(a, b metav1.Condition) int {
				return strings.Compare(a.Type, b.Type)
			}
			Expect(slices.IsSortedFunc(cons, compareConditions)).To(BeFalse(), "conditions in the test object are already sorted, unable to test sorting")
			updated, changed := conditions.ConditionUpdater(cons, false).Conditions()
			Expect(changed).To(BeFalse())
			Expect(len(updated)).To(BeNumerically(">", 1), "test object does not contain enough conditions to test sorting")
			Expect(len(updated)).To(Equal(len(cons)))
			Expect(slices.IsSortedFunc(updated, compareConditions)).To(BeTrue(), "conditions are not sorted")
		})

		It("should remove a condition", func() {
			cons := testConditionSet()
			updated, changed := conditions.ConditionUpdater(cons, false).RemoveCondition("true").Conditions()
			Expect(changed).To(BeTrue())
			Expect(updated).To(HaveLen(len(cons) - 1))
			con := conditions.GetCondition(updated, "true")
			Expect(con).To(BeNil())

			// removing a condition that does not exist should not change anything
			updated, changed = conditions.ConditionUpdater(cons, false).RemoveCondition("doesNotExist").Conditions()
			Expect(changed).To(BeFalse())
			Expect(updated).To(HaveLen(len(cons)))
		})

		It("should not mark a condition as changed if it has the same values as before", func() {
			cons := testConditionSet()
			con := conditions.GetCondition(cons, "true")
			updated, changed := conditions.ConditionUpdater(cons, false).UpdateCondition(con.Type, con.Status, con.ObservedGeneration, con.Reason, con.Message).Conditions()
			Expect(changed).To(BeFalse())
			Expect(updated).To(HaveLen(len(cons)))
		})

		It("should return that a condition exists only if it will be contained in the returned list", func() {
			cons := testConditionSet()
			updater := conditions.ConditionUpdater(cons, false)
			Expect(updater.HasCondition("true")).To(BeTrue())
			Expect(updater.HasCondition("doesNotExist")).To(BeFalse())
			updater = conditions.ConditionUpdater(cons, true)
			Expect(updater.HasCondition("true")).To(BeFalse())
			Expect(updater.HasCondition("doesNotExist")).To(BeFalse())
			updater.UpdateCondition("true", metav1.ConditionTrue, 1, "reason", "message")
			Expect(updater.HasCondition("true")).To(BeTrue())
		})

	})

})
