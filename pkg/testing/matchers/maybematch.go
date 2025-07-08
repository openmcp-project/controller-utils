package matchers

import (
	"fmt"

	"github.com/onsi/gomega/types"
)

// MaybeMatch returns a Gomega matcher that passes the matching logic to the provided matcher,
// but always succeeds if the passed in matcher is nil.
func MaybeMatch(matcher types.GomegaMatcher) types.GomegaMatcher {
	return &maybeMatcher{matcher: matcher}
}

type maybeMatcher struct {
	matcher types.GomegaMatcher
}

func (m *maybeMatcher) GomegaString() string {
	if m == nil || m.matcher == nil {
		return "<nil>"
	}
	return fmt.Sprintf("MaybeMatch(%v)", m.matcher)
}

var _ types.GomegaMatcher = &maybeMatcher{}

// Match implements types.GomegaMatcher.
func (m *maybeMatcher) Match(actualRaw any) (success bool, err error) {
	if m.matcher == nil {
		return true, nil
	}
	return m.matcher.Match(actualRaw)
}

// FailureMessage implements types.GomegaMatcher.
func (m *maybeMatcher) FailureMessage(actual any) (message string) {
	return m.matcher.FailureMessage(actual)
}

// NegatedFailureMessage implements types.GomegaMatcher.
func (m *maybeMatcher) NegatedFailureMessage(actual any) (message string) {
	return m.matcher.NegatedFailureMessage(actual)
}
