package conditions_test

import (
	"slices"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/openmcp-project/controller-utils/pkg/testing/matchers"

	"github.com/openmcp-project/controller-utils/pkg/conditions"
)

func testConditionSet() []conditions.Condition[bool] {
	now := time.Now().Add((-24) * time.Hour)
	return []conditions.Condition[bool]{
		NewConditionImplFromValues("true", true, "reason", "message", now),
		NewConditionImplFromValues("false", false, "reason", "message", now),
		NewConditionImplFromValues("alsoTrue", true, "alsoReason", "alsoMessage", now),
	}
}

var _ = Describe("Conditions", func() {

	Context("GetCondition", func() {

		It("should return the requested condition", func() {
			cons := testConditionSet()

			con := conditions.GetCondition(cons, "true")
			Expect(con).ToNot(BeNil())
			Expect(con.GetType()).To(Equal("true"))
			Expect(con.GetStatus()).To(BeTrue())

			con = conditions.GetCondition(cons, "false")
			Expect(con).ToNot(BeNil())
			Expect(con.GetType()).To(Equal("false"))
			Expect(con.GetStatus()).To(BeFalse())

			con = conditions.GetCondition(cons, "alsoTrue")
			Expect(con).ToNot(BeNil())
			Expect(con.GetType()).To(Equal("alsoTrue"))
			Expect(con.GetStatus()).To(BeTrue())
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
			Expect(con.GetType()).To(Equal("true"))
			Expect(con.GetStatus()).To(BeTrue())

			con.SetStatus(false)
			con = conditions.GetCondition(cons, "true")
			Expect(con).ToNot(BeNil())
			Expect(con.GetType()).To(Equal("true"))
			Expect(con.GetStatus()).To(BeFalse())
		})

	})

	Context("ConditionUpdater", func() {

		It("should update the condition (same value, keep other cons)", func() {
			cons := testConditionSet()
			oldCon := conditions.GetCondition(cons, "true")
			updated, changed := conditions.ConditionUpdater(NewCondition[bool], cons, false).UpdateCondition(oldCon.GetType(), oldCon.GetStatus(), "newReason", "newMessage").Conditions()
			Expect(changed).To(BeTrue())
			newCon := conditions.GetCondition(updated, "true")
			Expect(updated).To(HaveLen(len(cons)))
			Expect(newCon).ToNot(Equal(oldCon))
			Expect(newCon.GetType()).To(Equal(oldCon.GetType()))
			Expect(newCon.GetStatus()).To(Equal(oldCon.GetStatus()))
			Expect(newCon.GetReason()).To(Equal("newReason"))
			Expect(newCon.GetMessage()).To(Equal("newMessage"))
			Expect(oldCon.GetReason()).To(Equal("reason"))
			Expect(oldCon.GetMessage()).To(Equal("message"))
			Expect(newCon.GetLastTransitionTime()).To(Equal(oldCon.GetLastTransitionTime()))
		})

		It("should update the condition (different value, keep other cons)", func() {
			cons := testConditionSet()
			oldCon := conditions.GetCondition(cons, "true")
			updated, changed := conditions.ConditionUpdater(NewCondition[bool], cons, false).UpdateCondition(oldCon.GetType(), !oldCon.GetStatus(), "newReason", "newMessage").Conditions()
			Expect(changed).To(BeTrue())
			newCon := conditions.GetCondition(updated, "true")
			Expect(updated).To(HaveLen(len(cons)))
			Expect(newCon).ToNot(Equal(oldCon))
			Expect(newCon.GetType()).To(Equal(oldCon.GetType()))
			Expect(newCon.GetStatus()).To(Equal(!oldCon.GetStatus()))
			Expect(newCon.GetReason()).To(Equal("newReason"))
			Expect(newCon.GetMessage()).To(Equal("newMessage"))
			Expect(oldCon.GetReason()).To(Equal("reason"))
			Expect(oldCon.GetMessage()).To(Equal("message"))
			Expect(newCon.GetLastTransitionTime()).ToNot(Equal(oldCon.GetLastTransitionTime()))
			Expect(newCon.GetLastTransitionTime().After(oldCon.GetLastTransitionTime())).To(BeTrue())
		})

		It("should update the condition (same value, discard other cons)", func() {
			cons := testConditionSet()
			oldCon := conditions.GetCondition(cons, "true")
			updated, changed := conditions.ConditionUpdater(NewCondition[bool], cons, true).UpdateCondition(oldCon.GetType(), oldCon.GetStatus(), "newReason", "newMessage").Conditions()
			Expect(changed).To(BeTrue())
			newCon := conditions.GetCondition(updated, "true")
			Expect(updated).To(HaveLen(1))
			Expect(newCon).ToNot(Equal(oldCon))
			Expect(newCon.GetType()).To(Equal(oldCon.GetType()))
			Expect(newCon.GetStatus()).To(Equal(oldCon.GetStatus()))
			Expect(newCon.GetReason()).To(Equal("newReason"))
			Expect(newCon.GetMessage()).To(Equal("newMessage"))
			Expect(oldCon.GetReason()).To(Equal("reason"))
			Expect(oldCon.GetMessage()).To(Equal("message"))
			Expect(newCon.GetLastTransitionTime()).To(Equal(oldCon.GetLastTransitionTime()))
		})

		It("should update the condition (different value, discard other cons)", func() {
			cons := testConditionSet()
			oldCon := conditions.GetCondition(cons, "true")
			updated, changed := conditions.ConditionUpdater(NewCondition[bool], cons, true).UpdateCondition(oldCon.GetType(), !oldCon.GetStatus(), "newReason", "newMessage").Conditions()
			Expect(changed).To(BeTrue())
			newCon := conditions.GetCondition(updated, "true")
			Expect(updated).To(HaveLen(1))
			Expect(newCon).ToNot(Equal(oldCon))
			Expect(newCon.GetType()).To(Equal(oldCon.GetType()))
			Expect(newCon.GetStatus()).To(Equal(!oldCon.GetStatus()))
			Expect(newCon.GetReason()).To(Equal("newReason"))
			Expect(newCon.GetMessage()).To(Equal("newMessage"))
			Expect(oldCon.GetReason()).To(Equal("reason"))
			Expect(oldCon.GetMessage()).To(Equal("message"))
			Expect(newCon.GetLastTransitionTime()).ToNot(Equal(oldCon.GetLastTransitionTime()))
			Expect(newCon.GetLastTransitionTime().After(oldCon.GetLastTransitionTime())).To(BeTrue())
		})

		It("should sort the conditions by type", func() {
			cons := []conditions.Condition[bool]{
				NewConditionImplFromValues("c", true, "reason", "message", time.Now()),
				NewConditionImplFromValues("d", true, "reason", "message", time.Now()),
				NewConditionImplFromValues("a", true, "reason", "message", time.Now()),
				NewConditionImplFromValues("b", true, "reason", "message", time.Now()),
			}
			compareConditions := func(a, b conditions.Condition[bool]) int {
				return strings.Compare(a.GetType(), b.GetType())
			}
			Expect(slices.IsSortedFunc(cons, compareConditions)).To(BeFalse(), "conditions in the test object are already sorted, unable to test sorting")
			updated, changed := conditions.ConditionUpdater(NewCondition[bool], cons, false).Conditions()
			Expect(changed).To(BeFalse())
			Expect(len(updated)).To(BeNumerically(">", 1), "test object does not contain enough conditions to test sorting")
			Expect(len(updated)).To(Equal(len(cons)))
			Expect(slices.IsSortedFunc(updated, compareConditions)).To(BeTrue(), "conditions are not sorted")
		})

		It("should remove a condition", func() {
			cons := testConditionSet()
			updated, changed := conditions.ConditionUpdater(NewCondition[bool], cons, false).RemoveCondition("true").Conditions()
			Expect(changed).To(BeTrue())
			Expect(updated).To(HaveLen(len(cons) - 1))
			con := conditions.GetCondition(updated, "true")
			Expect(con).To(BeNil())

			// removing a condition that does not exist should not change anything
			updated, changed = conditions.ConditionUpdater(NewCondition[bool], cons, false).RemoveCondition("doesNotExist").Conditions()
			Expect(changed).To(BeFalse())
			Expect(updated).To(HaveLen(len(cons)))
		})

		It("should not mark a condition as changed if it has the same values as before", func() {
			cons := testConditionSet()
			con := conditions.GetCondition(cons, "true")
			updated, changed := conditions.ConditionUpdater(NewCondition[bool], cons, false).UpdateCondition(con.GetType(), con.GetStatus(), con.GetReason(), con.GetMessage()).Conditions()
			Expect(changed).To(BeFalse())
			Expect(updated).To(HaveLen(len(cons)))
		})

		It("should return that a condition exists only if it will be contained in the returned list", func() {
			cons := testConditionSet()
			updater := conditions.ConditionUpdater(NewCondition[bool], cons, false)
			Expect(updater.HasCondition("true")).To(BeTrue())
			Expect(updater.HasCondition("doesNotExist")).To(BeFalse())
			updater = conditions.ConditionUpdater(NewCondition[bool], cons, true)
			Expect(updater.HasCondition("true")).To(BeFalse())
			Expect(updater.HasCondition("doesNotExist")).To(BeFalse())
			updater.UpdateCondition("true", true, "reason", "message")
			Expect(updater.HasCondition("true")).To(BeTrue())
		})

	})

})
