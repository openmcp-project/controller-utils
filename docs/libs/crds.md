# Custom Resource Definitions

The `pkg/crds` package allows user to deploy CRDs from yaml files to a target cluster.
A typical use case is to use `embed.FS` to embed the CRDs in the controller binary and deploy them to the target clusters.

The decision on which cluster a CRD should be deployed to is made by the `CRDManager` based on the labels of the CRDs and the labels of the clusters.
The label key is passed to the `CRDManager` when it is created.
Each cluster is then registered with a label value at the `CRDManager`.

## Example

```go
package main

import (
	"context"
	"embed"
	
	"github.com/openmcp-project/controller-utils/pkg/clusters"
	"github.com/openmcp-project/controller-utils/pkg/crds"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

//go:embed crds
var crdsFS embed.FS
var crdsPath = "crds"

func main() {
	ctx := context.Background()

	onboardingCluster := clusters.NewTestClusterFromClient("onboarding", getOnboardingClient())
	workloadCluster := clusters.NewTestClusterFromClient("workload", getWorkloadClient())

	// use "openmcp.cloud/cluster" as the CRD label key
	crdManager := crds.NewCRDManager("openmcp.cloud/cluster", func() ([]*apiextv1.CustomResourceDefinition, error) {
		return crds.CRDsFromFileSystem(crdsFS, crdsPath)
	})
	
	// register the onboarding cluster with label value "onboarding"
	crdManager.AddCRDLabelToClusterMapping("onboarding", onboardingCluster)
	// register the workload cluster with label value "workload"
	crdManager.AddCRDLabelToClusterMapping("workload", workloadCluster)
	
	// create/update the CRDs in all clusters
	err := crdManager.CreateOrUpdateCRDs(ctx, nil)
	if err != nil {
        panic(err)
    }
}
```

The CRDs need to be annotated with the label key and the label value of the cluster they should be deployed to.

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: testresources.example.com
  labels:
    openmcp.cloud/cluster: "onboarding"
...
```
