// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package logging_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/openmcp-project/controller-utils/pkg/collections"
	"github.com/openmcp-project/controller-utils/pkg/logging"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Installations Test Suite")
}

var _ = Describe("Logging Framework Tests", func() {

	It("should not modify the logger if any method is called", func() {
		compareToLogger := logging.Wrap(logging.PreventKeyConflicts(logr.Discard()))
		log := logging.Wrap(logging.PreventKeyConflicts(logr.Discard()))
		Expect(reflect.DeepEqual(log, compareToLogger)).To(BeTrue())

		log.Debug("foo", "bar", "baz", "bar", "baz")
		Expect(reflect.DeepEqual(log, compareToLogger)).To(BeTrue(), "calling log.Debug should not modify the logger")

		log.Info("foo", "bar", "baz", "bar", "baz")
		Expect(reflect.DeepEqual(log, compareToLogger)).To(BeTrue(), "calling log.Info should not modify the logger")

		log.Error(nil, "foo", "bar", "baz", "bar", "baz")
		Expect(reflect.DeepEqual(log, compareToLogger)).To(BeTrue(), "calling log.Error should not modify the logger")

		log.WithName("myname")
		Expect(reflect.DeepEqual(log, compareToLogger)).To(BeTrue(), "calling log.WithName should not modify the logger")

		log.WithValues("foo", "bar")
		Expect(reflect.DeepEqual(log, compareToLogger)).To(BeTrue(), "calling log.WithValues should not modify the logger")
	})

	Context("LogRequeue", func() {

		var log logging.Logger
		var sink *TestLogSink

		BeforeEach(func() {
			sink = NewTestLogSink(logging.DEBUG)
			log = logging.Wrap(logr.New(sink))
		})

		It("should not log anything if RequeueAfter is 0", func() {
			log.LogRequeue(reconcile.Result{})
			Expect(sink.Messages.Size()).To(Equal(0))
		})

		It("should log a message if RequeueAfter is set", func() {
			now := time.Now()
			requeueAfter := 42 * time.Second
			log.LogRequeue(reconcile.Result{RequeueAfter: requeueAfter})
			Expect(sink.Messages.Size()).To(Equal(1))
			msg := sink.Messages.Poll()
			Expect(msg.Verbosity).To(Equal(logging.LevelToVerbosity(logging.DEBUG)))
			Expect(msg.Message).To(Equal("Requeuing object for reconciliation"))
			Expect(msg.KeysAndVals).To(HaveKeyWithValue("after", requeueAfter.String()))
			Expect(msg.KeysAndVals).To(HaveKey("at"))
			at, err := time.Parse(time.RFC3339, msg.KeysAndVals["at"].(string))
			Expect(err).NotTo(HaveOccurred())
			Expect(at).To(BeTemporally("~", now.Add(requeueAfter), time.Second))
		})

		It("should log at the provided verbosity level", func() {
			now := time.Now()
			requeueAfter := 42 * time.Second
			log.LogRequeue(reconcile.Result{RequeueAfter: requeueAfter}, logging.INFO)
			Expect(sink.Messages.Size()).To(Equal(1))
			msg := sink.Messages.Poll()
			Expect(msg.Verbosity).To(Equal(logging.LevelToVerbosity(logging.INFO)))
			Expect(msg.Message).To(Equal("Requeuing object for reconciliation"))
			Expect(msg.KeysAndVals).To(HaveKeyWithValue("after", requeueAfter.String()))
			Expect(msg.KeysAndVals).To(HaveKey("at"))
			at, err := time.Parse(time.RFC3339, msg.KeysAndVals["at"].(string))
			Expect(err).NotTo(HaveOccurred())
			Expect(at).To(BeTemporally("~", now.Add(requeueAfter), time.Second))
		})

	})

})

func NewTestLogSink(level logging.LogLevel) *TestLogSink {
	return &TestLogSink{
		Messages:     collections.NewLinkedList[*LogMessage](),
		enabledLevel: level,
	}
}

type TestLogSink struct {
	Messages     collections.Queue[*LogMessage]
	enabledLevel logging.LogLevel
}

type LogMessage struct {
	Error       error
	Verbosity   int
	Message     string
	KeysAndVals map[string]any
}

// Enabled implements logr.LogSink.
func (t *TestLogSink) Enabled(level int) bool {
	return t.enabledLevel >= logging.LogLevel(level)
}

// Error implements logr.LogSink.
func (t *TestLogSink) Error(err error, msg string, keysAndValues ...any) {
	t.log(int(logging.ERROR), err, msg, keysAndValues...)
}

// Info implements logr.LogSink.
func (t *TestLogSink) Info(level int, msg string, keysAndValues ...any) {
	t.log(level, nil, msg, keysAndValues...)
}

func (t *TestLogSink) log(verbosity int, err error, msg string, keysAndValues ...any) {
	kv := make(map[string]any, (len(keysAndValues)+1)/2)
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		k, ok := keysAndValues[i].(string)
		if !ok {
			k = fmt.Sprint(keysAndValues[i])
		}
		kv[k] = keysAndValues[i+1]
	}
	_ = t.Messages.Push(&LogMessage{
		Error:       err,
		Verbosity:   verbosity,
		Message:     msg,
		KeysAndVals: kv,
	})
}

// Init implements logr.LogSink.
func (t *TestLogSink) Init(_ logr.RuntimeInfo) {}

// WithName implements logr.LogSink.
func (t *TestLogSink) WithName(name string) logr.LogSink {
	panic("not implemented")
}

// WithValues implements logr.LogSink.
func (t *TestLogSink) WithValues(keysAndValues ...any) logr.LogSink {
	panic("not implemented")
}

var _ logr.LogSink = &TestLogSink{}
