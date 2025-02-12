package testing

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"time"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/openmcp-project/controller-utils/pkg/logging"
)

/////////////////
/// CONSTANTS ///
/////////////////

func DefaultScheme() *runtime.Scheme {
	sc := runtime.NewScheme()
	err := clientgoscheme.AddToScheme(sc)
	if err != nil {
		panic(err)
	}
	return sc
}

///////////////////////////
/// COMPLEX ENVIRONMENT ///
///////////////////////////

// ComplexEnvironment helps with testing controllers.
// Construct a new ComplexEnvironment via its builder using NewEnvironmentBuilder().
type ComplexEnvironment struct {
	Ctx         context.Context
	Log         logging.Logger
	Clusters    map[string]client.Client
	Reconcilers map[string]reconcile.Reconciler
}

// Client returns the cluster client for the cluster with the given name.
func (e *ComplexEnvironment) Client(name string) client.Client {
	return e.Clusters[name]
}

// Reconciler returns the reconciler with the given name.
func (e *ComplexEnvironment) Reconciler(name string) reconcile.Reconciler {
	return e.Reconcilers[name]
}

// ShouldReconcile calls the given reconciler with the given request and expects no error.
func (e *ComplexEnvironment) ShouldReconcile(reconciler string, req reconcile.Request, optionalDescription ...interface{}) reconcile.Result {
	return e.shouldReconcile(reconciler, req, optionalDescription...)
}

func (e *ComplexEnvironment) shouldReconcile(reconciler string, req reconcile.Request, optionalDescription ...interface{}) reconcile.Result {
	res, err := e.Reconcilers[reconciler].Reconcile(e.Ctx, req)
	ExpectWithOffset(2, err).ToNot(HaveOccurred(), optionalDescription...)
	return res
}

// ShouldEventuallyReconcile calls the given reconciler with the given request and retries until no error occurred or the timeout is reached.
func (e *ComplexEnvironment) ShouldEventuallyReconcile(reconciler string, req reconcile.Request, timeout, poll time.Duration, optionalDescription ...interface{}) reconcile.Result {
	return e.shouldEventuallyReconcile(reconciler, req, timeout, poll, optionalDescription...)
}

func (e *ComplexEnvironment) shouldEventuallyReconcile(reconciler string, req reconcile.Request, timeout, poll time.Duration, optionalDescription ...interface{}) reconcile.Result {
	var err error
	var res reconcile.Result
	EventuallyWithOffset(1, func() error {
		res, err = e.Reconcilers[reconciler].Reconcile(e.Ctx, req)
		return err
	}, timeout, poll).Should(Succeed(), optionalDescription...)
	return res
}

// ShouldNotReconcile calls the given reconciler with the given request and expects an error.
func (e *ComplexEnvironment) ShouldNotReconcile(reconciler string, req reconcile.Request, optionalDescription ...interface{}) reconcile.Result {
	return e.shouldNotReconcile(reconciler, req, optionalDescription...)
}

func (e *ComplexEnvironment) shouldNotReconcile(reconciler string, req reconcile.Request, optionalDescription ...interface{}) reconcile.Result {
	res, err := e.Reconcilers[reconciler].Reconcile(e.Ctx, req)
	ExpectWithOffset(2, err).To(HaveOccurred(), optionalDescription...)
	return res
}

// ShouldEventuallyNotReconcile calls the given reconciler with the given request and retries until an error occurred or the timeout is reached.
func (e *ComplexEnvironment) ShouldEventuallyNotReconcile(reconciler string, req reconcile.Request, timeout, poll time.Duration, optionalDescription ...interface{}) reconcile.Result {
	return e.shouldEventuallyNotReconcile(reconciler, req, timeout, poll, optionalDescription...)
}

func (e *ComplexEnvironment) shouldEventuallyNotReconcile(reconciler string, req reconcile.Request, timeout, poll time.Duration, optionalDescription ...interface{}) reconcile.Result {
	var err error
	var res reconcile.Result
	EventuallyWithOffset(1, func() error {
		res, err = e.Reconcilers[reconciler].Reconcile(e.Ctx, req)
		return err
	}, timeout, poll).ShouldNot(Succeed(), optionalDescription...)
	return res
}

