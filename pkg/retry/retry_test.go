package retry_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/openmcp-project/controller-utils/pkg/retry"
	testutils "github.com/openmcp-project/controller-utils/pkg/testing"
)

func TestConditions(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Retry Test Suite")
}

// mockControl is a helper struct to control the behavior of the fake client.
// Each attempt will increase the 'attempts' counter.
// Returns a mockError if
// - 'fail' is less than 0
// - 'fail' is greater than 0 and the number of attempts is less than or equal to 'fail'
// Returns nil otherwise.
type mockControl struct {
	fail     int
	attempts int
}

func (mc *mockControl) reset(failCount int) {
	mc.fail = failCount
	mc.attempts = 0
}

func (mc *mockControl) try() error {
	mc.attempts++
	if mc.fail < 0 || (mc.fail > 0 && mc.attempts <= mc.fail) {
		return errMock
	}
	return nil
}

var errMock = fmt.Errorf("mock error")

func defaultTestSetup() (*testutils.Environment, *mockControl) {
	mc := &mockControl{}
	return testutils.NewEnvironmentBuilder().
		WithFakeClient(nil).
		WithFakeClientBuilderCall("WithInterceptorFuncs", interceptor.Funcs{
			Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				if err := mc.try(); err != nil {
					return err
				}
				return client.Get(ctx, key, obj, opts...)
			},
			List: func(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
				if err := mc.try(); err != nil {
					return err
				}
				return client.List(ctx, list, opts...)
			},
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				if err := mc.try(); err != nil {
					return err
				}
				return client.Create(ctx, obj, opts...)
			},
			Delete: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
				if err := mc.try(); err != nil {
					return err
				}
				return client.Delete(ctx, obj, opts...)
			},
			DeleteAllOf: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.DeleteAllOfOption) error {
				if err := mc.try(); err != nil {
					return err
				}
				return client.DeleteAllOf(ctx, obj, opts...)
			},
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				if err := mc.try(); err != nil {
					return err
				}
				return client.Update(ctx, obj, opts...)
			},
			Patch: func(ctx context.Context, client client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
				if err := mc.try(); err != nil {
					return err
				}
				return client.Patch(ctx, obj, patch, opts...)
			},
		}).
		Build(), mc
}

