package controller_test

import (
	"fmt"
	"slices"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openmcp-project/controller-utils/pkg/conditions"
	"github.com/openmcp-project/controller-utils/pkg/controller/smartrequeue"
	. "github.com/openmcp-project/controller-utils/pkg/testing/matchers"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/scheme"

	"github.com/openmcp-project/controller-utils/pkg/controller"
	"github.com/openmcp-project/controller-utils/pkg/errors"
	testutils "github.com/openmcp-project/controller-utils/pkg/testing"
)

var coScheme *runtime.Scheme

var _ = Describe("Status Updater", func() {

	It("should update an empty status", func() {
		env := testutils.NewEnvironmentBuilder().WithFakeClient(coScheme).WithInitObjectPath("testdata", "test-02").WithDynamicObjectsWithStatus(&CustomObject{}).Build()
		obj := &CustomObject{}
		Expect(env.Client().Get(env.Ctx, controller.ObjectKey("nostatus", "default"), obj)).To(Succeed())
		rr := controller.ReconcileResult[*CustomObject]{
			Object:         obj,
			ReconcileError: errors.WithReason(fmt.Errorf("test error"), "TestError"),
			Conditions:     dummyConditions(),
		}
		su := preconfiguredStatusUpdaterBuilder().Build()
		now := time.Now()
		res, err := su.UpdateStatus(env.Ctx, env.Client(), rr)
		Expect(res).To(Equal(rr.Result))
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(rr.ReconcileError))
		Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(obj), obj)).To(Succeed())

		Expect(obj.Status.Phase).To(Equal(PhaseFailed))
		Expect(obj.Status.ObservedGeneration).To(Equal(obj.GetGeneration()))
		Expect(obj.Status.Reason).To(Equal("TestError"))
		Expect(obj.Status.Message).To(ContainSubstring("test error"))
		Expect(obj.Status.LastReconcileTime.Time).To(BeTemporally("~", now, 1*time.Second))
		Expect(obj.Status.Conditions).To(ConsistOf(
			MatchCondition(TestCondition().
				WithType("TestConditionTrue").
				WithStatus(metav1.ConditionTrue).
				WithObservedGeneration(10).
				WithReason("TestReasonTrue").
				WithMessage("TestMessageTrue").
				WithLastTransitionTime(obj.Status.LastReconcileTime)),
			MatchCondition(TestCondition().
				WithType("TestConditionFalse").
				WithStatus(metav1.ConditionFalse).
				WithObservedGeneration(10).
				WithReason("TestReasonFalse").
				WithMessage("TestMessageFalse").
				WithLastTransitionTime(obj.Status.LastReconcileTime)),
		))
	})

	It("should not hide a reconciliation error if the object is nil", func() {
		env := testutils.NewEnvironmentBuilder().WithFakeClient(coScheme).WithInitObjectPath("testdata", "test-02").WithDynamicObjectsWithStatus(&CustomObject{}).Build()
		rr := controller.ReconcileResult[*CustomObject]{
			ReconcileError: errors.WithReason(fmt.Errorf("test error"), "TestError"),
		}
		su := preconfiguredStatusUpdaterBuilder().Build()
		_, err := su.UpdateStatus(env.Ctx, env.Client(), rr)
		Expect(err).To(HaveOccurred())
	})

	It("should update an existing status", func() {
		env := testutils.NewEnvironmentBuilder().WithFakeClient(coScheme).WithInitObjectPath("testdata", "test-02").WithDynamicObjectsWithStatus(&CustomObject{}).Build()
		obj := &CustomObject{}
		Expect(env.Client().Get(env.Ctx, controller.ObjectKey("status", "default"), obj)).To(Succeed())
		oldTransitionTime := conditions.GetCondition(obj.Status.Conditions, "TestConditionTrue").LastTransitionTime
		rr := controller.ReconcileResult[*CustomObject]{
			Object:     obj,
			Conditions: dummyConditions(),
		}
		su := preconfiguredStatusUpdaterBuilder().WithPhaseUpdateFunc(func(obj *CustomObject, rr controller.ReconcileResult[*CustomObject]) (string, error) {
			return PhaseSucceeded, nil
		}).Build()
		now := time.Now()
		res, err := su.UpdateStatus(env.Ctx, env.Client(), rr)
		Expect(res).To(Equal(rr.Result))
		Expect(err).ToNot(HaveOccurred())
		Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(obj), obj)).To(Succeed())

		Expect(obj.Status.Phase).To(Equal(PhaseSucceeded))
		Expect(obj.Status.ObservedGeneration).To(Equal(obj.GetGeneration()))
		Expect(obj.Status.Reason).To(BeEmpty())
		Expect(obj.Status.Message).To(BeEmpty())
		Expect(obj.Status.LastReconcileTime.Time).To(BeTemporally("~", now, 1*time.Second))
		Expect(obj.Status.Conditions).To(ConsistOf(
			MatchCondition(TestCondition().
				WithType("TestConditionTrue").
				WithStatus(metav1.ConditionTrue).
				WithObservedGeneration(10).
				WithReason("TestReasonTrue").
				WithMessage("TestMessageTrue").
				WithLastTransitionTime(oldTransitionTime)),
			MatchCondition(TestCondition().
				WithType("TestConditionFalse").
				WithStatus(metav1.ConditionFalse).
				WithObservedGeneration(10).
				WithReason("TestReasonFalse").
				WithMessage("TestMessageFalse").
				WithLastTransitionTime(obj.Status.LastReconcileTime)),
		))
	})

	It("should replace illegal characters in condition type and reason", func() {
		env := testutils.NewEnvironmentBuilder().WithFakeClient(coScheme).WithInitObjectPath("testdata", "test-02").WithDynamicObjectsWithStatus(&CustomObject{}).Build()
		obj := &CustomObject{}
		Expect(env.Client().Get(env.Ctx, controller.ObjectKey("status", "default"), obj)).To(Succeed())
		rr := &controller.ReconcileResult[*CustomObject]{
			Object: obj,
		}
		condFunc := controller.GenerateCreateConditionFunc(rr)

		condFunc("CondType :,;-_.Test02@", metav1.ConditionTrue, "Reason -.,:_Test93$", "Message")
		Expect(rr.Conditions).To(HaveLen(1))
		Expect(rr.Conditions[0].Type).To(Equal("CondType____-_.Test02_"))
		Expect(rr.Conditions[0].Reason).To(Equal("Reason___,:_Test93_"))
	})

	It("should not update disabled fields", func() {
		env := testutils.NewEnvironmentBuilder().WithFakeClient(coScheme).WithInitObjectPath("testdata", "test-02").WithDynamicObjectsWithStatus(&CustomObject{}).Build()
		obj := &CustomObject{}
		Expect(env.Client().Get(env.Ctx, controller.ObjectKey("status", "default"), obj)).To(Succeed())
		before := obj.DeepCopy()
		for _, disabledField := range controller.AllStatusFields() {
			By(fmt.Sprintf("Testing disabled field %s", disabledField))
			// reset object to remove changes from previous loop executions
			modified := obj.DeepCopy()
			obj.Status = before.Status
			Expect(env.Client().Status().Patch(env.Ctx, obj, client.MergeFrom(modified))).To(Succeed())
			rr := controller.ReconcileResult[*CustomObject]{
				Object:             obj,
				Conditions:         dummyConditions(),
				Reason:             "TestReason",
				Message:            "TestMessage",
				ConditionsToRemove: []string{"TestConditionTrue"},
			}
			su := preconfiguredStatusUpdaterBuilder().WithPhaseUpdateFunc(func(obj *CustomObject, rr controller.ReconcileResult[*CustomObject]) (string, error) {
				return PhaseSucceeded, nil
			}).WithoutFields(disabledField).Build()
			now := metav1.Now()
			res, err := su.UpdateStatus(env.Ctx, env.Client(), rr)
			Expect(res).To(Equal(rr.Result))

			Expect(err).ToNot(HaveOccurred())
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(obj), obj)).To(Succeed())

			if disabledField == controller.STATUS_FIELD_PHASE {
				Expect(obj.Status.Phase).To(Equal(before.Status.Phase))
			} else {
				Expect(obj.Status.Phase).To(Equal(PhaseSucceeded))
			}
			if disabledField == controller.STATUS_FIELD_OBSERVED_GENERATION {
				Expect(obj.Status.ObservedGeneration).To(Equal(before.Status.ObservedGeneration))
			} else {
				Expect(obj.Status.ObservedGeneration).To(Equal(obj.GetGeneration()))
			}
			if disabledField == controller.STATUS_FIELD_REASON {
				Expect(obj.Status.Reason).To(Equal(before.Status.Reason))
			} else {
				Expect(obj.Status.Reason).To(Equal(rr.Reason))
			}
			if disabledField == controller.STATUS_FIELD_MESSAGE {
				Expect(obj.Status.Message).To(Equal(before.Status.Message))
			} else {
				Expect(obj.Status.Message).To(Equal(rr.Message))
			}
			if disabledField == controller.STATUS_FIELD_LAST_RECONCILE_TIME {
				Expect(obj.Status.LastReconcileTime.Time).To(Equal(before.Status.LastReconcileTime.Time))
			} else {
				Expect(obj.Status.LastReconcileTime.Time).To(BeTemporally("~", now.Time, 1*time.Second))
			}
			if disabledField == controller.STATUS_FIELD_CONDITIONS {
				Expect(obj.Status.Conditions).To(Equal(before.Status.Conditions))
			} else {
				Expect(obj.Status.Conditions).To(ConsistOf(
					MatchCondition(TestCondition().
						WithType("TestConditionFalse").
						WithStatus(metav1.ConditionFalse).
						WithObservedGeneration(10).
						WithReason("TestReasonFalse").
						WithMessage("TestMessageFalse").
						WithLastTransitionTime(now).
						WithTimestampTolerance(1 * time.Second)),
				))
			}
		}
	})

	Context("Smart Requeue", func() {

		It("should add a requeueAfter duration if configured", func() {
			env := testutils.NewEnvironmentBuilder().WithFakeClient(coScheme).WithInitObjectPath("testdata", "test-02").WithDynamicObjectsWithStatus(&CustomObject{}).Build()
			obj := &CustomObject{}
			Expect(env.Client().Get(env.Ctx, controller.ObjectKey("status", "default"), obj)).To(Succeed())
			rr := controller.ReconcileResult[*CustomObject]{
				Object:       obj,
				Conditions:   dummyConditions(),
				SmartRequeue: controller.SR_RESET,
			}
			store := smartrequeue.NewStore(1*time.Second, 10*time.Second, 2.0)
			su := preconfiguredStatusUpdaterBuilder().WithPhaseUpdateFunc(func(obj *CustomObject, rr controller.ReconcileResult[*CustomObject]) (string, error) {
				return PhaseSucceeded, nil
			}).WithSmartRequeue(store).Build()
			res, err := su.UpdateStatus(env.Ctx, env.Client(), rr)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.RequeueAfter).To(Equal(1 * time.Second))
		})

		It("should keep the smaller requeueAfter duration if smart requeue and RequeueAfter in the ReconcileResult are set", func() {
			env := testutils.NewEnvironmentBuilder().WithFakeClient(coScheme).WithInitObjectPath("testdata", "test-02").WithDynamicObjectsWithStatus(&CustomObject{}).Build()
			obj := &CustomObject{}
			Expect(env.Client().Get(env.Ctx, controller.ObjectKey("status", "default"), obj)).To(Succeed())
			rr := controller.ReconcileResult[*CustomObject]{
				Object:     obj,
				Conditions: dummyConditions(),
				Result: ctrl.Result{
					RequeueAfter: 30 * time.Second,
				},
				SmartRequeue: controller.SR_RESET,
			}
			store := smartrequeue.NewStore(1*time.Second, 10*time.Second, 2.0)
			su := preconfiguredStatusUpdaterBuilder().WithPhaseUpdateFunc(func(obj *CustomObject, rr controller.ReconcileResult[*CustomObject]) (string, error) {
				return PhaseSucceeded, nil
			}).WithSmartRequeue(store).Build()
			res, err := su.UpdateStatus(env.Ctx, env.Client(), rr)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.RequeueAfter).To(Equal(1 * time.Second))

			rr.Result.RequeueAfter = 30 * time.Second
			store = smartrequeue.NewStore(1*time.Minute, 10*time.Minute, 2.0)
			su = preconfiguredStatusUpdaterBuilder().WithPhaseUpdateFunc(func(obj *CustomObject, rr controller.ReconcileResult[*CustomObject]) (string, error) {
				return PhaseSucceeded, nil
			}).WithSmartRequeue(store).Build()
			res, err = su.UpdateStatus(env.Ctx, env.Client(), rr)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.RequeueAfter).To(Equal(30 * time.Second))
		})

		It("should use the SmartRequeueConditional functions if specified", func() {
			env := testutils.NewEnvironmentBuilder().WithFakeClient(coScheme).WithInitObjectPath("testdata", "test-02").WithDynamicObjectsWithStatus(&CustomObject{}).Build()
			obj := &CustomObject{}
			Expect(env.Client().Get(env.Ctx, controller.ObjectKey("status", "default"), obj)).To(Succeed())
			rr := controller.ReconcileResult[*CustomObject]{
				Object:       obj,
				Conditions:   dummyConditions(),
				SmartRequeue: controller.SR_NO_REQUEUE,
			}
			store := smartrequeue.NewStore(1*time.Second, 10*time.Second, 2.0)
			su := preconfiguredStatusUpdaterBuilder().WithPhaseUpdateFunc(func(obj *CustomObject, rr controller.ReconcileResult[*CustomObject]) (string, error) {
				return PhaseSucceeded, nil
			}).WithSmartRequeue(store, func(rr controller.ReconcileResult[*CustomObject]) controller.SmartRequeueAction {
				return controller.SR_NO_REQUEUE
			}, func(rr controller.ReconcileResult[*CustomObject]) controller.SmartRequeueAction {
				return controller.SR_RESET
			}).Build()
			res, err := su.UpdateStatus(env.Ctx, env.Client(), rr)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.RequeueAfter).To(Equal(1 * time.Second))
		})

	})

	Context("GenerateCreateConditionFunc", func() {

		It("should add the condition to the given ReconcileResult", func() {
			rr := controller.ReconcileResult[*CustomObject]{
				Object: &CustomObject{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "test",
						Namespace:  "default",
						Generation: 15,
					},
				},
				Conditions: dummyConditions(),
			}
			createCon := controller.GenerateCreateConditionFunc(&rr)
			createCon("TestConditionFoo", metav1.ConditionTrue, "TestReasonFoo", "TestMessageFoo")
			Expect(rr.Conditions).To(ConsistOf(
				MatchCondition(TestConditionFromCondition(dummyConditions()[0])),
				MatchCondition(TestConditionFromCondition(dummyConditions()[1])),
				MatchCondition(TestConditionFromValues("TestConditionFoo", metav1.ConditionTrue, 15, "TestReasonFoo", "TestMessageFoo", metav1.Time{})),
			))
		})

		It("should create the condition list, if it is nil", func() {
			rr := controller.ReconcileResult[*CustomObject]{
				Object: &CustomObject{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "test",
						Namespace:  "default",
						Generation: 15,
					},
				},
			}
			createCon := controller.GenerateCreateConditionFunc(&rr)
			createCon("TestConditionFoo", metav1.ConditionTrue, "TestReasonFoo", "TestMessageFoo")
			Expect(rr.Conditions).To(ConsistOf(
				MatchCondition(TestConditionFromValues("TestConditionFoo", metav1.ConditionTrue, 15, "TestReasonFoo", "TestMessageFoo", metav1.Time{})),
			))
		})

	})

})

