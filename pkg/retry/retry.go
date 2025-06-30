package retry

import (
	"context"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Client struct {
	internal          client.Client
	interval          time.Duration
	backoffMultiplier float64
	maxAttempts       int
	timeout           time.Duration
}

// NewRetryingClient returns a retry.Client that implements client.Client, but retries each operation that can fail with the specified parameters.
// Returns nil if the provided client is nil.
// The default parameters are:
// - interval: 100 milliseconds
// - backoffMultiplier: 1.0 (no backoff)
// - maxAttempts: 0 (no limit on attempts)
// - timeout: 1 second (timeout for retries)
// Use the builder-style With... methods to adapt the parameters.
func NewRetryingClient(c client.Client) *Client {
	if c == nil {
		return nil
	}
	return &Client{
		internal:          c,
		interval:          100 * time.Millisecond, // default retry interval
		backoffMultiplier: 1.0,                    // default backoff multiplier
		maxAttempts:       0,                      // default max retries
		timeout:           1 * time.Second,        // default timeout for retries
	}
}

var _ client.Client = &Client{}

/////////////
// GETTERS //
/////////////

// Interval returns the configured retry interval.
func (rc *Client) Interval() time.Duration {
	return rc.interval
}

// BackoffMultiplier returns the configured backoff multiplier for retries.
func (rc *Client) BackoffMultiplier() float64 {
	return rc.backoffMultiplier
}

// MaxRetries returns the configured maximum number of retries.
func (rc *Client) MaxRetries() int {
	return rc.maxAttempts
}

// Timeout returns the configured timeout for retries.
func (rc *Client) Timeout() time.Duration {
	return rc.timeout
}

/////////////
// SETTERS //
/////////////

// WithInterval sets the retry interval for the Client.
// Default is 100 milliseconds.
// Noop if the interval is less than or equal to 0.
// It returns the Client for chaining.
func (rc *Client) WithInterval(interval time.Duration) *Client {
	if interval > 0 {
		rc.interval = interval
	}
	return rc
}

// WithBackoffMultiplier sets the backoff multiplier for the Client.
// After each retry, the configured interval is multiplied by this factor.
// Setting it to a value less than 1 will default it to 1.
// Default is 1.0, meaning no backoff.
// Noop if the multiplier is less than 1.
// It returns the Client for chaining.
func (rc *Client) WithBackoffMultiplier(multiplier float64) *Client {
	if multiplier >= 1 {
		rc.backoffMultiplier = multiplier
	}
	return rc
}

// WithMaxAttempts sets the maximum number of attempts for the Client.
// If set to 0, it will retry indefinitely until the timeout is reached.
// Default is 0, meaning no limit on attempts.
// Noop if the maxAttempts is less than 0.
// It returns the Client for chaining.
func (rc *Client) WithMaxAttempts(maxAttempts int) *Client {
	if maxAttempts >= 0 {
		rc.maxAttempts = maxAttempts
	}
	return rc
}

// WithTimeout sets the timeout for retries in the Client.
// If set to 0, there is no timeout and it will retry until the maximum number of retries is reached.
// Default is 1 second.
// Noop if the timeout is less than 0.
// It returns the Client for chaining.
func (rc *Client) WithTimeout(timeout time.Duration) *Client {
	if timeout >= 0 {
		rc.timeout = timeout
	}
	return rc
}

///////////////////////////
// CLIENT IMPLEMENTATION //
///////////////////////////

type operation struct {
	parent    *Client
	interval  time.Duration
	attempts  int
	startTime time.Time
	method    reflect.Value
	args      []reflect.Value
}

func (rc *Client) newOperation(method reflect.Value, args ...any) *operation {
	op := &operation{
		parent:    rc,
		interval:  rc.interval,
		attempts:  0,
		startTime: time.Now(),
		method:    method,
	}
	if method.Type().IsVariadic() {
		argCountWithoutVariadic := len(args) - 1
		last := args[argCountWithoutVariadic]
		lastVal := reflect.ValueOf(last)
		argCountVariadic := lastVal.Len()
		op.args = make([]reflect.Value, argCountWithoutVariadic+argCountVariadic)
		for i, arg := range args[:argCountWithoutVariadic] {
			op.args[i] = reflect.ValueOf(arg)
		}
		for i := range argCountVariadic {
			op.args[argCountWithoutVariadic+i] = lastVal.Index(i)
		}
	} else {
		op.args = make([]reflect.Value, len(args))
		for i, arg := range args {
			op.args[i] = reflect.ValueOf(arg)
		}
	}
	return op
}

// try attempts the operation.
// The first return value indicates success (true) or failure (false).
// The second return value is the duration to wait before the next retry.
//
//	If it is 0, no retry is needed.
//	This can be because the operation succeeded, or because the timeout or retry limit was reached.
//
// The third return value contains the return values of the operation.
func (op *operation) try() (bool, time.Duration, []reflect.Value) {
	res := op.method.Call(op.args)

	// check for success by converting the last return value to an error
	success := true
	if len(res) > 0 {
		if err, ok := res[len(res)-1].Interface().(error); ok && err != nil {
			success = false
		}
	}

	// if the operation succeeded, return true and no retry
	if success {
		return true, 0, res
	}

	// if the operation failed, check if we should retry
	op.attempts++
	retryAfter := op.interval
	op.interval = time.Duration(float64(op.interval) * op.parent.backoffMultiplier)
	if (op.parent.maxAttempts > 0 && op.attempts >= op.parent.maxAttempts) || (op.parent.timeout > 0 && time.Now().Add(retryAfter).After(op.startTime.Add(op.parent.timeout))) {
		// if we reached the maximum number of retries or the next retry would exceed the timeout, return false and no retry
		return false, 0, res
	}

	return false, retryAfter, res
}

// retry executes the given method with the provided arguments, retrying on failure.
func (rc *Client) retry(method reflect.Value, args ...any) []reflect.Value {
	op := rc.newOperation(method, args...)
	var ctx context.Context
	if len(args) > 0 {
		if ctxArg, ok := args[0].(context.Context); ok {
			ctx = ctxArg
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if rc.Timeout() > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, op.startTime.Add(rc.timeout))
		defer cancel()
	}
	interruptedOrTimeouted := ctx.Done()
	success, retryAfter, res := op.try()
	for !success && retryAfter > 0 {
		opCtx, opCancel := context.WithTimeout(ctx, retryAfter)
		expired := opCtx.Done()
		select {
		case <-interruptedOrTimeouted:
			retryAfter = 0 // stop retrying if the context was cancelled
		case <-expired:
			success, retryAfter, res = op.try()
		}
		opCancel()
	}
	return res
}

func errOrNil(val reflect.Value) error {
	if val.IsNil() {
		return nil
	}
	return val.Interface().(error)
}

// CreateOrUpdate wraps the controllerutil.CreateOrUpdate function and retries it on failure.
func (rc *Client) CreateOrUpdate(ctx context.Context, obj client.Object, f controllerutil.MutateFn) (controllerutil.OperationResult, error) {
	res := rc.retry(reflect.ValueOf(controllerutil.CreateOrUpdate), ctx, rc.internal, obj, f)
	return res[0].Interface().(controllerutil.OperationResult), errOrNil(res[1])
}

// CreateOrPatch wraps the controllerutil.CreateOrPatch function and retries it on failure.
func (rc *Client) CreateOrPatch(ctx context.Context, obj client.Object, f controllerutil.MutateFn) (controllerutil.OperationResult, error) {
	res := rc.retry(reflect.ValueOf(controllerutil.CreateOrPatch), ctx, rc.internal, obj, f)
	return res[0].Interface().(controllerutil.OperationResult), errOrNil(res[1])
}

// Create wraps the client's Create method and retries it on failure.
func (rc *Client) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	res := rc.retry(reflect.ValueOf(rc.internal.Create), ctx, obj, opts)
	return errOrNil(res[0])
}

