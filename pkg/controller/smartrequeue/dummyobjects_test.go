package smartrequeue

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type dummyObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
}

var _ client.Object = &dummyObject{}

func (d *dummyObject) DeepCopyObject() runtime.Object {
	return &dummyObject{
		TypeMeta:   d.TypeMeta,
		ObjectMeta: *d.DeepCopy(),
	}
}

type anotherDummyObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
}

var _ client.Object = &anotherDummyObject{}

func (d *anotherDummyObject) DeepCopyObject() runtime.Object {
	return &anotherDummyObject{
		TypeMeta:   d.TypeMeta,
		ObjectMeta: *d.DeepCopy(),
	}
}