///////////////////////////////////
/// COMPLEX ENVIRONMENT BUILDER ///
///////////////////////////////////

type ReconcilerConstructorForMultipleClusters func(...client.Client) reconcile.Reconciler

type ComplexEnvironmentBuilder struct {
	internal                *ComplexEnvironment
	Clusters                map[string]*ClusterEnvironment
	Reconcilers             map[string]*ReconcilerEnvironment
	ClusterInitObjects      map[string][]client.Object
	ClusterStatusObjects    map[string][]client.Object
	ClusterInitObjectPaths  map[string][]string
	ClientCreationCallbacks map[string][]func(client.Client)
	loggerIsSet             bool
}

type ClusterEnvironment struct {
	// Client is the client for accessing the cluster.
	Client client.Client
	// Scheme is the scheme used by the client.
	Scheme *runtime.Scheme
	// FakeClientBuilderMethodCalls are the method calls that should be made on the fake.ClientBuilder during client creation.
	FakeClientBuilderMethodCalls []FakeClientBuilderMethodCall
}

type ReconcilerEnvironment struct {
	// Reconciler is the reconciler to be tested.
	// Takes precedence over ReconcilerConstructor.
	Reconciler reconcile.Reconciler
	// ReconcilerConstructor is a function that provides the reconciler to be tested.
	// If the Reconciler field is set, this field is ignored.
	ReconcilerConstructor ReconcilerConstructorForMultipleClusters
	// Targets references the names clusterEnvironments, which represent the clusters that the controller interacts with.
	// Has no effect if the Reconciler is set directly.
	Targets []string
}

// FakeClientBuilderMethodCall represents a method call on a fake.ClientBuilder.
type FakeClientBuilderMethodCall struct {
	// Method is the name of the method that should be called.
	Method string
	// Args are the arguments that should be passed to the method.
	// They will be passed in in the same order as they are listed here.
	Args []any
}

// NewComplexEnvironmentBuilder creates a new EnvironmentBuilder.
func NewComplexEnvironmentBuilder() *ComplexEnvironmentBuilder {
	return &ComplexEnvironmentBuilder{
		internal:                &ComplexEnvironment{},
		Clusters:                map[string]*ClusterEnvironment{},
		Reconcilers:             map[string]*ReconcilerEnvironment{},
		ClusterInitObjects:      map[string][]client.Object{},
		ClusterStatusObjects:    map[string][]client.Object{},
		ClusterInitObjectPaths:  map[string][]string{},
		ClientCreationCallbacks: map[string][]func(client.Client){},
	}
}

// WithContext sets the context for the environment.
func (eb *ComplexEnvironmentBuilder) WithContext(ctx context.Context) *ComplexEnvironmentBuilder {
	eb.internal.Ctx = ctx
	return eb
}

// WithLogger sets the logger for the environment.
// If the context is not set, the logger is injected into a new context.
func (eb *ComplexEnvironmentBuilder) WithLogger(logger logging.Logger) *ComplexEnvironmentBuilder {
	eb.internal.Log = logger
	eb.loggerIsSet = true
	return eb
}

// WithFakeClient sets a fake client for the cluster with the given name.
// If no specific scheme is required, set it to nil or DefaultScheme().
// You should use either WithFakeClient or WithClient for each cluster, but not both.
func (eb *ComplexEnvironmentBuilder) WithFakeClient(name string, scheme *runtime.Scheme) *ComplexEnvironmentBuilder {
	_, ok := eb.Clusters[name]
	if !ok {
		eb.Clusters[name] = &ClusterEnvironment{}
	}
	if scheme == nil {
		scheme = DefaultScheme()
	}
	eb.Clusters[name].Client = nil
	eb.Clusters[name].Scheme = scheme
	return eb
}

// WithClient sets the client for the cluster with the given name.
// You should use either WithFakeClient or WithClient for each cluster, but not both.
func (eb *ComplexEnvironmentBuilder) WithClient(name string, client client.Client) *ComplexEnvironmentBuilder {
	_, ok := eb.Clusters[name]
	if !ok {
		eb.Clusters[name] = &ClusterEnvironment{}
	}
	eb.Clusters[name].Client = client
	eb.Clusters[name].Scheme = client.Scheme()
	return eb
}

