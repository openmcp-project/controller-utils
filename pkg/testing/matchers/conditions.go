package matchers

import (
	"fmt"
	"reflect"
	"time"

	"github.com/onsi/gomega/types"

	"github.com/openmcp-project/controller-utils/pkg/conditions"
)

// MatchCondition returns a Gomega matcher that checks if a Condition is equal to the expected one.
// If the passed in 'actual' is not a Condition, the matcher will fail.
// All fields which are set to their zero value in the expected condition will be ignored.
func MatchCondition[T comparable](con *ConditionImpl[T]) types.GomegaMatcher {
	return &conditionMatcher[T]{expected: con}
}

type conditionMatcher[T comparable] struct {
	expected *ConditionImpl[T]
}

func (c *conditionMatcher[T]) GomegaString() string {
	if c == nil || c.expected == nil {
		return "<nil>"
	}
	return c.expected.String()
}

var _ types.GomegaMatcher = &conditionMatcher[bool]{}

// Match implements types.GomegaMatcher.
func (c *conditionMatcher[T]) Match(actualRaw any) (success bool, err error) {
	actual, converted := actualRaw.(conditions.Condition[T])
	if !converted {
		// actualRaw doesn't implement conditions.Condition[T], check if a pointer to it does
		ptrValue := reflect.New(reflect.TypeOf(actualRaw))
		reflect.Indirect(ptrValue).Set(reflect.ValueOf(actualRaw))
		actual, converted = ptrValue.Interface().(conditions.Condition[T])
	}
	if !converted {
		return false, fmt.Errorf("expected actual (or &actual) to be of type Condition[%s], got %T", reflect.TypeFor[T]().Name(), actualRaw)
	}
	if actual == nil && c.expected == nil {
		return true, nil
	}
	if actual == nil || c.expected == nil {
		return false, nil
	}
	if c.expected.HasType() && c.expected.GetType() != actual.GetType() {
		return false, nil
	}
	if c.expected.HasStatus() && c.expected.GetStatus() != actual.GetStatus() {
		return false, nil
	}
	if c.expected.HasReason() && c.expected.GetReason() != actual.GetReason() {
		return false, nil
	}
	if c.expected.HasMessage() && c.expected.GetMessage() != actual.GetMessage() {
		return false, nil
	}
	if c.expected.HasLastTransitionTime() && c.expected.GetLastTransitionTime().Sub(actual.GetLastTransitionTime()) > c.expected.timestampTolerance {
		return false, nil
	}
	return true, nil
}

// FailureMessage implements types.GomegaMatcher.
func (c *conditionMatcher[T]) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto equal \n\t%#v", actual, c.expected)
}

// NegatedFailureMessage implements types.GomegaMatcher.
func (c *conditionMatcher[T]) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto not equal \n\t%#v", actual, c.expected)
}

func NewCondition[T comparable]() conditions.Condition[T] {
	return NewConditionImpl[T]()
}

func NewConditionImpl[T comparable]() *ConditionImpl[T] {
	return &ConditionImpl[T]{}
}

func NewConditionImplFromValues[T comparable](conType string, status T, reason, message string, now time.Time) *ConditionImpl[T] {
	return &ConditionImpl[T]{
		status:             &status,
		conType:            &conType,
		reason:             &reason,
		message:            &message,
		lastTransitionTime: &now,
	}
}

func NewConditionImplFromCondition[T comparable](con conditions.Condition[T]) *ConditionImpl[T] {
	return NewConditionImplFromValues(con.GetType(), con.GetStatus(), con.GetReason(), con.GetMessage(), con.GetLastTransitionTime())
}

type ConditionImpl[T comparable] struct {
	status             *T
	conType            *string
	reason             *string
	message            *string
	lastTransitionTime *time.Time
	timestampTolerance time.Duration
}

func (c *ConditionImpl[T]) String() string {
	if c == nil {
		return "<nil>"
	}
	var status, conType, reason, message, lastTransitionTime string
	if c.status == nil {
		status = "<arbitrary>"
	} else {
		status = fmt.Sprintf("%v", *c.status)
	}
	if c.conType == nil {
		conType = "<arbitrary>"
	} else {
		conType = *c.conType
	}
	if c.reason == nil {
		reason = "<arbitrary>"
	} else {
		reason = *c.reason
	}
	if c.message == nil {
		message = "<arbitrary>"
	} else {
		message = *c.message
	}
	if c.lastTransitionTime == nil {
		lastTransitionTime = "<arbitrary>"
	} else {
		lastTransitionTime = c.lastTransitionTime.Format(time.RFC3339)
	}
	return fmt.Sprintf("Condition[%s]{\n\tType: %q,\n\tStatus: %s,\n\tReason: %q,\n\tMessage: %q,\n\tLastTransitionTime: %v,\n}", reflect.TypeFor[T]().Name(), conType, status, reason, message, lastTransitionTime)
}

var _ conditions.Condition[bool] = &ConditionImpl[bool]{}

func (c *ConditionImpl[T]) GetLastTransitionTime() time.Time {
	return *c.lastTransitionTime
}

func (c *ConditionImpl[T]) GetType() string {
	return *c.conType
}

func (c *ConditionImpl[T]) GetStatus() T {
	return *c.status
}

func (c *ConditionImpl[T]) GetReason() string {
	return *c.reason
}

func (c *ConditionImpl[T]) GetMessage() string {
	return *c.message
}

func (c *ConditionImpl[T]) SetStatus(status T) {
	c.status = &status
}

func (c *ConditionImpl[T]) SetType(conType string) {
	c.conType = &conType
}

func (c *ConditionImpl[T]) SetLastTransitionTime(timestamp time.Time) {
	c.lastTransitionTime = &timestamp
}

func (c *ConditionImpl[T]) SetReason(reason string) {
	c.reason = &reason
}

func (c *ConditionImpl[T]) SetMessage(message string) {
	c.message = &message
}

func (c *ConditionImpl[T]) HasLastTransitionTime() bool {
	return c.lastTransitionTime != nil
}

func (c *ConditionImpl[T]) HasType() bool {
	return c.conType != nil
}

func (c *ConditionImpl[T]) HasStatus() bool {
	return c.status != nil
}

func (c *ConditionImpl[T]) HasReason() bool {
	return c.reason != nil
}

func (c *ConditionImpl[T]) HasMessage() bool {
	return c.message != nil
}

func (c *ConditionImpl[T]) WithLastTransitionTime(timestamp time.Time) *ConditionImpl[T] {
	c.SetLastTransitionTime(timestamp)
	return c
}

func (c *ConditionImpl[T]) WithType(conType string) *ConditionImpl[T] {
	c.SetType(conType)
	return c
}

func (c *ConditionImpl[T]) WithStatus(status T) *ConditionImpl[T] {
	c.SetStatus(status)
	return c
}

func (c *ConditionImpl[T]) WithReason(reason string) *ConditionImpl[T] {
	c.SetReason(reason)
	return c
}

func (c *ConditionImpl[T]) WithMessage(message string) *ConditionImpl[T] {
	c.SetMessage(message)
	return c
}

func (c *ConditionImpl[T]) WithTimestampTolerance(t time.Duration) *ConditionImpl[T] {
	c.timestampTolerance = t
	return c
}

func Ptr[T any](v T) *T {
	return &v
}
