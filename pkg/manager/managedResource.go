package manager

// ManagedResource is the read-only view of a reconciled object returned by Apply and Delete.
// It is an interface so pkg/manager stays decoupled from any CRD schema: consumers define
// their own concrete CRD-embedded type and populate it from this interface.
type ManagedResource interface {
	GetAPIGroup() string
	GetKind() string
	GetName() string
	GetNamespace() string
	GetGeneration() int64
	GetLocation() ClusterType
	GetStatus() ManagedResourceStatus
}

// managedResourceResult is the package-private concrete implementation of ManagedResource
// built inside resultsToResources. It never escapes pkg/manager.
type managedResourceResult struct {
	apiGroup   string
	kind       string
	name       string
	namespace  string
	generation int64
	location   ClusterType
	status     ManagedResourceStatus
}

func (r *managedResourceResult) GetAPIGroup() string              { return r.apiGroup }
func (r *managedResourceResult) GetKind() string                  { return r.kind }
func (r *managedResourceResult) GetName() string                  { return r.name }
func (r *managedResourceResult) GetNamespace() string             { return r.namespace }
func (r *managedResourceResult) GetGeneration() int64             { return r.generation }
func (r *managedResourceResult) GetLocation() ClusterType         { return r.location }
func (r *managedResourceResult) GetStatus() ManagedResourceStatus { return r.status }