func preconfiguredStatusUpdaterBuilder() *controller.StatusUpdaterBuilder[*CustomObject] {
	nestedFields := controller.AllStatusFields()
	phaseIdx := slices.Index(nestedFields, controller.STATUS_FIELD_PHASE)
	if phaseIdx >= 0 {
		nestedFields = slices.Delete(nestedFields, phaseIdx, phaseIdx+1)
	}
	return controller.NewStatusUpdaterBuilder[*CustomObject]().WithNestedStruct("CommonStatus", nestedFields...).WithPhaseUpdateFunc(dummyPhaseUpdateFunc).WithConditionUpdater(true)
}

func dummyConditions() []metav1.Condition {
	return []metav1.Condition{
		{
			Type:               "TestConditionTrue",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: 10,
			Reason:             "TestReasonTrue",
			Message:            "TestMessageTrue",
		},
		{
			Type:               "TestConditionFalse",
			Status:             metav1.ConditionFalse,
			ObservedGeneration: 10,
			Reason:             "TestReasonFalse",
			Message:            "TestMessageFalse",
		},
	}
}

func dummyPhaseUpdateFunc(obj *CustomObject, rr controller.ReconcileResult[*CustomObject]) (string, error) {
	if rr.ReconcileError != nil {
		return PhaseFailed, nil
	}
	if len(obj.Status.Conditions) > 0 {
		for _, con := range obj.Status.Conditions {
			if con.Status != metav1.ConditionTrue {
				return PhaseFailed, nil
			}
		}
	}
	return PhaseSucceeded, nil
}

