package externalsecrets

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// HelmValues define the helm values that are explicitly processed during reconciliation
type HelmValues struct {
	NamespaceOverride string `json:"namespaceOverride,omitempty"`
	Global            Global `json:"global,omitempty"`
}

// Global define the global settings that are explicitly process during reconciliation
type Global struct {
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// ExtractHelmValues extract helm values required for processing
func ExtractHelmValues(values *apiextensionsv1.JSON) (*HelmValues, error) {
	if values == nil || len(values.Raw) == 0 {
		return &HelmValues{}, nil
	}

	vals := &HelmValues{}
	if err := json.Unmarshal(values.Raw, vals); err != nil {
		return nil, err
	}

	return vals, nil
}
