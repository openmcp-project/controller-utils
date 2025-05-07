package resources

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type baseMutator struct {
	Labels      map[string]string
	Annotations map[string]string
}

var _ Mutator[client.Object] = &baseMutator{}

func NewBaseMutator(labels, annotations map[string]string) Mutator[client.Object] {
	return &baseMutator{
		Labels:      labels,
		Annotations: annotations,
	}
}
