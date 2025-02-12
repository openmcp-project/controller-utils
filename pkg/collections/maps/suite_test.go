package maps_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIterators(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Maps Test Suite")
}
