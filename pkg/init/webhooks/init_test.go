package webhooks

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func setEnv() {
	os.Setenv("WEBHOOK_SERVICE_NAME", "myservice")
	os.Setenv("WEBHOOK_SERVICE_NAMESPACE", "mynamespace")
	os.Setenv("WEBHOOK_SECRET_NAME", "mysecret")
	os.Setenv("WEBHOOK_SECRET_NAMESPACE", "mynamespace")
}

func Test_GenerateCertificate(t *testing.T) {
	testCases := []struct {
		desc     string
		setup    func(ctx context.Context, c client.Client) error
		validate func(ctx context.Context, c client.Client, t *testing.T, testErr error) error
		options  []CertOption
	}{
		{
			desc: "should generate certificate",
			setup: func(ctx context.Context, c client.Client) error {
				setEnv()

				ns := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mynamespace",
					},
				}
				return c.Create(ctx, ns)
			},
			validate: func(ctx context.Context, c client.Client, t *testing.T, testErr error) error {
				assert.NoError(t, testErr)

				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mysecret",
						Namespace: "mynamespace",
					},
				}
				if err := c.Get(ctx, client.ObjectKeyFromObject(secret), secret); err != nil {
					return err
				}

				assert.NotEmpty(t, secret.Data[corev1.TLSCertKey])
				assert.NotEmpty(t, secret.Data[corev1.TLSPrivateKeyKey])
				return nil
			},
		},
		{
			desc: "should generate certificate with custom object names",
			options: []CertOption{
				WithWebhookSecret{Name: "myothersecret", Namespace: "myothernamespace"},
				WithWebhookService{Name: "myotherservice", Namespace: "myothernamespace"},
				WithAdditionalDNSNames{"some.other.name.example.com"},
			},
			setup: func(ctx context.Context, c client.Client) error {
				setEnv()

				ns := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "myothernamespace",
					},
				}
				return c.Create(ctx, ns)
			},
			validate: func(ctx context.Context, c client.Client, t *testing.T, testErr error) error {
				assert.NoError(t, testErr)

				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myothersecret",
						Namespace: "myothernamespace",
					},
				}
				if err := c.Get(ctx, client.ObjectKeyFromObject(secret), secret); err != nil {
					return err
				}

				assert.NotEmpty(t, secret.Data[corev1.TLSCertKey])
				assert.NotEmpty(t, secret.Data[corev1.TLSPrivateKeyKey])
				return nil
			},
		},
		{
			desc: "should not override existing certificate",
			setup: func(ctx context.Context, c client.Client) error {
				setEnv()

				ns := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mynamespace",
					},
				}
				if err := c.Create(ctx, ns); err != nil {
					return err
				}

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
			},
			validate: func(ctx context.Context, c client.Client, t *testing.T, testErr error) error {
				assert.NoError(t, testErr)

				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mysecret",
						Namespace: "mynamespace",
					},
				}
				if err := c.Get(ctx, client.ObjectKeyFromObject(secret), secret); err != nil {
					return err
				}

				assert.Equal(t, []byte("abc"), secret.Data[corev1.TLSCertKey])
				assert.Equal(t, []byte("def"), secret.Data[corev1.TLSPrivateKeyKey])
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

			testErr := GenerateCertificate(ctx, c, tC.options...)

			if err := tC.validate(ctx, c, t, testErr); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_Install(t *testing.T) {
	testCases := []struct {
		desc     string
		setup    func(ctx context.Context, c client.Client) error
		validate func(ctx context.Context, c client.Client, t *testing.T, testErr error) error
		options  []InstallOption
	}{
		{
			desc: "should create webhook configurations for TestObj",
			setup: func(ctx context.Context, c client.Client) error {
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
			},
			validate: func(ctx context.Context, c client.Client, t *testing.T, testErr error) error {
				assert.NoError(t, testErr)

				vwc := &admissionregistrationv1.ValidatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: generateValidateName(testObjGVK),
					},
				}
				err := c.Get(ctx, client.ObjectKeyFromObject(vwc), vwc)
				assert.NoError(t, err)
				assert.Len(t, vwc.Webhooks, 1)
				assert.Equal(t, vwc.Webhooks[0].ClientConfig.CABundle, []byte("abc"))
				assert.Equal(t, vwc.Webhooks[0].ClientConfig.Service, &admissionregistrationv1.ServiceReference{
					Name:      "myservice",
					Namespace: "mynamespace",
					Path:      ptr.To(generateValidatePath(testObjGVK)),
					Port:      ptr.To[int32](443),
				})

				mwc := &admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: generateMutateName(testObjGVK),
					},
				}
				err = c.Get(ctx, client.ObjectKeyFromObject(mwc), mwc)
				assert.NoError(t, err)
				assert.Len(t, mwc.Webhooks, 1)
				assert.Equal(t, mwc.Webhooks[0].ClientConfig.CABundle, []byte("abc"))
				assert.Equal(t, mwc.Webhooks[0].ClientConfig.Service, &admissionregistrationv1.ServiceReference{
					Name:      "myservice",
					Namespace: "mynamespace",
					Path:      ptr.To(generateMutatePath(testObjGVK)),
					Port:      ptr.To[int32](443),
				})

				return nil
			},
		},
		{
			desc: "should create webhook configurations for TestObj with custom values",
			options: []InstallOption{
				WithoutCA,
				WithCustomBaseURL("https://webhooks.example.com"),
			},
			setup: func(ctx context.Context, c client.Client) error { return nil },
			validate: func(ctx context.Context, c client.Client, t *testing.T, testErr error) error {
				assert.NoError(t, testErr)

				vwc := &admissionregistrationv1.ValidatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: generateValidateName(testObjGVK),
					},
				}
				err := c.Get(ctx, client.ObjectKeyFromObject(vwc), vwc)
				assert.NoError(t, err)
				assert.Len(t, vwc.Webhooks, 1)
				assert.Nil(t, vwc.Webhooks[0].ClientConfig.CABundle)
				assert.Nil(t, vwc.Webhooks[0].ClientConfig.Service)
				assert.Equal(t, *vwc.Webhooks[0].ClientConfig.URL, "https://webhooks.example.com"+generateValidatePath(testObjGVK))

				mwc := &admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: generateMutateName(testObjGVK),
					},
				}
				err = c.Get(ctx, client.ObjectKeyFromObject(mwc), mwc)
				assert.NoError(t, err)
				assert.Len(t, mwc.Webhooks, 1)
				assert.Len(t, mwc.Webhooks, 1)
				assert.Nil(t, mwc.Webhooks[0].ClientConfig.CABundle)
				assert.Nil(t, mwc.Webhooks[0].ClientConfig.Service)
				assert.Equal(t, *mwc.Webhooks[0].ClientConfig.URL, "https://webhooks.example.com"+generateMutatePath(testObjGVK))

				return nil
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			c := fake.NewClientBuilder().Build()
			c.Scheme().AddKnownTypes(groupVersion, &TestObj{})
			ctx := context.Background()

			if err := tC.setup(ctx, c); err != nil {
				t.Fatal(err)
			}

			apiTypes := []APITypes{
				{
					Obj:       &TestObj{},
					Validator: true,
					Defaulter: true,
				},
			}
			testErr := Install(ctx, c, c.Scheme(), apiTypes, tC.options...)

			if err := tC.validate(ctx, c, t, testErr); err != nil {
				t.Fatal(err)
			}
		})
	}
}

