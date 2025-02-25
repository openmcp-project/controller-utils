package testing

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/openmcp-project/controller-utils/pkg/logging"
)

const (
	SimpleEnvironmentDefaultKey = "default"
)

//////////////////////////
/// SIMPLE ENVIRONMENT ///
//////////////////////////

// Environment is a wrapper around ComplexEnvironment.
// It is meant to ease the usage for simple use-cases (meaning only one reconciler and only one cluster).
// Use the EnvironmentBuilder to construct a new Environment.
type Environment struct {
	*ComplexEnvironment
}

// Client returns the client for the cluster.
func (e *Environment) Client() client.Client {
	return e.ComplexEnvironment.Client(SimpleEnvironmentDefaultKey)
}

// Reconciler returns the reconciler.
func (e *Environment) Reconciler() reconcile.Reconciler {
	return e.ComplexEnvironment.Reconciler(SimpleEnvironmentDefaultKey)
}

// ShouldReconcile calls the given reconciler with the given request and expects no error.
func (e *Environment) ShouldReconcile(req reconcile.Request, optionalDescription ...interface{}) reconcile.Result {
	return e.shouldReconcile(SimpleEnvironmentDefaultKey, req, optionalDescription...)
}

// ShouldEventuallyReconcile calls the given reconciler with the given request and retries until no error occurred or the timeout is reached.
func (e *Environment) ShouldEventuallyReconcile(req reconcile.Request, timeout, poll time.Duration, optionalDescription ...interface{}) reconcile.Result {
	return e.shouldEventuallyReconcile(SimpleEnvironmentDefaultKey, req, timeout, poll, optionalDescription...)
}

// ShouldNotReconcile calls the given reconciler with the given request and expects an error.
func (e *Environment) ShouldNotReconcile(req reconcile.Request, optionalDescription ...interface{}) reconcile.Result {
	return e.shouldNotReconcile(SimpleEnvironmentDefaultKey, req, optionalDescription...)
}

// ShouldEventuallyNotReconcile calls the given reconciler with the given request and retries until an error occurred or the timeout is reached.
func (e *Environment) ShouldEventuallyNotReconcile(req reconcile.Request, timeout, poll time.Duration, optionalDescription ...interface{}) reconcile.Result {
	return e.shouldEventuallyNotReconcile(SimpleEnvironmentDefaultKey, req, timeout, poll, optionalDescription...)
}

//////////////////////////////////
/// SIMPLE ENVIRONMENT BUILDER ///
//////////////////////////////////

type ReconcilerConstructor func(client.Client) reconcile.Reconciler

type EnvironmentBuilder struct {
	*ComplexEnvironmentBuilder
}

// NewEnvironmentBuilder creates a new SimpleEnvironmentBuilder.
// Use this to construct a SimpleEnvironment, if you need only one reconciler and one cluster.
// For more complex test scenarios, use NewEnvironmentBuilder() instead.
func NewEnvironmentBuilder() *EnvironmentBuilder {
	res := &EnvironmentBuilder{
		ComplexEnvironmentBuilder: NewComplexEnvironmentBuilder(),
	}
	res.WithFakeClient(nil)
	return res
}

// WithContext sets the context for the environment.
func (eb *EnvironmentBuilder) WithContext(ctx context.Context) *EnvironmentBuilder {
	eb.ComplexEnvironmentBuilder.WithContext(ctx)
	return eb
}

// WithLogger sets the logger for the environment.
// If the context is not set, the logger is injected into a new context.
func (eb *EnvironmentBuilder) WithLogger(log logging.Logger) *EnvironmentBuilder {
	eb.ComplexEnvironmentBuilder.WithLogger(log)
	return eb
}

// WithFakeClient requests a fake client.
// If no specific scheme is required, set it to nil or DefaultScheme().
// You should use either WithFakeClient or WithClient, not both.
func (eb *EnvironmentBuilder) WithFakeClient(scheme *runtime.Scheme) *EnvironmentBuilder {
	eb.ComplexEnvironmentBuilder.WithFakeClient(SimpleEnvironmentDefaultKey, scheme)
	return eb
}