// WithInitObjects sets the initial objects for the cluster with the given name.
// If the objects should be loaded from files, use WithInitObjectPath instead.
// If both are specified, the resulting object lists are concatenated.
// Has no effect if the client for the respective cluster is passed in directly.
func (eb *ComplexEnvironmentBuilder) WithInitObjects(name string, objects ...client.Object) *ComplexEnvironmentBuilder {
	eb.ClusterInitObjects[name] = append(eb.ClusterInitObjects[name], objects...)
	return eb
}

// WithDynamicObjectsWithStatus enables the status subresource for the given objects for the cluster with the given name.
// All objects that are created on a cluster during a test (after Build() has been called) and where interaction with the object's status is desired must be added here,
// otherwise the fake client will return that no status was found for the object.
func (eb *ComplexEnvironmentBuilder) WithDynamicObjectsWithStatus(name string, objects ...client.Object) *ComplexEnvironmentBuilder {
	eb.ClusterStatusObjects[name] = append(eb.ClusterStatusObjects[name], objects...)
	return eb
}

// WithInitObjectPath adds a path to the list of paths from which to load initial objects for the cluster with the given name.
// If the objects should be specified directly, use WithInitObjects instead.
// If both are specified, the resulting object lists are concatenated.
// Note that this function concatenates all arguments into a single path, if you want to load files from multiple paths, call this function multiple times.
// Has no effect if the client for the respective cluster is passed in directly.
func (eb *ComplexEnvironmentBuilder) WithInitObjectPath(name string, pathSegments ...string) *ComplexEnvironmentBuilder {
	eb.ClusterInitObjectPaths[name] = append(eb.ClusterInitObjectPaths[name], path.Join(pathSegments...))
	return eb
}

// WithReconciler sets the reconciler for the given name.
func (eb *ComplexEnvironmentBuilder) WithReconciler(name string, reconciler reconcile.Reconciler) *ComplexEnvironmentBuilder {
	_, ok := eb.Reconcilers[name]
	if !ok {
		eb.Reconcilers[name] = &ReconcilerEnvironment{}
	}
	eb.Reconcilers[name].Reconciler = reconciler
	return eb
}

// WithReconcilerConstructor sets the constructor for the Reconciler for the given name.
// The Reconciler will be constructed during Build() from the given function and the clients retrieved from the passed in targets list.
// The clients are passed in the function in the same order as they are listed in the targets list.
func (eb *ComplexEnvironmentBuilder) WithReconcilerConstructor(name string, constructor ReconcilerConstructorForMultipleClusters, targets ...string) *ComplexEnvironmentBuilder {
	_, ok := eb.Reconcilers[name]
	if !ok {
		eb.Reconcilers[name] = &ReconcilerEnvironment{}
	}
	eb.Reconcilers[name].ReconcilerConstructor = constructor
	eb.Reconcilers[name].Targets = targets
	return eb
}

// WithAfterClientCreationCallback adds a callback function that will be called with the client with the given name as argument after the client has been created (during Build()).
func (eb *ComplexEnvironmentBuilder) WithAfterClientCreationCallback(name string, callback func(client.Client)) *ComplexEnvironmentBuilder {
	eb.ClientCreationCallbacks[name] = append(eb.ClientCreationCallbacks[name], callback)
	return eb
}

// WithFakeClientBuilderCall allows to inject method calls to fake.ClientBuilder when the fake clients are created during Build().
// The fake clients are usually created using WithScheme(...).WithObjects(...).WithStatusSubresource(...).Build().
// This function allows to inject additional method calls. It is only required for advanced use-cases.
// The method calls are executed using reflection, so take care to not make any mistakes with the spelling of the method name or the order or type of the arguments.
// Has no effect if the client for the respective cluster is passed in directly (and thus no fake client is constructed).
func (eb *ComplexEnvironmentBuilder) WithFakeClientBuilderCall(name string, method string, args ...any) *ComplexEnvironmentBuilder {
	_, ok := eb.Clusters[name]
	if !ok {
		eb.Clusters[name] = &ClusterEnvironment{}
	}
	eb.Clusters[name].FakeClientBuilderMethodCalls = append(eb.Clusters[name].FakeClientBuilderMethodCalls, FakeClientBuilderMethodCall{
		Method: method,
		Args:   args,
	})
	return eb
}

