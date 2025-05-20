package clusteraccess

import (
	"slices"
	"strings"
)

// L is a constructor for Label.
func L(key, value string) Label {
	return Label{
		Key:   key,
		Value: value,
	}
}

// Label represents a metadata label.
// It can be constrcuted using the L() function.
type Label struct {
	Key   string
	Value string
}

// Compare returns -1 if l < other, 0 if l == other, and 1 if l > other.
// It is used for sorting Labels alphabetically by their keys.
func (l *Label) Compare(other Label) int {
	return strings.Compare(l.Key, other.Key)
}

// SortLabels sorts a list of Labels alphabetically by their keys.
// The sort happens in-place.
// It is not stable, but since keys are expected to be unique, that should not matter.
func SortLabels(labels []Label) {
	slices.SortFunc(labels, func(a, b Label) int {
		return a.Compare(b)
	})
}

// LabelMapToList converts a map[string]string to a list of Labels.
// Note that the order of of the list is arbitrary.
func LabelMapToList(labels map[string]string) []Label {
	res := make([]Label, 0, len(labels))
	for k, v := range labels {
		res = append(res, Label{Key: k, Value: v})
	}
	return res
}

// LabelListToMap converts a list of Labels to a map[string]string.
func LabelListToMap(labels []Label) map[string]string {
	res := make(map[string]string, len(labels))
	for _, l := range labels {
		res[l.Key] = l.Value
	}
	return res
}