// WithClient sets the client for the cluster.
// If not called, a fake client will be constructed.
func (eb *EnvironmentBuilder) WithClient(client client.Client) *EnvironmentBuilder {
	eb.ComplexEnvironmentBuilder.WithClient(SimpleEnvironmentDefaultKey, client)
	return eb
}

// WithInitObjects sets the initial objects for the cluster.
// If the objects should be loaded from files, use WithInitObjectPath instead.
// If both are specified, the resulting object lists are concatenated.
// Has no effect if the client for the cluster is set directly.
func (eb *EnvironmentBuilder) WithInitObjects(objects ...client.Object) *EnvironmentBuilder {
	eb.ComplexEnvironmentBuilder.WithInitObjects(SimpleEnvironmentDefaultKey, objects...)
	return eb
}

// WithDynamicObjectsWithStatus enables the status subresource for the given objects.
// All objects that are created on a cluster during a test (after Build() has been called) and where interaction with the object's status is desired must be added here,
// otherwise the fake client will return that no status was found for the object.
func (eb *EnvironmentBuilder) WithDynamicObjectsWithStatus(objects ...client.Object) *EnvironmentBuilder {
	eb.ComplexEnvironmentBuilder.WithDynamicObjectsWithStatus(SimpleEnvironmentDefaultKey, objects...)
	return eb
}

// WithInitObjectPath adds a path to the list of paths from which to load initial objects for the cluster.
// If the objects should be specified directly, use WithInitObjects instead.
// If both are specified, the resulting object lists are concatenated.
// Note that this function concatenates all arguments into a single path, if you want to load files from multiple paths, call this function multiple times.
// Has no effect if the client for the cluster is set directly.
func (eb *EnvironmentBuilder) WithInitObjectPath(pathSegments ...string) *EnvironmentBuilder {
	eb.ComplexEnvironmentBuilder.WithInitObjectPath(SimpleEnvironmentDefaultKey, pathSegments...)
	return eb
}

// WithReconciler sets the reconciler.
// Takes precedence over WithReconcilerConstructor, if both are called.
func (eb *EnvironmentBuilder) WithReconciler(reconciler reconcile.Reconciler) *EnvironmentBuilder {
	eb.ComplexEnvironmentBuilder.WithReconciler(SimpleEnvironmentDefaultKey, reconciler)
	return eb
}

// WithReconcilerConstructor sets the constructor for the Reconciler.
// The Reconciler will be constructed during Build() using the client from this constructor.
// No effect if the reconciler is set directly via WithReconciler.
func (eb *EnvironmentBuilder) WithReconcilerConstructor(constructor ReconcilerConstructor) *EnvironmentBuilder {
	eb.ComplexEnvironmentBuilder.WithReconcilerConstructor(SimpleEnvironmentDefaultKey, func(c ...client.Client) reconcile.Reconciler { return constructor(c[0]) }, SimpleEnvironmentDefaultKey)
	return eb
}

// WithAfterClientCreationCallback adds a callback function that will be called with the client as argument after it has been created (during Build()).
func (eb *EnvironmentBuilder) WithAfterClientCreationCallback(callback func(client.Client)) *EnvironmentBuilder {
	eb.ComplexEnvironmentBuilder.WithAfterClientCreationCallback(SimpleEnvironmentDefaultKey, callback)
	return eb
}

// WithFakeClientBuilderCall allows to inject method calls to fake.ClientBuilder when the fake client is created during Build().
// The fake client is usually created using WithScheme(...).WithObjects(...).WithStatusSubresource(...).Build().
// This function allows to inject additional method calls. It is only required for advanced use-cases.
// The method calls are executed using reflection, so take care to not make any mistakes with the spelling of the method name or the order or type of the arguments.
// Has no effect if the client is passed in directly (and thus no fake client is constructed).
func (eb *EnvironmentBuilder) WithFakeClientBuilderCall(method string, args ...any) *EnvironmentBuilder {
	eb.ComplexEnvironmentBuilder.WithFakeClientBuilderCall(SimpleEnvironmentDefaultKey, method, args...)
	return eb
}

// Build constructs the environment from the builder.
// Note that this function panics instead of throwing an error,
// as it is intended to be used in tests, where all information is static anyway.
func (eb *EnvironmentBuilder) Build() *Environment {
	return &Environment{
		ComplexEnvironment: eb.ComplexEnvironmentBuilder.Build(),
	}
}
