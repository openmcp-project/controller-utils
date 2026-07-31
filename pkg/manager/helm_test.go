package externalsecrets

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestExtractHelmValues(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		values  *apiextensionsv1.JSON
		want    *HelmValues
		wantErr bool
	}{
		{
			name: "Extract Namespace",
			values: &apiextensionsv1.JSON{
				Raw: []byte(`{"namespaceOverride": "eso-system"}`),
			},
			want: &HelmValues{
				NamespaceOverride: "eso-system",
			},
			wantErr: false,
		},
		{
			name: "Extract ImagePullSecrets",
			values: &apiextensionsv1.JSON{
				Raw: []byte(`{"global": {"imagePullSecrets": [{"name": "test"}]}}`),
			},
			want: &HelmValues{
				Global: Global{
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "test",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Ignore unknown values",
			values: &apiextensionsv1.JSON{
				Raw: []byte(`{"replicaCount": 1}`),
			},
			want:    &HelmValues{},
			wantErr: false,
		},
		{
			name: "Error on invalid JSON",
			values: &apiextensionsv1.JSON{
				Raw: []byte("invalid json"),
			},
			want:    &HelmValues{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := ExtractHelmValues(tt.values)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ExtractHelmValues() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ExtractHelmValues() succeeded unexpectedly")
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
