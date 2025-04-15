package controller_test

import (
	"fmt"
	"slices"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/openmcp-project/controller-utils/pkg/testing/matchers"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/scheme"

	"github.com/openmcp-project/controller-utils/pkg/conditions"
	"github.com/openmcp-project/controller-utils/pkg/controller"
	"github.com/openmcp-project/controller-utils/pkg/errors"
	testutils "github.com/openmcp-project/controller-utils/pkg/testing"
)

var coScheme *runtime.Scheme

var _ = Describe("Status Updater", func() {

	It("should update an empty status", func() {
		env := testutils.NewEnvironmentBuilder().WithFakeClient(coScheme).WithInitObjectPath("testdata", "test-02").WithDynamicObjectsWithStatus(&CustomObject{}).Build()
		obj := &CustomObject{}
		Expect(env.Client().Get(env.Ctx, testutils.ObjectKey("nostatus", "default"), obj)).To(Succeed())
		rr := controller.ReconcileResult[*CustomObject, ConditionStatus]{
			Object:         obj,
			ReconcileError: errors.WithReason(fmt.Errorf("test error"), "TestError"),
			Conditions:     dummyConditions(),
		}
		su := preconfiguredStatusUpdaterBuilder().Build()
		now := time.Now()
		Expect(su.UpdateStatus(env.Ctx, env.Client(), rr)).To(Succeed())
		Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(obj), obj)).To(Succeed())

		Expect(obj.Status.Phase).To(Equal(PhaseFailed))
		Expect(obj.Status.ObservedGeneration).To(Equal(obj.GetGeneration()))
		Expect(obj.Status.Reason).To(Equal("TestError"))
		Expect(obj.Status.Message).To(ContainSubstring("test error"))
		Expect(obj.Status.LastReconcileTime.Time).To(BeTemporally("~", now, 1*time.Second))
		Expect(obj.Status.Conditions).To(ConsistOf(
			MatchCondition(NewConditionImpl[ConditionStatus]().
				WithType("TestConditionTrue").
				WithStatus(ConditionStatusTrue).
				WithReason("TestReasonTrue").
				WithMessage("TestMessageTrue").
				WithLastTransitionTime(obj.Status.LastReconcileTime.Time)),
			MatchCondition(NewConditionImpl[ConditionStatus]().
				WithType("TestConditionFalse").
				WithStatus(ConditionStatusFalse).
				WithReason("TestReasonFalse").
				WithMessage("TestMessageFalse").
				WithLastTransitionTime(obj.Status.LastReconcileTime.Time)),
		))
	})

})

func preconfiguredStatusUpdaterBuilder() *controller.StatusUpdaterBuilder[*CustomObject, CustomObjectPhase, ConditionStatus] {
	nestedFields := controller.AllStatusFields()
	phaseIdx := slices.Index(nestedFields, controller.STATUS_FIELD_PHASE)
	if phaseIdx >= 0 {
		nestedFields = slices.Delete(nestedFields, phaseIdx, phaseIdx+1)
	}
	return controller.NewStatusUpdaterBuilder[*CustomObject, CustomObjectPhase, ConditionStatus]().WithNestedStruct("CommonStatus", nestedFields...).WithPhaseUpdateFunc(dummyPhaseUpdateFunc).WithConditionUpdater(func() conditions.Condition[ConditionStatus] { return &Condition{} }, true)
}

func dummyConditions() []conditions.Condition[ConditionStatus] {
	return []conditions.Condition[ConditionStatus]{
		&Condition{
			Type:    "TestConditionTrue",
			Status:  ConditionStatusTrue,
			Reason:  "TestReasonTrue",
			Message: "TestMessageTrue",
		},
		&Condition{
			Type:    "TestConditionFalse",
			Status:  ConditionStatusFalse,
			Reason:  "TestReasonFalse",
			Message: "TestMessageFalse",
		},
	}
}

func dummyPhaseUpdateFunc(obj *CustomObject, rr controller.ReconcileResult[*CustomObject, ConditionStatus]) (CustomObjectPhase, error) {
	if rr.ReconcileError != nil {
		return PhaseFailed, nil
	}
	if len(obj.Status.Conditions) > 0 {
		for _, con := range obj.Status.Conditions {
			if con.GetStatus() != ConditionStatusTrue {
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
	Phase CustomObjectPhase `json:"phase"`
}

type CustomObjectPhase string

const (
	PhaseSucceeded CustomObjectPhase = "Succeeded"
	PhaseFailed    CustomObjectPhase = "Failed"
)

type ConditionStatus string

const (
	ConditionStatusTrue    ConditionStatus = "True"
	ConditionStatusFalse   ConditionStatus = "False"
	ConditionStatusUnknown ConditionStatus = "Unknown"
)

type Condition struct {
	// Type is the type of the condition.
	// Must be unique within the resource.
	Type string `json:"type"`

	// Status is the status of the condition.
	Status ConditionStatus `json:"status"`

	// Reason is expected to contain a CamelCased string that provides further information regarding the condition.
	// It should have a fixed value set (like an enum) to be machine-readable. The value set depends on the condition type.
	// It is optional, but should be filled at least when Status is not "True".
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message contains further details regarding the condition.
	// It is meant for human users, Reason should be used for programmatic evaluation instead.
	// It is optional, but should be filled at least when Status is not "True".
	// +optional
	Message string `json:"message,omitempty"`

	// LastTransitionTime specifies the time when this condition's status last changed.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// Implement the Condition interface from our controller-utils library
func (c *Condition) GetType() string {
	return c.Type
}
func (c *Condition) SetType(t string) {
	c.Type = t
}
func (c *Condition) GetStatus() ConditionStatus {
	return c.Status
}
func (c *Condition) SetStatus(s ConditionStatus) {
	c.Status = s
}
func (c *Condition) GetReason() string {
	return c.Reason
}
func (c *Condition) SetReason(r string) {
	c.Reason = r
}
func (c *Condition) GetMessage() string {
	return c.Message
}
func (c *Condition) SetMessage(m string) {
	c.Message = m
}
func (c *Condition) GetLastTransitionTime() time.Time {
	return c.LastTransitionTime.Time
}
func (c *Condition) SetLastTransitionTime(t time.Time) {
	c.LastTransitionTime = metav1.NewTime(t)
}

// ConditionList is a list of Conditions.
type ConditionList []*Condition

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
	Conditions ConditionList `json:"conditions,omitempty"`
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
		*out = make(ConditionList, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Condition)
				(*in).DeepCopyInto(*out)
			}
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
func (in *Condition) DeepCopyInto(out *Condition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Condition.
func (in *Condition) DeepCopy() *Condition {
	if in == nil {
		return nil
	}
	out := new(Condition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in ConditionList) DeepCopyInto(out *ConditionList) {
	{
		in := &in
		*out = make(ConditionList, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Condition)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConditionList.
func (in ConditionList) DeepCopy() ConditionList {
	if in == nil {
		return nil
	}
	out := new(ConditionList)
	in.DeepCopyInto(out)
	return *out
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
