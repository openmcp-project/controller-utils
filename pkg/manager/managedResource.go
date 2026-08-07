package manager

// ManagedResource is the read-only view of a reconciled object returned by Apply and Delete.
// It is an interface so pkg/manager stays decoupled from any CRD schema: consumers define
// their own concrete CRD-embedded type and populate it from this interface.
type ManagedResource interface {
	GetAPIVersion() string
	GetKind() string
	GetName() string
	GetNamespace() *string // nil when the namespace is empty
	GetLocation() ClusterType
	GetStatus() ManagedResourceStatus
}

// managedResourceResult is the package-private concrete implementation of ManagedResource
// built inside resultsToResources. It never escapes pkg/manager.
type managedResourceResult struct {
	apiVersion string
	kind       string
	name       string
	namespace  *string
	location   ClusterType
	status     ManagedResourceStatus
}

func (r *managedResourceResult) GetAPIVersion() string    { return r.apiVersion }
func (r *managedResourceResult) GetKind() string          { return r.kind }
func (r *managedResourceResult) GetName() string          { return r.name }
func (r *managedResourceResult) GetNamespace() *string    { return r.namespace }
func (r *managedResourceResult) GetLocation() ClusterType { return r.location }
func (r *managedResourceResult) GetStatus() ManagedResourceStatus { return r.status }
