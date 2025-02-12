package collections_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var baseData = []int{1, 3, 2, 4}

func TestCollections(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Collection Test Suite")
}
