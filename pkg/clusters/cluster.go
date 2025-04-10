package clusters

import (
	"fmt"

	flag "github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"

	"github.com/openmcp-project/controller-utils/pkg/controller"
)

type Cluster struct {
	// identifier (for logging purposes only)
	id string
	// path to kubeconfig
	cfgPath string
	// cluster config
	restCfg *rest.Config
	// client
	client client.Client
	// cluster
	cluster cluster.Cluster
}

// Initializes a new cluster.
// Panics if id is empty.
func New(id string) *Cluster {
	c := &Cluster{}
	c.InitializeID(id)
	return c
}

// WithConfigPath sets the config path for the cluster.
// Returns the cluster for chaining.
func (c *Cluster) WithConfigPath(cfgPath string) *Cluster {
	c.cfgPath = cfgPath
	return c
}

// RegisterConfigPathFlag adds a flag '--<id>-cluster' for the cluster's config path to the given flag set.
// Panics if the cluster's id is not set.
func (c *Cluster) RegisterConfigPathFlag(flags *flag.FlagSet) {
	if !c.HasID() {
		panic("cluster id must be set before registering the config path flag")
	}
	flags.StringVar(&c.cfgPath, fmt.Sprintf("%s-cluster", c.id), "", fmt.Sprintf("Path to the %s cluster kubeconfig file or directory containing either a kubeconfig or host, token, and ca file. Leave empty to use in-cluster config.", c.id))
}

///////////////////
// STATUS CHECKS //
///////////////////

// HasID returns true if the cluster has an id.
// If this returns false, initialize a new cluster via New() or InitializeID().
func (c *Cluster) HasID() bool {
	return c != nil && c.id != ""
}

// HasRESTConfig returns true if the cluster has a REST config.
// If this returns false, load the config via InitializeRESTConfig().
func (c *Cluster) HasRESTConfig() bool {
	return c != nil && c.restCfg != nil
}

// HasClient returns true if the cluster has a client.
// If this returns false, create a client via InitializeClient().
func (c *Cluster) HasClient() bool {
	return c != nil && c.client != nil
}

//////////////////
// INITIALIZERS //
//////////////////

// InitializeID sets the cluster's id.
// Panics if id is empty.
func (c *Cluster) InitializeID(id string) {
	if id == "" {
		panic("id must not be empty")
	}
	c.id = id
}

// InitializeRESTConfig loads the cluster's REST config.
// If the config has already been loaded, this is a no-op.
// Panics if the cluster's id is not set (InitializeID must be called first).
func (c *Cluster) InitializeRESTConfig() error {
	if !c.HasID() {
		panic("cluster id must be set before loading the config")
	}
	if c.HasRESTConfig() {
		return nil
	}
	cfg, err := controller.LoadKubeconfig(c.cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load '%s' cluster kubeconfig: %w", c.ID(), err)
	}
	c.restCfg = cfg
	return nil
}

// InitializeClient creates a new client for the cluster.
// This also initializes the cluster's controller-runtime 'Cluster' representation.
// If the client has already been initialized, this is a no-op.
// Panics if the cluster's REST config has not been loaded (InitializeRESTConfig must be called first).
func (c *Cluster) InitializeClient(scheme *runtime.Scheme) error {
	if !c.HasRESTConfig() {
		panic("cluster REST config must be set before creating the client")
	}
	if c.HasClient() {
		return nil
	}
	cli, err := client.New(c.restCfg, client.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("failed to create '%s' cluster client: %w", c.ID(), err)
	}
	clu, err := cluster.New(c.restCfg, func(o *cluster.Options) { o.Scheme = scheme })
	if err != nil {
		return fmt.Errorf("failed to create '%s' cluster Cluster representation: %w", c.ID(), err)
	}
	c.client = cli
	c.cluster = clu
	return nil
}

/////////////
// GETTERS //
/////////////

// ID returns the cluster's id.
func (c *Cluster) ID() string {
	return c.id
}

// ConfigPath returns the cluster's config path.
func (c *Cluster) ConfigPath() string {
	return c.cfgPath
}

// RESTConfig returns the cluster's REST config.
// This returns a pointer, but modification can lead to inconsistent behavior and is not recommended.
func (c *Cluster) RESTConfig() *rest.Config {
	return c.restCfg
}

// Client returns the cluster's client.
func (c *Cluster) Client() client.Client {
	return c.client
}

// Cluster returns the cluster's controller-runtime 'Cluster' representation.
func (c *Cluster) Cluster() cluster.Cluster {
	return c.cluster
}

// Scheme returns the cluster's scheme.
// Returns nil if the client has not been initialized.
func (c *Cluster) Scheme() *runtime.Scheme {
	if c.cluster == nil {
		return nil
	}
	return c.cluster.GetScheme()
}

// APIServerEndpoint returns the cluster's API server endpoint.
// Returns an empty string if the REST config has not been initialized.
func (c *Cluster) APIServerEndpoint() string {
	if c.restCfg == nil {
		return ""
	}
	return c.restCfg.Host
}
