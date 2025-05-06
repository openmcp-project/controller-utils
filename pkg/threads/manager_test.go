package threads_test

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/threads"
)

type testValue struct {
	val atomic.Int32
}

func (t *testValue) AddFuncRun(i int) func(context.Context) error {
	return func(ctx context.Context) error {
		t.val.Add(int32(i))
		return nil
	}
}

func (t *testValue) AddFuncOnFinish(i int) func(context.Context, threads.ThreadReturn) {
	return func(context.Context, threads.ThreadReturn) {
		t.val.Add(int32(i))
	}
}

func (t *testValue) Value() int32 {
	return t.val.Load()
}

var _ = Describe("ThreadManager", func() {

	Context("ThreadManager", func() {

		It("should execute multiple threads", func() {
			t := &testValue{}
			mgr := threads.NewThreadManager(context.Background(), context.Background(), nil)
			threadCount := 5
			addPerThread := 1
			for i := range threadCount {
				mgr.Run(strconv.Itoa(i), t.AddFuncRun(addPerThread), nil)
			}
			mgr.Start()
			for i := range threadCount {
				mgr.Run(strconv.Itoa(i), t.AddFuncRun(addPerThread), nil)
			}
			mgr.Stop()
			Expect(t.Value()).To(BeNumerically("==", 2*threadCount*addPerThread))
		})

		It("should execute onFinish functions in threads", func() {
			t := &testValue{}
			mgr := threads.NewThreadManager(context.Background(), context.Background(), nil)
			threadCount := 5
			addPerThread := 1
			for i := range threadCount {
				mgr.Run(strconv.Itoa(i), t.AddFuncRun(addPerThread), t.AddFuncOnFinish((-1)*addPerThread))
			}
			mgr.Start()
			for i := range threadCount {
				mgr.Run(strconv.Itoa(i), t.AddFuncRun(addPerThread), nil)
			}
			mgr.Stop()
			Expect(t.Value()).To(BeNumerically("==", threadCount*addPerThread))
		})

		It("should execute onFinish functions in thread manager", func() {
			t := &testValue{}
			addPerThread := 1
			mgr := threads.NewThreadManager(context.Background(), context.Background(), t.AddFuncOnFinish((-1)*addPerThread))
			threadCount := 5
			for i := range threadCount {
				mgr.Run(strconv.Itoa(i), t.AddFuncRun(addPerThread), nil)
			}
			mgr.Start()
			for i := range threadCount {
				mgr.Run(strconv.Itoa(i), t.AddFuncRun(addPerThread), nil)
			}
			mgr.Stop()
			Expect(t.Value()).To(BeNumerically("==", 0))
		})

		It("should correctly return whether the manager has been started or stopped", func() {
			mgr := threads.NewThreadManager(context.Background(), context.Background(), nil)
			Expect(mgr.IsStarted()).To(BeFalse())
			Expect(mgr.IsStopped()).To(BeFalse())
			Expect(mgr.IsRunning()).To(BeFalse())
			mgr.Start()
			Expect(mgr.IsStarted()).To(BeTrue())
			Expect(mgr.IsStopped()).To(BeFalse())
			Expect(mgr.IsRunning()).To(BeTrue())
			mgr.Stop()
			Expect(mgr.IsStarted()).To(BeTrue())
			Expect(mgr.IsStopped()).To(BeTrue())
			Expect(mgr.IsRunning()).To(BeFalse())
		})

		It("should stop if the Stop() method is invoked", func() {
			mgr := threads.NewThreadManager(context.Background(), context.Background(), nil)
			now := time.Now()
			mgr.Run("sleep", func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(10 * time.Second):
					return nil
				}
			}, nil)
			mgr.Start()
			mgr.Stop()
			Expect(time.Now()).To(BeTemporally("<", now.Add(3*time.Second)))
		})

		It("should stop if the manager context is cancelled", func() {
			mgrCtx, cancel := context.WithCancel(context.Background())
			mgr := threads.NewThreadManager(mgrCtx, context.Background(), nil)
			now := time.Now()
			mgr.Run("sleep", func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(10 * time.Second):
					return nil
				}
			}, nil)
			mgr.Start()
			cancel()
			Expect(time.Now()).To(BeTemporally("<", now.Add(3*time.Second)))
		})

		It("should have no effect if the manager is started multiple times", func() {
			t := &testValue{}
			mgr := threads.NewThreadManager(context.Background(), context.Background(), nil)
			threadCount := 5
			addPerThread := 1
			for i := range threadCount {
				mgr.Run(strconv.Itoa(i), t.AddFuncRun(addPerThread), nil)
			}
			mgr.Start()
			mgr.Start()
			mgr.Start()
			for i := range threadCount {
				mgr.Run(strconv.Itoa(i), t.AddFuncRun(addPerThread), nil)
			}
			mgr.Start()
			mgr.Start()
			mgr.Start()
			mgr.Stop()
			Expect(t.Value()).To(BeNumerically("==", 2*threadCount*addPerThread))
		})

		It("should have no effect if the manager is stopped multiple times", func() {
			t := &testValue{}
			mgr := threads.NewThreadManager(context.Background(), context.Background(), nil)
			threadCount := 5
			addPerThread := 1
			for i := range threadCount {
				mgr.Run(strconv.Itoa(i), t.AddFuncRun(addPerThread), nil)
			}
			mgr.Start()
			for i := range threadCount {
				mgr.Run(strconv.Itoa(i), t.AddFuncRun(addPerThread), nil)
			}
			mgr.Stop()
			mgr.Stop()
			mgr.Stop()
			Expect(t.Value()).To(BeNumerically("==", 2*threadCount*addPerThread))
		})

		It("should panic if Start() is called after Stop()", func() {
			mgr := threads.NewThreadManager(context.Background(), context.Background(), nil)
			mgr.Start()
			mgr.Stop()
			Expect(mgr.Start).To(Panic())
		})

	})

})