// Delete wraps the client's Delete method and retries it on failure.
func (rc *Client) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	res := rc.retry(reflect.ValueOf(rc.internal.Delete), ctx, obj, opts)
	return errOrNil(res[0])
}

// DeleteAllOf wraps the client's DeleteAllOf method and retries it on failure.
func (rc *Client) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	res := rc.retry(reflect.ValueOf(rc.internal.DeleteAllOf), ctx, obj, opts)
	return errOrNil(res[0])
}

// Get wraps the client's Get method and retries it on failure.
func (rc *Client) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	res := rc.retry(reflect.ValueOf(rc.internal.Get), ctx, key, obj, opts)
	return errOrNil(res[0])
}

// List wraps the client's List method and retries it on failure.
func (rc *Client) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	res := rc.retry(reflect.ValueOf(rc.internal.List), ctx, list, opts)
	return errOrNil(res[0])
}

// Patch wraps the client's Patch method and retries it on failure.
func (rc *Client) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	res := rc.retry(reflect.ValueOf(rc.internal.Patch), ctx, obj, patch, opts)
	return errOrNil(res[0])
}

// Update wraps the client's Update method and retries it on failure.
func (rc *Client) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	res := rc.retry(reflect.ValueOf(rc.internal.Update), ctx, obj, opts)
	return errOrNil(res[0])
}

// GroupVersionKindFor wraps the client's GroupVersionKindFor method and retries it on failure.
func (rc *Client) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	res := rc.retry(reflect.ValueOf(rc.internal.GroupVersionKindFor), obj)
	return res[0].Interface().(schema.GroupVersionKind), errOrNil(res[1])
}

// IsObjectNamespaced wraps the client's IsObjectNamespaced method and retries it on failure.
func (rc *Client) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	res := rc.retry(reflect.ValueOf(rc.internal.IsObjectNamespaced), obj)
	return res[0].Interface().(bool), errOrNil(res[1])
}

// RESTMapper calls the internal client's RESTMapper method.
func (rc *Client) RESTMapper() meta.RESTMapper {
	return rc.internal.RESTMapper()
}

// Scheme calls the internal client's Scheme method.
func (rc *Client) Scheme() *runtime.Scheme {
	return rc.internal.Scheme()
}

// Status calls the internal client's Status method.
func (rc *Client) Status() client.SubResourceWriter {
	return rc.internal.Status()
}

// SubResource calls the internal client's SubResource method.
func (rc *Client) SubResource(subResource string) client.SubResourceClient {
	return rc.internal.SubResource(subResource)
}
