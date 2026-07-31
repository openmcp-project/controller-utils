package externalsecrets

import (
	"strings"

	"github.com/openmcp-project/controller-utils/pkg/clusters"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ManagedControlPlane indicates that a cluster is a managed control plane.
	ManagedControlPlane ClusterType = "ManagedControlPlane"
	// PlatformCluster indicates that a cluster is a platform cluster.
	PlatformCluster ClusterType = "PlatformCluster"
	// WorkloadCluster indicates that a cluster is a workload cluster.
	WorkloadCluster ClusterType = "WorkloadCluster"
)

// ClusterType distinguishes between managed control plane and workload clusters.
type ClusterType string

// NewManagedCluster creates a new ManagedCluster instance.
func NewManagedCluster(c *clusters.Cluster, cfg *rest.Config, ns string, ct ClusterType) ManagedCluster {
	return &managedCluster{
		cluster:          c,
		cfg:              cfg,
		objects:          []ManagedObject{},
		defaultNamespace: ns,
		clusterType:      ct,
	}
}

// ManagedCluster holds a set of ManagedObjects.
type ManagedCluster interface {
	AddObject(o ManagedObject)
	GetObjects() []ManagedObject
	GetDefaultNamespace() string
	GetHostAndPort() (string, string)
	GetConfig() *rest.Config
	GetClient() client.Client
	GetCluster() *clusters.Cluster
	GetClusterType() ClusterType
}

var _ ManagedCluster = &managedCluster{}

type managedCluster struct {
	cluster          *clusters.Cluster
	cfg              *rest.Config
	objects          []ManagedObject
	defaultNamespace string
	clusterType      ClusterType
}

// GetClient implements ManagedCluster.
func (m *managedCluster) GetClient() client.Client {
	return m.cluster.Client()
}

// GetConfig implements ManagedCluster.
func (m *managedCluster) GetConfig() *rest.Config {
	return m.cfg
}

// GetHostAndPort implements ManagedCluster.
func (m *managedCluster) GetHostAndPort() (string, string) {
	input := strings.TrimPrefix(m.cfg.Host, "https://")
	host, port, found := strings.Cut(input, ":")
	if !found {
		port = "443"
	}
	return host, port
}

// GetDefaultNamespace implements ManagedCluster.
func (m *managedCluster) GetDefaultNamespace() string {
	return m.defaultNamespace
}

// AddObject implements ManagedCluster.
func (m *managedCluster) AddObject(o ManagedObject) {
	m.objects = append(m.objects, o)
}

// GetObjects implements ManagedCluster.
func (m *managedCluster) GetObjects() []ManagedObject {
	return m.objects
}

// GetCluster implements ManagedCluster.
func (m *managedCluster) GetCluster() *clusters.Cluster {
	return m.cluster
}

// GetClusterType implements ManagedCluster.
func (m *managedCluster) GetClusterType() ClusterType {
	return m.clusterType
}