/////////////////////////////////
// DUMMY OBJECT IMPLEMENTATION //
/////////////////////////////////

var _ client.Object = &CustomObject{}

// This is a dummy k8s object implementation to test the status updater on.
type CustomObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomObjectSpec   `json:"spec,omitempty"`
	Status CustomObjectStatus `json:"status,omitempty"`
}

type CustomObjectSpec struct {
}

type CustomObjectStatus struct {
	CommonStatus `json:",inline"`

	// Phase is the current phase of the cluster.
	Phase string `json:"phase"`
}

const (
	PhaseSucceeded = "Succeeded"
	PhaseFailed    = "Failed"
)

type CommonStatus struct {
	ObservedGeneration int64 `json:"observedGeneration"`

	// LastReconcileTime is the time when the resource was last reconciled by the controller.
	LastReconcileTime metav1.Time `json:"lastReconcileTime"`

	// Reason is expected to contain a CamelCased string that provides further information in a machine-readable format.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message contains further details in a human-readable format.
	// +optional
	Message string `json:"message,omitempty"`

	// Conditions contains the conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type CustomObjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomObject `json:"items"`
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonStatus) DeepCopyInto(out *CommonStatus) {
	*out = *in
	in.LastReconcileTime.DeepCopyInto(&out.LastReconcileTime)
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonStatus.
func (in *CommonStatus) DeepCopy() *CommonStatus {
	if in == nil {
		return nil
	}
	out := new(CommonStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomObject) DeepCopyInto(out *CustomObject) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomObject.
func (in *CustomObject) DeepCopy() *CustomObject {
	if in == nil {
		return nil
	}
	out := new(CustomObject)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomObject) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomObjectSpec) DeepCopyInto(out *CustomObjectSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomObjectSpec.
func (in *CustomObjectSpec) DeepCopy() *CustomObjectSpec {
	if in == nil {
		return nil
	}
	out := new(CustomObjectSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomObjectStatus) DeepCopyInto(out *CustomObjectStatus) {
	*out = *in
	in.CommonStatus.DeepCopyInto(&out.CommonStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomObjectStatus.
func (in *CustomObjectStatus) DeepCopy() *CustomObjectStatus {
	if in == nil {
		return nil
	}
	out := new(CustomObjectStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomObjectList) DeepCopyInto(out *CustomObjectList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CustomObject, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomObjectList.
func (in *CustomObjectList) DeepCopy() *CustomObjectList {
	if in == nil {
		return nil
	}
	out := new(CustomObjectList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomObjectList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&CustomObject{}, &CustomObjectList{})
	coScheme = runtime.NewScheme()
	err := SchemeBuilder.AddToScheme(coScheme)
	if err != nil {
		panic(err)
	}
}

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "testing.openmcp.cloud", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}
)
