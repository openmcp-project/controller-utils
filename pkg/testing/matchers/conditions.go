package matchers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/onsi/gomega/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MatchCondition returns a Gomega matcher that checks if a Condition is equal to the expected one.
// If the passed in 'actual' is not a Condition, the matcher will fail.
// All fields which are set to their zero value in the expected condition will be ignored.
// Use one of the TestCondition... constructors to create a Condition that can be used with this matcher.
func MatchCondition(con *Condition) types.GomegaMatcher {
	return &conditionMatcher{expected: con}
}

type conditionMatcher struct {
	expected *Condition
}

func (c *conditionMatcher) GomegaString() string {
	if c == nil || c.expected == nil {
		return "<nil>"
	}
	return c.expected.String()
}

var _ types.GomegaMatcher = &conditionMatcher{}

// Match implements types.GomegaMatcher.
func (c *conditionMatcher) Match(actualRaw any) (success bool, err error) {
	if actualRaw == nil {
		return c.expected.Matches(nil), nil
	}
	switch actual := actualRaw.(type) {
	case *Condition:
		return c.expected.Matches(actual), nil
	case Condition:
		return c.expected.Matches(&actual), nil
	case *metav1.Condition:
		return c.expected.Matches(TestConditionFromCondition(*actual)), nil
	case metav1.Condition:
		return c.expected.Matches(TestConditionFromCondition(actual)), nil
	default:
		return false, fmt.Errorf("expected actual (or &actual) to be of type Condition or metav1.Condition, got %T", actualRaw)
	}
}

// FailureMessage implements types.GomegaMatcher.
func (c *conditionMatcher) FailureMessage(actual any) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto equal \n\t%#v", actual, c.expected)
}

// NegatedFailureMessage implements types.GomegaMatcher.
func (c *conditionMatcher) NegatedFailureMessage(actual any) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto not equal \n\t%#v", actual, c.expected)
}

type Condition struct {
	status             *metav1.ConditionStatus
	conType            *string
	observedGeneration *int64
	reason             *string
	message            *string
	lastTransitionTime *metav1.Time
	timestampTolerance time.Duration
}

func (c *Condition) String() string {
	if c == nil {
		return "<nil>"
	}
	var status, conType, observedGeneration, reason, message, lastTransitionTime string
	if c.status == nil {
		status = "<arbitrary>"
	} else {
		status = string(*c.status)
	}
	if c.conType == nil {
		conType = "<arbitrary>"
	} else {
		conType = *c.conType
	}
	if c.observedGeneration == nil {
		observedGeneration = "<arbitrary>"
	} else {
		observedGeneration = strconv.FormatInt(*c.observedGeneration, 10)
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
	return fmt.Sprintf("Condition{\n\tType: %q,\n\tStatus: %s,\n\tObservedGeneration: %s,\n\tReason: %q,\n\tMessage: %q,\n\tLastTransitionTime: %v,\n}", conType, status, observedGeneration, reason, message, lastTransitionTime)
}

func TestCondition() *Condition {
	return &Condition{}
}

func TestConditionFromCondition(con metav1.Condition) *Condition {
	return TestConditionFromValues(con.Type, con.Status, con.ObservedGeneration, con.Reason, con.Message, con.LastTransitionTime)
}

func TestConditionFromValues(conType string, status metav1.ConditionStatus, observedGeneration int64, reason, message string, lastTransitionTime metav1.Time) *Condition {
	return TestCondition().
		WithType(conType).
		WithStatus(status).
		WithObservedGeneration(observedGeneration).
		WithReason(reason).
		WithMessage(message).
		WithLastTransitionTime(lastTransitionTime)
}

func (c *Condition) ToCondition() metav1.Condition {
	if c == nil {
		return metav1.Condition{}
	}
	res := metav1.Condition{}
	if c.conType != nil {
		res.Type = *c.conType
	}
	if c.status != nil {
		res.Status = *c.status
	} else {
		res.Status = metav1.ConditionUnknown
	}
	if c.observedGeneration != nil {
		res.ObservedGeneration = *c.observedGeneration
	}
	if c.reason != nil {
		res.Reason = *c.reason
	}
	if c.message != nil {
		res.Message = *c.message
	}
	if c.lastTransitionTime != nil {
		res.LastTransitionTime = *c.lastTransitionTime
	}
	return res
}

func (c *Condition) WithType(conType string) *Condition {
	cp := conType
	c.conType = &cp
	return c
}

func (c *Condition) WithStatus(status metav1.ConditionStatus) *Condition {
	cp := status
	c.status = &cp
	return c
}

func (c *Condition) WithObservedGeneration(observedGeneration int64) *Condition {
	cp := observedGeneration
	c.observedGeneration = &cp
	return c
}

func (c *Condition) WithReason(reason string) *Condition {
	cp := reason
	c.reason = &cp
	return c
}

func (c *Condition) WithMessage(message string) *Condition {
	cp := message
	c.message = &cp
	return c
}

func (c *Condition) WithLastTransitionTime(timestamp metav1.Time) *Condition {
	cp := timestamp
	c.lastTransitionTime = &cp
	return c
}

func (c *Condition) WithTimestampTolerance(t time.Duration) *Condition {
	c.timestampTolerance = t
	return c
}

// Matches checks if the current Condition matches the other Condition.
// Note that this method is not symmetrical, meaning that a.Matches(b) may yield a different result than b.Matches(a).
// The reason for this is that nil fields in the receiver Condition are considered "arbitrary" and will match any value in the other Condition.
// If both Conditions are nil, they match.
func (c *Condition) Matches(other *Condition) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil || other == nil {
		return false
	}

	if c.conType != nil && (other.conType == nil || *c.conType != *other.conType) {
		return false
	}
	if c.status != nil && (other.status == nil || *c.status != *other.status) {
		return false
	}
	if c.observedGeneration != nil && (other.observedGeneration == nil || *c.observedGeneration != *other.observedGeneration) {
		return false
	}
	if c.reason != nil && (other.reason == nil || *c.reason != *other.reason) {
		return false
	}
	if c.message != nil && (other.message == nil || *c.message != *other.message) {
		return false
	}
	if c.lastTransitionTime != nil && (other.lastTransitionTime == nil || c.lastTransitionTime.Sub(other.lastTransitionTime.Time) > c.timestampTolerance) {
		return false
	}

	return true
}
