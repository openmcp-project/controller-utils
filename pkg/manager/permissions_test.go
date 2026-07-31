package externalsecrets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	externalsecretsoperatorsv1alpha1 "github.com/openmcp-project/service-provider-external-secrets/api/v1alpha1"
)

func TestResolveEsoNamespace(t *testing.T) {
	tests := []struct {
		name    string
		obj     *externalsecretsoperatorsv1alpha1.ExternalSecretsOperator
		pc      *externalsecretsoperatorsv1alpha1.ProviderConfig
		want    any
		wantErr bool
	}{
		{
			name: "No Namespace Override",
			obj: &externalsecretsoperatorsv1alpha1.ExternalSecretsOperator{
				Spec: externalsecretsoperatorsv1alpha1.ExternalSecretsOperatorSpec{
					Version: "v1.0.0",
				},
			},
			pc: &externalsecretsoperatorsv1alpha1.ProviderConfig{
				Spec: externalsecretsoperatorsv1alpha1.ProviderConfigSpec{
					Versions: []externalsecretsoperatorsv1alpha1.ExternalSecretsVersion{
						{
							Version: "v1.0.0",
						},
					},
				},
			},
			want: ESONamespace(DefaultNamespace),
		},
		{
			name: "Namespace Override",
			obj: &externalsecretsoperatorsv1alpha1.ExternalSecretsOperator{
				Spec: externalsecretsoperatorsv1alpha1.ExternalSecretsOperatorSpec{
					Version: "v1.0.0",
				},
			},
			pc: &externalsecretsoperatorsv1alpha1.ProviderConfig{
				Spec: externalsecretsoperatorsv1alpha1.ProviderConfigSpec{
					Versions: []externalsecretsoperatorsv1alpha1.ExternalSecretsVersion{
						{
							Version: "v1.0.0",
							HelmValues: &apiextensionsv1.JSON{
								Raw: []byte(`{"namespaceOverride": "eso-system"}`),
							},
						},
					},
				},
			},
			want: ESONamespace("eso-system"),
		},
		{
			name: "Empty String Namespace Override",
			obj: &externalsecretsoperatorsv1alpha1.ExternalSecretsOperator{
				Spec: externalsecretsoperatorsv1alpha1.ExternalSecretsOperatorSpec{
					Version: "v1.0.0",
				},
			},
			pc: &externalsecretsoperatorsv1alpha1.ProviderConfig{
				Spec: externalsecretsoperatorsv1alpha1.ProviderConfigSpec{
					Versions: []externalsecretsoperatorsv1alpha1.ExternalSecretsVersion{
						{
							Version: "v1.0.0",
							HelmValues: &apiextensionsv1.JSON{
								Raw: []byte(`{"namespaceOverride": ""}`),
							},
						},
					},
				},
			},
			want: ESONamespace(DefaultNamespace),
		},
		{
			name:    "ProviderConfig nil",
			obj:     &externalsecretsoperatorsv1alpha1.ExternalSecretsOperator{},
			pc:      nil,
			wantErr: true,
		},
		{
			name:    "Obj nil",
			obj:     nil,
			pc:      &externalsecretsoperatorsv1alpha1.ProviderConfig{},
			wantErr: true,
		},
		{
			name: "Version Not Found",
			obj: &externalsecretsoperatorsv1alpha1.ExternalSecretsOperator{
				Spec: externalsecretsoperatorsv1alpha1.ExternalSecretsOperatorSpec{
					Version: "v2.0.0",
				},
			},
			pc: &externalsecretsoperatorsv1alpha1.ProviderConfig{
				Spec: externalsecretsoperatorsv1alpha1.ProviderConfigSpec{
					Versions: []externalsecretsoperatorsv1alpha1.ExternalSecretsVersion{
						{
							Version: "v1.0.0",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid Helm Values",
			obj: &externalsecretsoperatorsv1alpha1.ExternalSecretsOperator{
				Spec: externalsecretsoperatorsv1alpha1.ExternalSecretsOperatorSpec{
					Version: "v1.0.0",
				},
			},
			pc: &externalsecretsoperatorsv1alpha1.ProviderConfig{
				Spec: externalsecretsoperatorsv1alpha1.ProviderConfigSpec{
					Versions: []externalsecretsoperatorsv1alpha1.ExternalSecretsVersion{
						{
							Version: "v1.0.0",
							HelmValues: &apiextensionsv1.JSON{
								Raw: []byte("invalid json"),
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := ResolveEsoNamespace(context.Background(), tt.obj, tt.pc)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("TestResolveEsoNamespace() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("TestResolveEsoNamespace() succeeded unexpectedly")
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
