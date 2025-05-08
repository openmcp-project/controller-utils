package resources

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type metadataMutator struct {
	Labels      map[string]string
	Annotations map[string]string
}

var _ Mutator[client.Object] = &metadataMutator{}

func NewMetadataMutator(labels map[string]string, annotations map[string]string) Mutator[client.Object] {
	return &metadataMutator{
		Labels:      labels,
		Annotations: annotations,
	}
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
		for k, v := range m.Labels {
			res.GetLabels()[k] = v
		}
	}

	if m.Annotations != nil {
		if res.GetAnnotations() == nil {
			res.SetAnnotations(make(map[string]string))
		}

		for k, v := range m.Annotations {
			res.GetAnnotations()[k] = v
		}
	}
	return nil
}
