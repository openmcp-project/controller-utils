package conditions_test

import (
	"slices"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/conditions"
)

type conImpl struct {
	status             bool
	conType            string
	reason             string
	message            string
	lastTransitionTime time.Time
}

func newConImplWithValues(conType string, status bool, reason, message string, lastTransitionTime time.Time) *conImpl {
	return &conImpl{
		conType:            conType,
		status:             status,
		reason:             reason,
		message:            message,
		lastTransitionTime: lastTransitionTime,
	}
}

var _ conditions.Condition[bool] = &conImpl{}

func (c *conImpl) GetLastTransitionTime() time.Time {
	return c.lastTransitionTime
}

func (c *conImpl) GetType() string {
	return c.conType
}

func (c *conImpl) GetStatus() bool {
	return c.status
}

func (c *conImpl) GetReason() string {
	return c.reason
}

func (c *conImpl) GetMessage() string {
	return c.message
}

func (c *conImpl) SetStatus(status bool) {
	c.status = status
}

func (c *conImpl) SetType(conType string) {
	c.conType = conType
}

func (c *conImpl) SetLastTransitionTime(timestamp time.Time) {
	c.lastTransitionTime = timestamp
}

func (c *conImpl) SetReason(reason string) {
	c.reason = reason
}

func (c *conImpl) SetMessage(message string) {
	c.message = message
}

func testConditionSet() []conditions.Condition[bool] {
	now := time.Now().Add((-24) * time.Hour)
	return []conditions.Condition[bool]{
		newConImplWithValues("true", true, "reason", "message", now),
		newConImplWithValues("false", false, "reason", "message", now),
		newConImplWithValues("alsoTrue", true, "alsoReason", "alsoMessage", now),
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
			updated := conditions.ConditionUpdater(func() conditions.Condition[bool] { return &conImpl{} }, cons, false).UpdateCondition(oldCon.GetType(), oldCon.GetStatus(), "newReason", "newMessage").Conditions()
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
			updated := conditions.ConditionUpdater(func() conditions.Condition[bool] { return &conImpl{} }, cons, false).UpdateCondition(oldCon.GetType(), !oldCon.GetStatus(), "newReason", "newMessage").Conditions()
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
			updated := conditions.ConditionUpdater(func() conditions.Condition[bool] { return &conImpl{} }, cons, true).UpdateCondition(oldCon.GetType(), oldCon.GetStatus(), "newReason", "newMessage").Conditions()
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
			updated := conditions.ConditionUpdater(func() conditions.Condition[bool] { return &conImpl{} }, cons, true).UpdateCondition(oldCon.GetType(), !oldCon.GetStatus(), "newReason", "newMessage").Conditions()
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
				newConImplWithValues("c", true, "reason", "message", time.Now()),
				newConImplWithValues("d", true, "reason", "message", time.Now()),
				newConImplWithValues("a", true, "reason", "message", time.Now()),
				newConImplWithValues("b", true, "reason", "message", time.Now()),
			}
			compareConditions := func(a, b conditions.Condition[bool]) int {
				return strings.Compare(a.GetType(), b.GetType())
			}
			Expect(slices.IsSortedFunc(cons, compareConditions)).To(BeFalse(), "conditions in the test object are already sorted, unable to test sorting")
			updated := conditions.ConditionUpdater(func() conditions.Condition[bool] { return &conImpl{} }, cons, false).Conditions()
			Expect(len(updated)).To(BeNumerically(">", 1), "test object does not contain enough conditions to test sorting")
			Expect(len(updated)).To(Equal(len(cons)))
			Expect(slices.IsSortedFunc(updated, compareConditions)).To(BeTrue(), "conditions are not sorted")
		})

	})

})
