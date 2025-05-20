package resources

import (
	"maps"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type metadataMutator struct {
	Labels          map[string]string
	Annotations     map[string]string
	OwnerReferences []metav1.OwnerReference
	Finalizers      []string
}

type MetadataMutator interface {
	Mutator[client.Object]
	WithOwnerReferences(ownerReferences []metav1.OwnerReference) MetadataMutator
	WithFinalizers(finalizers []string) MetadataMutator
	WithLabels(labels map[string]string) MetadataMutator
	WithAnnotations(annotations map[string]string) MetadataMutator
}

var _ Mutator[client.Object] = &metadataMutator{}
var _ MetadataMutator = &metadataMutator{}

func NewMetadataMutator() MetadataMutator {
	return &metadataMutator{}
}

func (m *metadataMutator) WithLabels(labels map[string]string) MetadataMutator {
	m.Labels = labels
	return m
}

func (m *metadataMutator) WithAnnotations(annotations map[string]string) MetadataMutator {
	m.Annotations = annotations
	return m
}

func (m *metadataMutator) WithOwnerReferences(ownerReferences []metav1.OwnerReference) MetadataMutator {
	m.OwnerReferences = ownerReferences
	return m
}

func (m *metadataMutator) WithFinalizers(finalizers []string) MetadataMutator {
	m.Finalizers = finalizers
	return m
}

func (m *metadataMutator) String() string {
	return "metadata"
}

func (m *metadataMutator) Empty() client.Object {
	return nil
}

func (m *metadataMutator) Mutate(res client.Object) error {
	if m.Labels != nil {
		if res.GetLabels() == nil {
			res.SetLabels(make(map[string]string))
		}
		maps.Copy(res.GetLabels(), m.Labels)
	}

	if m.Annotations != nil {
		if res.GetAnnotations() == nil {
			res.SetAnnotations(make(map[string]string))
		}
		maps.Copy(res.GetAnnotations(), m.Annotations)
	}

	if m.OwnerReferences != nil {
		// ensure that all owner references in m are also in res
		if len(res.GetOwnerReferences()) == 0 {
			res.SetOwnerReferences(make([]metav1.OwnerReference, len(m.OwnerReferences)))
			for i, ownerRef := range m.OwnerReferences {
				res.GetOwnerReferences()[i] = *ownerRef.DeepCopy()
			}
		} else {
			for _, ownerRef := range m.OwnerReferences {
				found := false
				for _, existingRef := range res.GetOwnerReferences() {
					if ownerRef.UID == existingRef.UID {
						found = true
						break
					}
				}
				if !found {
					res.SetOwnerReferences(append(res.GetOwnerReferences(), *ownerRef.DeepCopy()))
				}
			}
		}
	}

	if m.Finalizers != nil {
		if len(res.GetFinalizers()) == 0 {
			res.SetFinalizers(make([]string, len(m.Finalizers)))
			copy(res.GetFinalizers(), m.Finalizers)
		} else {
			for _, fin := range m.Finalizers {
				found := false
				for _, existingFin := range res.GetFinalizers() {
					if fin == existingFin {
						found = true
						break
					}
				}
				if !found {
					res.SetFinalizers(append(res.GetFinalizers(), fin))
				}
			}
		}
	}

	return nil
}

func (m *metadataMutator) MetadataMutator() MetadataMutator {
	return m
}