var (
	groupVersion = schema.GroupVersion{Group: "example.org", Version: "v1alpha1"}
	testObjGVK   = groupVersion.WithKind("TestObj")
)

var _ client.Object = &TestObj{}
var _ admission.Defaulter[*TestObj] = &TestObj{}
var _ admission.Validator[*TestObj] = &TestObj{}

type TestObj struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// Default implements admission.Defaulter.
func (*TestObj) Default(_ context.Context, _ *TestObj) error {
	panic("unimplemented")
}

// ValidateCreate implements admission.Validator.
func (*TestObj) ValidateCreate(_ context.Context, _ *TestObj) (warnings admission.Warnings, err error) {
	panic("unimplemented")
}

// ValidateDelete implements admission.Validator.
func (*TestObj) ValidateDelete(_ context.Context, _ *TestObj) (warnings admission.Warnings, err error) {
	panic("unimplemented")
}

// ValidateUpdate implements admission.Validator.
func (*TestObj) ValidateUpdate(_ context.Context, _ *TestObj, _ *TestObj) (warnings admission.Warnings, err error) {
	panic("unimplemented")
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TestObj) DeepCopyInto(out *TestObj) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CoffeeBean.
func (in *TestObj) DeepCopy() *TestObj {
	if in == nil {
		return nil
	}
	out := new(TestObj)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TestObj) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
