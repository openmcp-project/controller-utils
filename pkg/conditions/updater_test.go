package conditions_test

import (
	"slices"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/openmcp-project/controller-utils/pkg/testing/matchers"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

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

		It("should not change the last transition time if the same condition is updated multiple times with the last update setting it to the original value again", func() {
			cons := testConditionSet()
			oldCon := conditions.GetCondition(cons, "true")
			updater := conditions.ConditionUpdater(cons, false)
			updated, changed := updater.UpdateCondition(oldCon.Type, invert(oldCon.Status), oldCon.ObservedGeneration+1, "newReason", "newMessage").
				UpdateCondition(oldCon.Type, oldCon.Status, oldCon.ObservedGeneration, oldCon.Reason, oldCon.Message).Conditions()
			Expect(changed).To(BeFalse())
			Expect(updated).To(ConsistOf(cons))
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

		It("should correctly add a reason if not given and replace invalid characters from the type", func() {
			cons := []metav1.Condition{}
			updated, _ := conditions.ConditionUpdater(cons, false).
				UpdateCondition("TestCondition", conditions.FromBool(true), 0, "", "").
				UpdateCondition("TestCondition.Test", conditions.FromBool(false), 0, "", "").
				Conditions()
			Expect(updated).To(ConsistOf(
				MatchCondition(TestCondition().
					WithType("TestCondition").
					WithStatus(metav1.ConditionTrue).
					WithReason("TestCondition_True").
					WithMessage("")),
				MatchCondition(TestCondition().
					WithType("TestCondition.Test").
					WithStatus(metav1.ConditionFalse).
					WithReason("TestCondition:Test_False").
					WithMessage("")),
			))
		})

	})

	Context("EventRecorder", func() {

		var recorder *record.FakeRecorder

		dummy := &dummyObject{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Dummy",
				APIVersion: "dummy/v1",
			},
		}

		BeforeEach(func() {
			recorder = record.NewFakeRecorder(100)
		})

		AfterEach(func() {
			close(recorder.Events)
		})

		Context("Verbosity: EventPerChange", func() {

			It("should record an event for each changed condition status", func() {
				cons := testConditionSet()
				updater := conditions.ConditionUpdater(cons, false).WithEventRecorder(recorder, conditions.EventPerChange)
				trueCon1 := conditions.GetCondition(cons, "true")
				trueCon2 := conditions.GetCondition(cons, "alsoTrue")
				_, changed := updater.
					UpdateCondition(trueCon1.Type, invert(trueCon1.Status), trueCon1.ObservedGeneration+1, "newReason", "newMessage").
					UpdateCondition(trueCon2.Type, invert(trueCon2.Status), trueCon2.ObservedGeneration+1, "newReason", "newMessage").
					Record(dummy).Conditions()
				Expect(changed).To(BeTrue())

				events := flush(recorder.Events)
				Expect(events).To(ConsistOf(
					ContainSubstring("Condition '%s' changed from '%s' to '%s'", trueCon1.Type, trueCon1.Status, invert(trueCon1.Status)),
					ContainSubstring("Condition '%s' changed from '%s' to '%s'", trueCon2.Type, trueCon2.Status, invert(trueCon2.Status)),
				))
			})

			It("should not record any events if no condition status changed, even if other fields changed", func() {
				cons := testConditionSet()
				updater := conditions.ConditionUpdater(cons, false).WithEventRecorder(recorder, conditions.EventPerChange)
				trueCon1 := conditions.GetCondition(cons, "true")
				trueCon2 := conditions.GetCondition(cons, "alsoTrue")
				_, changed := updater.
					UpdateCondition(trueCon1.Type, trueCon1.Status, trueCon1.ObservedGeneration+1, "newReason", "newMessage").
					UpdateCondition(trueCon2.Type, trueCon2.Status, trueCon2.ObservedGeneration+1, "newReason", "newMessage").
					Record(dummy).Conditions()
				Expect(changed).To(BeTrue())

				events := flush(recorder.Events)
				Expect(events).To(BeEmpty())
			})

			It("should record added and lost conditions", func() {
				cons := testConditionSet()
				updater := conditions.ConditionUpdater(cons, false).WithEventRecorder(recorder, conditions.EventPerChange)
				_, changed := updater.
					UpdateCondition("new", metav1.ConditionUnknown, 1, "newReason", "newMessage").
					RemoveCondition("true").
					Record(dummy).Conditions()
				Expect(changed).To(BeTrue())

				events := flush(recorder.Events)
				Expect(events).To(ConsistOf(
					ContainSubstring("Condition 'new' added with status '%s'", metav1.ConditionUnknown),
					ContainSubstring("Condition 'true' with status '%s' removed", metav1.ConditionTrue),
				))
			})

		})

		Context("Verbosity: EventPerNewStatus", func() {

			It("should record an event for each new status that any condition has reached", func() {
				cons := testConditionSet()
				updater := conditions.ConditionUpdater(cons, false).WithEventRecorder(recorder, conditions.EventPerNewStatus)
				trueCon1 := conditions.GetCondition(cons, "true")
				trueCon2 := conditions.GetCondition(cons, "alsoTrue")
				falseCon := conditions.GetCondition(cons, "false")
				_, changed := updater.
					UpdateCondition(trueCon1.Type, metav1.ConditionUnknown, trueCon1.ObservedGeneration+1, "newReason", "newMessage").
					UpdateCondition(trueCon2.Type, metav1.ConditionUnknown, trueCon2.ObservedGeneration+1, "newReason", "newMessage").
					UpdateCondition("anotherOne", metav1.ConditionUnknown, 1, "newReason", "newMessage").
					UpdateCondition(falseCon.Type, invert(falseCon.Status), falseCon.ObservedGeneration+1, "newReason", "newMessage").
					Record(dummy).Conditions()
				Expect(changed).To(BeTrue())

				events := flush(recorder.Events)
				Expect(events).To(ConsistOf(
					And(
						ContainSubstring("The following conditions changed to '%s'", metav1.ConditionUnknown),
						ContainSubstring(trueCon1.Type),
						ContainSubstring(trueCon2.Type),
						ContainSubstring("anotherOne"),
					),
					And(
						ContainSubstring("The following conditions changed to '%s'", invert(falseCon.Status)),
						ContainSubstring(falseCon.Type),
					),
				))
			})

			It("should not record any events if no condition status changed, even if other fields changed", func() {
				cons := testConditionSet()
				updater := conditions.ConditionUpdater(cons, false).WithEventRecorder(recorder, conditions.EventPerNewStatus)
				trueCon1 := conditions.GetCondition(cons, "true")
				trueCon2 := conditions.GetCondition(cons, "alsoTrue")
				_, changed := updater.
					UpdateCondition(trueCon1.Type, trueCon1.Status, trueCon1.ObservedGeneration+1, "newReason", "newMessage").
					UpdateCondition(trueCon2.Type, trueCon2.Status, trueCon2.ObservedGeneration+1, "newReason", "newMessage").
					Record(dummy).Conditions()
				Expect(changed).To(BeTrue())

				events := flush(recorder.Events)
				Expect(events).To(BeEmpty())
			})

			It("should record lost conditions", func() {
				cons := testConditionSet()
				updater := conditions.ConditionUpdater(cons, true).WithEventRecorder(recorder, conditions.EventPerNewStatus)
				_, changed := updater.Record(dummy).Conditions()
				Expect(changed).To(BeTrue())

				events := flush(recorder.Events)
				Expect(events).To(ConsistOf(
					And(
						ContainSubstring("The following conditions were removed"),
						ContainSubstring("true"),
						ContainSubstring("false"),
						ContainSubstring("alsoTrue"),
					),
				))
			})

		})

		Context("Verbosity: EventIfChanged", func() {

			It("should only record a single event, no matter how many conditions changed", func() {
				cons := testConditionSet()
				updater := conditions.ConditionUpdater(cons, false).WithEventRecorder(recorder, conditions.EventIfChanged)
				trueCon := conditions.GetCondition(cons, "true")
				falseCon := conditions.GetCondition(cons, "false")
				_, changed := updater.
					UpdateCondition(trueCon.Type, invert(trueCon.Status), trueCon.ObservedGeneration+1, "newReason", "newMessage").
					UpdateCondition(falseCon.Type, invert(falseCon.Status), falseCon.ObservedGeneration+1, "newReason", "newMessage").
					UpdateCondition("new", metav1.ConditionUnknown, 1, "newReason", "newMessage").
					RemoveCondition("alsoTrue").
					Record(dummy).Conditions()
				Expect(changed).To(BeTrue())

				events := flush(recorder.Events)
				Expect(events).To(ConsistOf(
					And(
						ContainSubstring("The following conditions have changed"),
						ContainSubstring("true"),
						ContainSubstring("false"),
						ContainSubstring("alsoTrue"),
						ContainSubstring("new"),
					),
				))
			})

			It("should not record any events if no condition status changed, even if other fields changed", func() {
				cons := testConditionSet()
				updater := conditions.ConditionUpdater(cons, false).WithEventRecorder(recorder, conditions.EventIfChanged)
				trueCon1 := conditions.GetCondition(cons, "true")
				trueCon2 := conditions.GetCondition(cons, "alsoTrue")
				_, changed := updater.
					UpdateCondition(trueCon1.Type, trueCon1.Status, trueCon1.ObservedGeneration+1, "newReason", "newMessage").
					UpdateCondition(trueCon2.Type, trueCon2.Status, trueCon2.ObservedGeneration+1, "newReason", "newMessage").
					Record(dummy).Conditions()
				Expect(changed).To(BeTrue())

				events := flush(recorder.Events)
				Expect(events).To(BeEmpty())
			})

		})

	})

})

type dummyObject struct {
	metav1.TypeMeta `json:",inline"`
}

func (d *dummyObject) DeepCopyObject() runtime.Object {
	return &dummyObject{
		TypeMeta: d.TypeMeta,
	}
}

func flush(c chan string) []string {
	res := []string{}
	for len(c) > 0 {
		res = append(res, <-c)
	}
	return res
}