// Build constructs the environment from the builder.
// Note that this function panics instead of throwing an error,
// as it is intended to be used in tests, where all information is static anyway.
func (eb *ComplexEnvironmentBuilder) Build() *ComplexEnvironment {
	res := eb.internal

	// initialize logger
	if !eb.loggerIsSet {
		log, err := logging.GetLogger()
		if err != nil {
			panic(fmt.Errorf("error getting logger: %w", err))
		}
		res.Log = log
	}
	ctrl.SetLogger(res.Log.Logr())

	// initialize context
	if res.Ctx == nil {
		res.Ctx = logging.NewContext(context.Background(), res.Log)
	}

	// initialize clusters
	if res.Clusters == nil {
		res.Clusters = map[string]client.Client{}
	}
	for name, ce := range eb.Clusters {
		if ce == nil {
			panic(fmt.Errorf("no ClusterEnvironment set for cluster '%s'", name))
		}
		if ce.Client != nil {
			if ce.Scheme == nil {
				// infer scheme from client
				ce.Scheme = ce.Client.Scheme()
			}
		} else {
			if ce.Scheme == nil {
				ce.Scheme = DefaultScheme()
			}
			// create fake client
			fcb := fake.NewClientBuilder().WithScheme(ce.Scheme)
			objs := []client.Object{}
			if len(eb.ClusterInitObjectPaths) > 0 {
				// load objects from paths
				for _, p := range eb.ClusterInitObjectPaths[name] {
					objects, err := LoadObjects(p, ce.Scheme)
					if err != nil {
						panic(fmt.Errorf("error loading objects for cluster '%s' from path '%s': %w", name, p, err))
					}
					objs = append(objs, objects...)
				}
			}
			if len(eb.ClusterInitObjects) > 0 {
				objs = append(objs, eb.ClusterInitObjects[name]...)
			}
			statusObjs := []client.Object{}
			statusObjs = append(statusObjs, objs...)
			statusObjs = append(statusObjs, eb.ClusterStatusObjects[name]...)
			fcb.WithObjects(objs...).WithStatusSubresource(statusObjs...)
			for _, call := range ce.FakeClientBuilderMethodCalls {
				method := reflect.ValueOf(fcb).MethodByName(call.Method)
				if !method.IsValid() {
					panic(fmt.Errorf("method '%s' not found on fake.ClientBuilder", call.Method))
				}
				args := make([]reflect.Value, len(call.Args))
				for i, arg := range call.Args {
					args[i] = reflect.ValueOf(arg)
				}
				method.Call(args)
			}
			ce.Client = fcb.Build()
		}
		res.Clusters[name] = ce.Client
	}

	// initialize reconcilers
	if res.Reconcilers == nil {
		res.Reconcilers = map[string]reconcile.Reconciler{}
	}
	for name, re := range eb.Reconcilers {
		if re == nil {
			continue
		}
		if re.Reconciler == nil {
			if re.ReconcilerConstructor == nil {
				panic(fmt.Errorf("no ReconcilerConstructor set for reconciler '%s'", name))
			}
			if len(re.Targets) == 0 {
				panic(fmt.Errorf("no target cluster set for reconciler '%s'", name))
			}
			targets := make([]client.Client, len(re.Targets))
			for i, target := range re.Targets {
				var ok bool
				targets[i], ok = res.Clusters[target]
				if !ok {
					panic(fmt.Errorf("unknown target cluster '%s' specified for reconciler '%s'", target, name))
				}
			}
			re.Reconciler = re.ReconcilerConstructor(targets...)
		}
		res.Reconcilers[name] = re.Reconciler
	}

	// call client creation callbacks
	for name, callbacks := range eb.ClientCreationCallbacks {
		client, ok := res.Clusters[name]
		if !ok {
			panic(fmt.Errorf("no client found for cluster '%s' to call client creation callbacks", name))
		}
		for _, callback := range callbacks {
			callback(client)
		}
	}

	return res
}