var _ = Describe("Client", func() {

	It("should not retry if the operation succeeds immediately", func() {
		env, mc := defaultTestSetup()
		c := retry.NewRetryingClient(env.Client())

		// create a Namespace
		ns := &corev1.Namespace{}
		ns.Name = "test"
		mc.reset(0)
		Expect(c.Create(env.Ctx, ns)).To(Succeed())
		Expect(mc.attempts).To(Equal(1))

		// get the Namespace
		mc.reset(0)
		Expect(c.Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
		Expect(mc.attempts).To(Equal(1))

		// list Namespaces
		mc.reset(0)
		nsList := &corev1.NamespaceList{}
		Expect(c.List(env.Ctx, nsList)).To(Succeed())
		Expect(mc.attempts).To(Equal(1))
		Expect(nsList.Items).To(ContainElement(MatchFields(IgnoreExtras, Fields{
			"ObjectMeta": MatchFields(IgnoreExtras, Fields{
				"Name": Equal("test"),
			}),
		})))

		// update the Namespace
		mc.reset(0)
		ns.Labels = map[string]string{"test": "label"}
		Expect(c.Update(env.Ctx, ns)).To(Succeed())
		Expect(mc.attempts).To(Equal(1))

		// patch the Namespace
		mc.reset(0)
		old := ns.DeepCopy()
		ns.Labels = nil
		Expect(c.Patch(env.Ctx, ns, client.MergeFrom(old))).To(Succeed())
		Expect(mc.attempts).To(Equal(1))

		// delete the Namespace
		mc.reset(0)
		Expect(c.Delete(env.Ctx, ns)).To(Succeed())
		Expect(mc.attempts).To(Equal(1))

		// delete all Namespaces
		mc.reset(0)
		Expect(c.DeleteAllOf(env.Ctx, &corev1.Namespace{})).To(Succeed())
		Expect(mc.attempts).To(Equal(1))
	})

	It("should retry if the operation does not succeed immediately", func() {
		env, mc := defaultTestSetup()
		c := retry.NewRetryingClient(env.Client()).WithMaxAttempts(5).WithTimeout(0)

		// create a Namespace
		ns := &corev1.Namespace{}
		ns.Name = "test"
		mc.reset(2)
		Expect(c.Create(env.Ctx, ns)).To(Succeed())
		Expect(mc.attempts).To(Equal(3))

		// get the Namespace
		mc.reset(2)
		Expect(c.Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
		Expect(mc.attempts).To(Equal(3))

		// list Namespaces
		mc.reset(2)
		nsList := &corev1.NamespaceList{}
		Expect(c.List(env.Ctx, nsList)).To(Succeed())
		Expect(mc.attempts).To(Equal(3))
		Expect(nsList.Items).To(ContainElement(MatchFields(IgnoreExtras, Fields{
			"ObjectMeta": MatchFields(IgnoreExtras, Fields{
				"Name": Equal("test"),
			}),
		})))

		// update the Namespace
		mc.reset(2)
		ns.Labels = map[string]string{"test": "label"}
		Expect(c.Update(env.Ctx, ns)).To(Succeed())
		Expect(mc.attempts).To(Equal(3))

		// patch the Namespace
		mc.reset(2)
		old := ns.DeepCopy()
		ns.Labels = nil
		Expect(c.Patch(env.Ctx, ns, client.MergeFrom(old))).To(Succeed())
		Expect(mc.attempts).To(Equal(3))

		// delete the Namespace
		mc.reset(2)
		Expect(c.Delete(env.Ctx, ns)).To(Succeed())
		Expect(mc.attempts).To(Equal(3))

		// delete all Namespaces
		mc.reset(2)
		Expect(c.DeleteAllOf(env.Ctx, &corev1.Namespace{})).To(Succeed())
		Expect(mc.attempts).To(Equal(3))
	})

	It("should not retry more often than configured", func() {
		env, mc := defaultTestSetup()
		c := retry.NewRetryingClient(env.Client()).WithMaxAttempts(5).WithTimeout(0)

		// create a Namespace
		ns := &corev1.Namespace{}
		ns.Name = "test"
		mc.reset(-1)
		Expect(c.Create(env.Ctx, ns)).ToNot(Succeed())
		Expect(mc.attempts).To(Equal(5))

		// get the Namespace
		mc.reset(-1)
		Expect(c.Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).ToNot(Succeed())
		Expect(mc.attempts).To(Equal(5))

		// list Namespaces
		mc.reset(-1)
		nsList := &corev1.NamespaceList{}
		Expect(c.List(env.Ctx, nsList)).ToNot(Succeed())
		Expect(mc.attempts).To(Equal(5))

		// update the Namespace
		mc.reset(-1)
		ns.Labels = map[string]string{"test": "label"}
		Expect(c.Update(env.Ctx, ns)).ToNot(Succeed())
		Expect(mc.attempts).To(Equal(5))

		// patch the Namespace
		mc.reset(-1)
		old := ns.DeepCopy()
		ns.Labels = nil
		Expect(c.Patch(env.Ctx, ns, client.MergeFrom(old))).ToNot(Succeed())
		Expect(mc.attempts).To(Equal(5))

		// delete the Namespace
		mc.reset(-1)
		Expect(c.Delete(env.Ctx, ns)).ToNot(Succeed())
		Expect(mc.attempts).To(Equal(5))

		// delete all Namespaces
		mc.reset(-1)
		Expect(c.DeleteAllOf(env.Ctx, &corev1.Namespace{})).ToNot(Succeed())
		Expect(mc.attempts).To(Equal(5))
	})

	It("should not retry longer than configured", func() {
		env, mc := defaultTestSetup()
		c := retry.NewRetryingClient(env.Client()).WithMaxAttempts(0).WithTimeout(500 * time.Millisecond)

		// for performance reasons, let's test this for Create only
		ns := &corev1.Namespace{}
		ns.Name = "test"
		mc.reset(-1)
		now := time.Now()
		timeoutCtx, cancel := context.WithTimeout(env.Ctx, 1*time.Second)
		defer cancel()
		Expect(c.Create(timeoutCtx, ns)).ToNot(Succeed())
		after := time.Now()
		Expect(after.Sub(now)).To(BeNumerically(">=", 400*time.Millisecond))
		Expect(after.Sub(now)).To(BeNumerically("<", 1*time.Second))
		Expect(mc.attempts).To(BeNumerically(">=", 4))
		Expect(mc.attempts).To(BeNumerically("<=", 5))
	})

	It("should apply the backoff multiplier correctly", func() {
		env, mc := defaultTestSetup()
		c := retry.NewRetryingClient(env.Client()).WithMaxAttempts(0).WithTimeout(500 * time.Millisecond).WithBackoffMultiplier(3.0)

		// for performance reasons, let's test this for Create only
		ns := &corev1.Namespace{}
		ns.Name = "test"
		mc.reset(-1)
		now := time.Now()
		timeoutCtx, cancel := context.WithTimeout(env.Ctx, 1*time.Second)
		defer cancel()
		Expect(c.Create(timeoutCtx, ns)).ToNot(Succeed())
		after := time.Now()
		Expect(after.Sub(now)).To(BeNumerically(">=", 400*time.Millisecond))
		Expect(after.Sub(now)).To(BeNumerically("<", 1*time.Second))
		Expect(mc.attempts).To(BeNumerically("==", 3))
	})

	It("should abort if the context is canceled", func() {
		env, mc := defaultTestSetup()
		c := retry.NewRetryingClient(env.Client()).WithMaxAttempts(0).WithTimeout(500 * time.Millisecond)

		// for performance reasons, let's test this for Create only
		ns := &corev1.Namespace{}
		ns.Name = "test"
		mc.reset(-1)
		now := time.Now()
		timeoutCtx, cancel := context.WithTimeout(env.Ctx, 200*time.Millisecond)
		defer cancel()
		Expect(c.Create(timeoutCtx, ns)).ToNot(Succeed())
		after := time.Now()
		Expect(after.Sub(now)).To(BeNumerically("<", 300*time.Millisecond))
		Expect(mc.attempts).To(BeNumerically("<=", 3))
	})

	It("should handle WithContext correctly", func() {
		env, mc := defaultTestSetup()
		c := retry.NewRetryingClient(env.Client()).WithMaxAttempts(0).WithTimeout(500 * time.Millisecond)

		type dummy struct {
			corev1.Namespace
		}

		mc.reset(-1)
		now := time.Now()
		timeoutCtx, cancel := context.WithTimeout(env.Ctx, 200*time.Millisecond)
		defer cancel()
		_, err := c.WithContext(timeoutCtx).GroupVersionKindFor(&dummy{})
		Expect(err).To(HaveOccurred())
		after := time.Now()
		Expect(after.Sub(now)).To(BeNumerically("<", 300*time.Millisecond))

		// should not use the same context again, it should have been reset
		now = time.Now()
		_, err = c.GroupVersionKindFor(&dummy{})
		Expect(err).To(HaveOccurred())
		after = time.Now()
		Expect(after.Sub(now)).To(BeNumerically(">", 300*time.Millisecond))
	})

	It("should pass the arguments through correctly", func() {
		env, mc := defaultTestSetup()
		c := retry.NewRetryingClient(env.Client())

		// for performance reasons, let's test this for Create only
		s1 := &corev1.Secret{}
		s1.Name = "test"
		s1.Namespace = "foo"
		Expect(env.Client().Create(env.Ctx, s1)).To(Succeed())
		s2 := &corev1.Secret{}
		s2.Name = "test"
		s2.Namespace = "bar"
		Expect(env.Client().Create(env.Ctx, s2)).To(Succeed())
		mc.reset(0)
		l1 := &corev1.SecretList{}
		Expect(c.List(env.Ctx, l1)).To(Succeed())
		Expect(mc.attempts).To(Equal(1))
		Expect(l1.Items).To(ConsistOf(
			MatchFields(IgnoreExtras, Fields{
				"ObjectMeta": MatchFields(IgnoreExtras, Fields{
					"Name":      Equal("test"),
					"Namespace": Equal("foo"),
				}),
			}),
			MatchFields(IgnoreExtras, Fields{
				"ObjectMeta": MatchFields(IgnoreExtras, Fields{
					"Name":      Equal("test"),
					"Namespace": Equal("bar"),
				}),
			}),
		))
		mc.reset(0)
		l2 := &corev1.SecretList{}
		Expect(c.List(env.Ctx, l2, client.InNamespace("foo"))).To(Succeed())
		Expect(mc.attempts).To(Equal(1))
		Expect(l2.Items).To(ConsistOf(
			MatchFields(IgnoreExtras, Fields{
				"ObjectMeta": MatchFields(IgnoreExtras, Fields{
					"Name":      Equal("test"),
					"Namespace": Equal("foo"),
				}),
			}),
		))
	})

	It("should correctly handle CreateOrUpdate and CreateOrPatch", func() {
		env, mc := defaultTestSetup()
		c := retry.NewRetryingClient(env.Client()).WithMaxAttempts(5).WithTimeout(0)

		// create or update namespace
		// we cannot check mc.attempts here, because CreateOrUpdate calls multiple methods on the client internally
		ns := &corev1.Namespace{}
		ns.Name = "test"
		mc.reset(0)
		Expect(c.CreateOrUpdate(env.Ctx, ns, func() error {
			return nil
		}))
		Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
		mc.reset(0)
		Expect(c.CreateOrUpdate(env.Ctx, ns, func() error {
			ns.Labels = map[string]string{"test": "label"}
			return nil
		}))
		Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
		Expect(ns.Labels).To(HaveKeyWithValue("test", "label"))
		mc.reset(2)
		Expect(c.CreateOrUpdate(env.Ctx, ns, func() error {
			ns.Labels = map[string]string{"test2": "label2"}
			return nil
		}))
		Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
		Expect(ns.Labels).To(HaveKeyWithValue("test2", "label2"))
		Expect(env.Client().Delete(env.Ctx, ns)).To(Succeed())

		// create or patch namespace
		ns = &corev1.Namespace{}
		ns.Name = "test"
		mc.reset(0)
		Expect(c.CreateOrPatch(env.Ctx, ns, func() error {
			return nil
		}))
		Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
		mc.reset(0)
		Expect(c.CreateOrPatch(env.Ctx, ns, func() error {
			ns.Labels = map[string]string{"test": "label"}
			return nil
		}))
		Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
		Expect(ns.Labels).To(HaveKeyWithValue("test", "label"))
		mc.reset(2)
		Expect(c.CreateOrUpdate(env.Ctx, ns, func() error {
			ns.Labels = map[string]string{"test2": "label2"}
			return nil
		}))
		Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
		Expect(ns.Labels).To(HaveKeyWithValue("test2", "label2"))
		Expect(env.Client().Delete(env.Ctx, ns)).To(Succeed())
	})

})
