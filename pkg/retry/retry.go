package retry

import (
	"context"
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
	context           context.Context
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
		context:           context.Background(),   // default context
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

// MaxAttempts returns the configured maximum number of retries.
func (rc *Client) MaxAttempts() int {
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

// WithContext sets the context for the next call of either GroupVersionKindFor or IsObjectNamespaced.
// Since the signature of these methods does not allow passing a context, and the retrying can not be cancelled without one,
// this method is required to inject the context to be used for the aforementioned methods.
// Note that any function of this Client that actually retries an operation will reset this context, but only for GroupVersionKindFor and IsObjectNamespaced it will actually be used.
// The intended use of this method is something like this:
//
//	c.WithContext(ctx).GroupVersionKindFor(obj)
//	c.WithContext(ctx).IsObjectNamespaced(obj)
//
// If no context is injected via this method, both GroupVersionKindFor and IsObjectNamespaced will use the default context.Background().
// It returns the Client for chaining.
func (rc *Client) WithContext(ctx context.Context) *Client {
	if ctx == nil {
		ctx = context.Background() // default to background context if nil
	}
	rc.context = ctx
	return rc
}

///////////////////////////
// CLIENT IMPLEMENTATION //
///////////////////////////

type callbackFn func(ctx context.Context) error

type operation struct {
	parent    *Client
	interval  time.Duration
	attempts  int
	startTime time.Time
	cfn       callbackFn
}

func (rc *Client) newOperation(cfn callbackFn) *operation {
	return &operation{
		parent:    rc,
		interval:  rc.interval,
		attempts:  0,
		startTime: time.Now(),
		cfn:       cfn,
	}
}

// try attempts the operation.
// The first return value indicates success (true) or failure (false).
// The second return value is the duration to wait before the next retry.
//
//	If it is 0, no retry is needed.
//	This can be because the operation succeeded, or because the timeout or retry limit was reached.
//
// The third return value contains the return values of the operation.
func (op *operation) try(ctx context.Context) (bool, time.Duration) {
	err := op.cfn(ctx)

	// if the operation succeeded, return true and no retry
	if err == nil {
		return true, 0
	}

	// if the operation failed, check if we should retry
	op.attempts++
	retryAfter := op.interval
	op.interval = time.Duration(float64(op.interval) * op.parent.backoffMultiplier)
	if (op.parent.maxAttempts > 0 && op.attempts >= op.parent.maxAttempts) || (op.parent.timeout > 0 && time.Now().Add(retryAfter).After(op.startTime.Add(op.parent.timeout))) {
		// if we reached the maximum number of retries or the next retry would exceed the timeout, return false and no retry
		return false, 0
	}

	return false, retryAfter
}

// retry executes the given method with the provided arguments, retrying on failure.
func (rc *Client) retry(ctx context.Context, cfn callbackFn) {
	rc.WithContext(context.Background()) // reset context
	op := rc.newOperation(cfn)
	if rc.Timeout() > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, op.startTime.Add(rc.timeout))
		defer cancel()
	}
	interruptedOrTimeouted := ctx.Done()
	success, retryAfter := op.try(ctx)
	for !success && retryAfter > 0 {
		opCtx, opCancel := context.WithTimeout(ctx, retryAfter)
		expired := opCtx.Done()
		select {
		case <-interruptedOrTimeouted:
			retryAfter = 0 // stop retrying if the context was cancelled
		case <-expired:
			success, retryAfter = op.try(ctx)
		}
		opCancel()
	}
}

// CreateOrUpdate wraps the controllerutil.CreateOrUpdate function and retries it on failure.
func (rc *Client) CreateOrUpdate(ctx context.Context, obj client.Object, f controllerutil.MutateFn) (res controllerutil.OperationResult, err error) {
	rc.retry(ctx, func(ctx context.Context) error {
		res, err = controllerutil.CreateOrUpdate(ctx, rc.internal, obj, f)
		return err
	})
	return
}

// CreateOrPatch wraps the controllerutil.CreateOrPatch function and retries it on failure.
func (rc *Client) CreateOrPatch(ctx context.Context, obj client.Object, f controllerutil.MutateFn) (res controllerutil.OperationResult, err error) {
	rc.retry(ctx, func(ctx context.Context) error {
		res, err = controllerutil.CreateOrPatch(ctx, rc.internal, obj, f)
		return err
	})
	return
}

// Create wraps the client's Create method and retries it on failure.
func (rc *Client) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) (err error) {
	rc.retry(ctx, func(ctx context.Context) error {
		err = rc.internal.Create(ctx, obj, opts...)
		return err
	})
	return
}

// Delete wraps the client's Delete method and retries it on failure.
func (rc *Client) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) (err error) {
	rc.retry(ctx, func(ctx context.Context) error {
		err = rc.internal.Delete(ctx, obj, opts...)
		return err
	})
	return
}

// DeleteAllOf wraps the client's DeleteAllOf method and retries it on failure.
func (rc *Client) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) (err error) {
	rc.retry(ctx, func(ctx context.Context) error {
		err = rc.internal.DeleteAllOf(ctx, obj, opts...)
		return err
	})
	return
}

// Get wraps the client's Get method and retries it on failure.
func (rc *Client) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) (err error) {
	rc.retry(ctx, func(ctx context.Context) error {
		err = rc.internal.Get(ctx, key, obj, opts...)
		return err
	})
	return
}

// List wraps the client's List method and retries it on failure.
func (rc *Client) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (err error) {
	rc.retry(ctx, func(ctx context.Context) error {
		err = rc.internal.List(ctx, list, opts...)
		return err
	})
	return
}

// Patch wraps the client's Patch method and retries it on failure.
func (rc *Client) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) (err error) {
	rc.retry(ctx, func(ctx context.Context) error {
		err = rc.internal.Patch(ctx, obj, patch, opts...)
		return err
	})
	return
}

// Update wraps the client's Update method and retries it on failure.
func (rc *Client) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) (err error) {
	rc.retry(ctx, func(ctx context.Context) error {
		err = rc.internal.Update(ctx, obj, opts...)
		return err
	})
	return
}

// GroupVersionKindFor wraps the client's GroupVersionKindFor method and retries it on failure.
func (rc *Client) GroupVersionKindFor(obj runtime.Object) (gvk schema.GroupVersionKind, err error) {
	rc.retry(rc.context, func(ctx context.Context) error {
		gvk, err = rc.internal.GroupVersionKindFor(obj)
		return err
	})
	return
}

// IsObjectNamespaced wraps the client's IsObjectNamespaced method and retries it on failure.
func (rc *Client) IsObjectNamespaced(obj runtime.Object) (namespaced bool, err error) {
	rc.retry(rc.context, func(ctx context.Context) error {
		namespaced, err = rc.internal.IsObjectNamespaced(obj)
		return err
	})
	return
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
