package crds_test

import (
	"context"
	"embed"
	"testing"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/openmcp-project/controller-utils/pkg/clusters"
	utilstest "github.com/openmcp-project/controller-utils/pkg/testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/crds"
)

//go:embed testdata/*
var testFS embed.FS

func TestCRDs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CRDs Test Suite")
}

var _ = Describe("CRDsFromFileSystem", func() {
	It("should correctly read and parse CRDs from the filesystem", func() {
		crdPath := "testdata"
		crdsList, err := crds.CRDsFromFileSystem(testFS, crdPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(crdsList).To(HaveLen(3))

		// Validate the first CRD
		Expect(crdsList[0].Name).To(Equal("testresources.example.com"))
		Expect(crdsList[0].Spec.Names.Kind).To(Equal("TestResource"))

		// Validate the second CRD
		Expect(crdsList[1].Name).To(Equal("sampleresources.example.com"))
		Expect(crdsList[1].Spec.Names.Kind).To(Equal("SampleResource"))
	})
})

var _ = Describe("CRDManager", func() {
	It("should correctly manage CRD mappings and create/update CRDs", func() {
		scheme := runtime.NewScheme()
		err := apiextv1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())

		// Create fake clients
		clientA, err := utilstest.GetFakeClient(scheme)
		Expect(err).NotTo(HaveOccurred())

		clientB, err := utilstest.GetFakeClient(scheme)
		Expect(err).NotTo(HaveOccurred())

		// Create fake clusters
		clusterA := clusters.NewTestClusterFromClient("cluster_a", clientA)
		clusterB := clusters.NewTestClusterFromClient("cluster_b", clientB)

		crdManager := crds.NewCRDManager("openmcp.cloud/cluster", func() ([]*apiextv1.CustomResourceDefinition, error) {
			return crds.CRDsFromFileSystem(testFS, "testdata")
		})

		crdManager.AddCRDLabelToClusterMapping("cluster_a", clusterA)
		crdManager.AddCRDLabelToClusterMapping("cluster_b", clusterB)

		ctx := context.Background()

		// CRD creation should fail due to unknown cluster label
		Expect(crdManager.CreateOrUpdateCRDs(ctx, nil)).To(HaveOccurred())

		crdManager.SkipCRDsWithClusterLabel("cluster_c")

		err = crdManager.CreateOrUpdateCRDs(ctx, nil)
		Expect(err).NotTo(HaveOccurred())

		// Verify that the CRDs were created in the respective clusters
		crdA := &apiextv1.CustomResourceDefinition{}
		err = clientA.Get(ctx, types.NamespacedName{Name: "testresources.example.com"}, crdA)
		Expect(err).NotTo(HaveOccurred())
		Expect(crdA.Labels["openmcp.cloud/cluster"]).To(Equal("cluster_a"))

		crdB := &apiextv1.CustomResourceDefinition{}
		err = clientB.Get(ctx, types.NamespacedName{Name: "sampleresources.example.com"}, crdB)
		Expect(err).NotTo(HaveOccurred())
		Expect(crdB.Labels["openmcp.cloud/cluster"]).To(Equal("cluster_b"))
	})
})
