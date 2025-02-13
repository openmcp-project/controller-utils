package crds

import (
	"context"
	"embed"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	//go:embed "testdata/crds"
	crdFiles embed.FS
)

func setEnv() {
	os.Setenv("WEBHOOK_SERVICE_NAME", "myservice")
	os.Setenv("WEBHOOK_SERVICE_NAMESPACE", "mynamespace")
	os.Setenv("WEBHOOK_SECRET_NAME", "mysecret")
	os.Setenv("WEBHOOK_SECRET_NAMESPACE", "mynamespace")
}

func createWebhookSecret(ctx context.Context, c client.Client) error {
	setEnv()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysecret",
			Namespace: "mynamespace",
		},
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte("abc"),
			corev1.TLSPrivateKeyKey: []byte("def"),
		},
	}
	return c.Create(ctx, secret)
}

func Test_Install(t *testing.T) {
	testCases := []struct {
		desc     string
		setup    func(ctx context.Context, c client.Client) error
		validate func(ctx context.Context, c client.Client, t *testing.T, testErr error) error
		options  []installOption
	}{
		{
			desc:  "should create CRD with and without conversion",
			setup: createWebhookSecret,
			validate: func(ctx context.Context, c client.Client, t *testing.T, testErr error) error {
				assert.NoError(t, testErr)

				crdNoConversion := &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "crontabs.stable.example.com",
					},
				}
				err := c.Get(ctx, client.ObjectKeyFromObject(crdNoConversion), crdNoConversion)
				assert.NoError(t, err)
				assert.Equal(t, crdNoConversion.Spec.Conversion, (*v1.CustomResourceConversion)(nil))

				crdConversion := &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "crontabsconversion.stable.example.com",
					},
				}
				err = c.Get(ctx, client.ObjectKeyFromObject(crdConversion), crdConversion)
				assert.NoError(t, err)
				assert.Equal(t, crdConversion.Spec.Conversion, &apiextensionsv1.CustomResourceConversion{
					Strategy: apiextensionsv1.WebhookConverter,
					Webhook: &apiextensionsv1.WebhookConversion{
						ConversionReviewVersions: []string{"v1"},
						ClientConfig: &apiextensionsv1.WebhookClientConfig{
							URL:      nil,
							CABundle: []byte("abc"),
							Service: &apiextensionsv1.ServiceReference{
								Name:      "myservice",
								Namespace: "mynamespace",
								Path:      ptr.To("/convert"),
								Port:      ptr.To[int32](443),
							},
						},
					},
				})

				return nil
			},
		},
		{
			desc: "should create CRD with custom webhook for conversion",
			options: []installOption{
				WithWebhookService{Name: "myotherservice", Namespace: "myothernamespace"},
				WithWebhookServicePort(1234),
			},
			setup: createWebhookSecret,
			validate: func(ctx context.Context, c client.Client, t *testing.T, testErr error) error {
				assert.NoError(t, testErr)

				crdConversion := &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "crontabsconversion.stable.example.com",
					},
				}
				err := c.Get(ctx, client.ObjectKeyFromObject(crdConversion), crdConversion)
				assert.NoError(t, err)
				assert.Equal(t, crdConversion.Spec.Conversion, &apiextensionsv1.CustomResourceConversion{
					Strategy: apiextensionsv1.WebhookConverter,
					Webhook: &apiextensionsv1.WebhookConversion{
						ConversionReviewVersions: []string{"v1"},
						ClientConfig: &apiextensionsv1.WebhookClientConfig{
							URL:      nil,
							CABundle: []byte("abc"),
							Service: &apiextensionsv1.ServiceReference{
								Name:      "myotherservice",
								Namespace: "myothernamespace",
								Path:      ptr.To("/convert"),
								Port:      ptr.To[int32](1234),
							},
						},
					},
				})

				return nil
			},
		},
		{
			desc: "should create CRD with custom URL for conversion",
			options: []installOption{
				WithCustomBaseURL("https://webhooks.example.com"),
				WithoutCA,
			},
			setup: createWebhookSecret,
			validate: func(ctx context.Context, c client.Client, t *testing.T, testErr error) error {
				assert.NoError(t, testErr)

				crdConversion := &apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "crontabsconversion.stable.example.com",
					},
				}
				err := c.Get(ctx, client.ObjectKeyFromObject(crdConversion), crdConversion)
				assert.NoError(t, err)
				assert.Equal(t, crdConversion.Spec.Conversion, &apiextensionsv1.CustomResourceConversion{
					Strategy: apiextensionsv1.WebhookConverter,
					Webhook: &apiextensionsv1.WebhookConversion{
						ConversionReviewVersions: []string{"v1"},
						ClientConfig: &apiextensionsv1.WebhookClientConfig{
							URL:      ptr.To("https://webhooks.example.com/convert"),
							CABundle: nil,
							Service:  nil,
						},
					},
				})

				return nil
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			c := fake.NewClientBuilder().Build()
			ctx := context.Background()

			if err := tC.setup(ctx, c); err != nil {
				t.Fatal(err)
			}

			testErr := Install(ctx, c, crdFiles, tC.options...)

			if err := tC.validate(ctx, c, t, testErr); err != nil {
				t.Fatal(err)
			}
		})
	}
}
