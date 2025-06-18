package clusteraccess_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openmcp-project/controller-utils/pkg/clusteraccess"
	"github.com/openmcp-project/controller-utils/pkg/pairs"
	testutils "github.com/openmcp-project/controller-utils/pkg/testing"
)

var testLabelsMap = map[string]string{
	"label1": "value1",
	"label2": "value2",
}
var testLabelsList = pairs.MapToPairs(testLabelsMap)

var _ = Describe("ClusterAccess", func() {

	Context("EnsureNamespace", func() {

		It("should create a namespace if it does not exist", func() {
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).Build()
			ns := &corev1.Namespace{}
			ns.SetName("testns")
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(MatchError(apierrors.IsNotFound, "namespace should not exist"))
			ns, err := clusteraccess.EnsureNamespace(env.Ctx, env.Client(), ns.Name, testLabelsList...)
			Expect(err).ToNot(HaveOccurred())
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
			Expect(ns.Labels).To(BeEquivalentTo(testLabelsMap))
		})

		It("should throw an error if the namespace exists, but is missing the expected labels", func() {
			ns := &corev1.Namespace{}
			ns.SetName("testns")
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).WithInitObjects(ns).Build()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
			Expect(ns.Labels).To(BeEmpty())
			_, err := clusteraccess.EnsureNamespace(env.Ctx, env.Client(), ns.Name, testLabelsList...)
			Expect(err).To(HaveOccurred())
		})

		It("should not fail if the namespace exists and has the expected labels", func() {
			ns := &corev1.Namespace{}
			ns.SetName("testns")
			ns.SetLabels(testLabelsMap)
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).WithInitObjects(ns).Build()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(ns), ns)).To(Succeed())
			Expect(ns.Labels).To(BeEquivalentTo(testLabelsMap))
			ns, err := clusteraccess.EnsureNamespace(env.Ctx, env.Client(), ns.Name, testLabelsList...)
			Expect(err).ToNot(HaveOccurred())
			Expect(ns.Labels).To(BeEquivalentTo(testLabelsMap))
		})

	})

	Context("EnsureServiceAccount", func() {

		It("should create a serviceaccount if it does not exist", func() {
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).Build()
			sa := &corev1.ServiceAccount{}
			sa.SetName("testsa")
			sa.SetNamespace("testns")
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(sa), sa)).To(MatchError(apierrors.IsNotFound, "sa should not exist"))
			sa, err := clusteraccess.EnsureServiceAccount(env.Ctx, env.Client(), sa.Name, sa.Namespace, testLabelsList...)
			Expect(err).ToNot(HaveOccurred())
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(sa), sa)).To(Succeed())
			Expect(sa.Labels).To(BeEquivalentTo(testLabelsMap))
		})

		It("should throw an error if the serviceaccount exists, but is missing the expected labels", func() {
			sa := &corev1.ServiceAccount{}
			sa.SetName("testsa")
			sa.SetNamespace("testns")
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).WithInitObjects(sa).Build()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(sa), sa)).To(Succeed())
			Expect(sa.Labels).To(BeEmpty())
			_, err := clusteraccess.EnsureServiceAccount(env.Ctx, env.Client(), sa.Name, sa.Namespace, testLabelsList...)
			Expect(err).To(HaveOccurred())
		})

		It("should not fail if the serviceaccount exists and has the expected labels", func() {
			sa := &corev1.ServiceAccount{}
			sa.SetName("testsa")
			sa.SetNamespace("testns")
			sa.SetLabels(testLabelsMap)
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).WithInitObjects(sa).Build()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(sa), sa)).To(Succeed())
			Expect(sa.Labels).To(BeEquivalentTo(testLabelsMap))
			sa, err := clusteraccess.EnsureServiceAccount(env.Ctx, env.Client(), sa.Name, sa.Namespace, testLabelsList...)
			Expect(err).ToNot(HaveOccurred())
			Expect(sa.Labels).To(BeEquivalentTo(testLabelsMap))
		})

	})

	Context("EnsureRole and EnsureClusterRole", func() {

		expectedRules := func() []rbacv1.PolicyRule {
			return []rbacv1.PolicyRule{
				{
					Verbs:     []string{"get", "list"},
					APIGroups: []string{"apps"},
					Resources: []string{"deployments"},
				},
				{
					Verbs:     []string{"*"},
					APIGroups: []string{""},
					Resources: []string{"namespaces"},
				},
			}
		}

		It("should create a role if it does not exist", func() {
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).Build()
			r := &rbacv1.Role{}
			r.SetName("testr")
			r.SetNamespace("testns")
			r.Rules = expectedRules()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(r), r)).To(MatchError(apierrors.IsNotFound, "role should not exist"))
			r, err := clusteraccess.EnsureRole(env.Ctx, env.Client(), r.Name, r.Namespace, expectedRules(), testLabelsList...)
			Expect(err).ToNot(HaveOccurred())
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(r), r)).To(Succeed())
			Expect(r.Labels).To(BeEquivalentTo(testLabelsMap))
			Expect(r.Rules).To(BeEquivalentTo(expectedRules()))
		})

		It("should throw an error if the role exists, but is missing the expected labels", func() {
			r := &rbacv1.Role{}
			r.SetName("testr")
			r.SetNamespace("testns")
			r.Rules = expectedRules()
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).WithInitObjects(r).Build()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(r), r)).To(Succeed())
			Expect(r.Labels).To(BeEmpty())
			_, err := clusteraccess.EnsureRole(env.Ctx, env.Client(), r.Name, r.Namespace, expectedRules(), testLabelsList...)
			Expect(err).To(HaveOccurred())
		})

		It("should update the rules if the role exists and has the expected labels", func() {
			r := &rbacv1.Role{}
			r.SetName("testr")
			r.SetNamespace("testns")
			r.SetLabels(testLabelsMap)
			r.Rules = []rbacv1.PolicyRule{
				{
					Verbs:     []string{"*"},
					APIGroups: []string{"*"},
					Resources: []string{"*"},
				},
			}
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).WithInitObjects(r).Build()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(r), r)).To(Succeed())
			Expect(r.Labels).To(BeEquivalentTo(testLabelsMap))
			r, err := clusteraccess.EnsureRole(env.Ctx, env.Client(), r.Name, r.Namespace, expectedRules(), testLabelsList...)
			Expect(err).ToNot(HaveOccurred())
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(r), r)).To(Succeed())
			Expect(r.Labels).To(BeEquivalentTo(testLabelsMap))
			Expect(r.Rules).To(BeEquivalentTo(expectedRules()))
		})

		It("should create a clusterrole if it does not exist", func() {
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).Build()
			cr := &rbacv1.ClusterRole{}
			cr.SetName("testcr")
			cr.Rules = expectedRules()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(cr), cr)).To(MatchError(apierrors.IsNotFound, "clusterrole should not exist"))
			cr, err := clusteraccess.EnsureClusterRole(env.Ctx, env.Client(), cr.Name, expectedRules(), testLabelsList...)
			Expect(err).ToNot(HaveOccurred())
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(cr), cr)).To(Succeed())
			Expect(cr.Labels).To(BeEquivalentTo(testLabelsMap))
			Expect(cr.Rules).To(BeEquivalentTo(expectedRules()))
		})

		It("should throw an error if the clusterrole exists, but is missing the expected labels", func() {
			cr := &rbacv1.ClusterRole{}
			cr.SetName("testcr")
			cr.Rules = expectedRules()
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).WithInitObjects(cr).Build()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(cr), cr)).To(Succeed())
			Expect(cr.Labels).To(BeEmpty())
			_, err := clusteraccess.EnsureClusterRole(env.Ctx, env.Client(), cr.Name, expectedRules(), testLabelsList...)
			Expect(err).To(HaveOccurred())
		})

		It("should update the rules if the clusterrole exists and has the expected labels", func() {
			cr := &rbacv1.ClusterRole{}
			cr.SetName("testcr")
			cr.SetLabels(testLabelsMap)
			cr.Rules = []rbacv1.PolicyRule{
				{
					Verbs:     []string{"*"},
					APIGroups: []string{"*"},
					Resources: []string{"*"},
				},
			}
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).WithInitObjects(cr).Build()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(cr), cr)).To(Succeed())
			Expect(cr.Labels).To(BeEquivalentTo(testLabelsMap))
			cr, err := clusteraccess.EnsureClusterRole(env.Ctx, env.Client(), cr.Name, expectedRules(), testLabelsList...)
			Expect(err).ToNot(HaveOccurred())
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(cr), cr)).To(Succeed())
			Expect(cr.Labels).To(BeEquivalentTo(testLabelsMap))
			Expect(cr.Rules).To(BeEquivalentTo(expectedRules()))
		})

	})

	Context("EnsureRoleBinding and EnsureClusterRoleBinding", func() {

		dummyRole := func() *rbacv1.Role {
			return &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testrole",
					Namespace: "testns",
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"*"},
						APIGroups: []string{"*"},
						Resources: []string{"*"},
					},
				},
			}
		}
		dummyClusterRole := func() *rbacv1.ClusterRole {
			return &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testclusterrole",
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"*"},
						APIGroups: []string{"*"},
						Resources: []string{"*"},
					},
				},
			}
		}
		dummyServiceAccount := func() *corev1.ServiceAccount {
			return &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testsa",
					Namespace: "testns",
				},
			}
		}
		expectedRoleRef := func() rbacv1.RoleRef {
			return rbacv1.RoleRef{
				APIGroup: rbacv1.SchemeGroupVersion.Group,
				Name:     "testrole",
				Kind:     "Role",
			}
		}
		expectedClusterRoleRef := func() rbacv1.RoleRef {
			return rbacv1.RoleRef{
				APIGroup: rbacv1.SchemeGroupVersion.Group,
				Name:     "testclusterrole",
				Kind:     "ClusterRole",
			}
		}
		expectedSubjects := func() []rbacv1.Subject {
			return []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      "testsa",
					Namespace: "testns",
				},
			}
		}

		It("should create a rolebinding if it does not exist", func() {
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).WithInitObjects(dummyRole(), dummyServiceAccount()).Build()
			rb := &rbacv1.RoleBinding{}
			rb.SetName("testrb")
			rb.SetNamespace("testns")
			rb.RoleRef = expectedRoleRef()
			rb.Subjects = expectedSubjects()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(rb), rb)).To(MatchError(apierrors.IsNotFound, "rolebinding should not exist"))
			rb, err := clusteraccess.EnsureRoleBinding(env.Ctx, env.Client(), rb.Name, rb.Namespace, rb.RoleRef.Name, rb.Subjects, testLabelsList...)
			Expect(err).ToNot(HaveOccurred())
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(rb), rb)).To(Succeed())
			Expect(rb.Labels).To(BeEquivalentTo(testLabelsMap))
			Expect(rb.RoleRef).To(BeEquivalentTo(expectedRoleRef()))
			Expect(rb.Subjects).To(BeEquivalentTo(expectedSubjects()))
		})

		It("should throw an error if the rolebinding exists, but is missing the expected labels", func() {
			rb := &rbacv1.RoleBinding{}
			rb.SetName("testrb")
			rb.SetNamespace("testns")
			rb.RoleRef = expectedRoleRef()
			rb.Subjects = expectedSubjects()
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).WithInitObjects(rb, dummyRole(), dummyServiceAccount()).Build()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(rb), rb)).To(Succeed())
			Expect(rb.Labels).To(BeEmpty())
			_, err := clusteraccess.EnsureRoleBinding(env.Ctx, env.Client(), rb.Name, rb.Namespace, rb.RoleRef.Name, rb.Subjects, testLabelsList...)
			Expect(err).To(HaveOccurred())
		})

		It("should update the rolebinding if it exists and has the expected labels", func() {
			rb := &rbacv1.RoleBinding{}
			rb.SetName("testrb")
			rb.SetNamespace("testns")
			rb.SetLabels(testLabelsMap)
			rb.RoleRef = rbacv1.RoleRef{
				APIGroup: rbacv1.SchemeGroupVersion.Group,
				Name:     "foo",
				Kind:     "Role",
			}
			rb.Subjects = []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      "bar",
					Namespace: "testns",
				},
			}
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).WithInitObjects(rb, dummyRole(), dummyServiceAccount()).Build()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(rb), rb)).To(Succeed())
			Expect(rb.Labels).To(BeEquivalentTo(testLabelsMap))
			rb, err := clusteraccess.EnsureRoleBinding(env.Ctx, env.Client(), rb.Name, rb.Namespace, expectedRoleRef().Name, expectedSubjects(), testLabelsList...)
			Expect(err).ToNot(HaveOccurred())
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(rb), rb)).To(Succeed())
			Expect(rb.Labels).To(BeEquivalentTo(testLabelsMap))
			Expect(rb.RoleRef).To(BeEquivalentTo(expectedRoleRef()))
			Expect(rb.Subjects).To(BeEquivalentTo(expectedSubjects()))
		})

		It("should create a clusterrolebinding if it does not exist", func() {
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).WithInitObjects(dummyClusterRole(), dummyServiceAccount()).Build()
			crb := &rbacv1.ClusterRoleBinding{}
			crb.SetName("testcrb")
			crb.RoleRef = expectedClusterRoleRef()
			crb.Subjects = expectedSubjects()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(crb), crb)).To(MatchError(apierrors.IsNotFound, "clusterrolebinding should not exist"))
			crb, err := clusteraccess.EnsureClusterRoleBinding(env.Ctx, env.Client(), crb.Name, crb.RoleRef.Name, crb.Subjects, testLabelsList...)
			Expect(err).ToNot(HaveOccurred())
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(crb), crb)).To(Succeed())
			Expect(crb.Labels).To(BeEquivalentTo(testLabelsMap))
			Expect(crb.RoleRef).To(BeEquivalentTo(expectedClusterRoleRef()))
			Expect(crb.Subjects).To(BeEquivalentTo(expectedSubjects()))
		})

		It("should throw an error if the clusterrolebinding exists, but is missing the expected labels", func() {
			crb := &rbacv1.ClusterRoleBinding{}
			crb.SetName("testcrb")
			crb.RoleRef = expectedClusterRoleRef()
			crb.Subjects = expectedSubjects()
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).WithInitObjects(crb, dummyClusterRole(), dummyServiceAccount()).Build()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(crb), crb)).To(Succeed())
			Expect(crb.Labels).To(BeEmpty())
			_, err := clusteraccess.EnsureClusterRoleBinding(env.Ctx, env.Client(), crb.Name, crb.RoleRef.Name, crb.Subjects, testLabelsList...)
			Expect(err).To(HaveOccurred())
		})

		It("should update the clusterrolebinding if it exists and has the expected labels", func() {
			crb := &rbacv1.ClusterRoleBinding{}
			crb.SetName("testcrb")
			crb.SetLabels(testLabelsMap)
			crb.RoleRef = rbacv1.RoleRef{
				APIGroup: rbacv1.SchemeGroupVersion.Group,
				Name:     "foo",
				Kind:     "ClusterRole",
			}
			crb.Subjects = []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      "bar",
					Namespace: "testns",
				},
			}
			env := testutils.NewEnvironmentBuilder().WithFakeClient(nil).WithInitObjects(crb, dummyClusterRole(), dummyServiceAccount()).Build()
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(crb), crb)).To(Succeed())
			Expect(crb.Labels).To(BeEquivalentTo(testLabelsMap))
			crb, err := clusteraccess.EnsureClusterRoleBinding(env.Ctx, env.Client(), crb.Name, expectedClusterRoleRef().Name, expectedSubjects(), testLabelsList...)
			Expect(err).ToNot(HaveOccurred())
			Expect(env.Client().Get(env.Ctx, client.ObjectKeyFromObject(crb), crb)).To(Succeed())
			Expect(crb.Labels).To(BeEquivalentTo(testLabelsMap))
			Expect(crb.RoleRef).To(BeEquivalentTo(expectedClusterRoleRef()))
			Expect(crb.Subjects).To(BeEquivalentTo(expectedSubjects()))
		})

	})

	Context("Marshal RESTConfig", func() {
		readRESTConfigFromKubeconfig := func(kubeconfig string) *rest.Config {
			data, err := os.ReadFile(fmt.Sprint("./testdata/kubeconfig/", kubeconfig))
			Expect(err).ToNot(HaveOccurred(), "failed to read kubeconfig file")

			config, err := clientcmd.RESTConfigFromKubeConfig(data)
			Expect(err).ToNot(HaveOccurred(), "failed to parse kubeconfig file")
			return config
		}

		It("should create an OIDC config", func() {
			restConfig := readRESTConfigFromKubeconfig("kubeconfig-token.yaml")

			oidcConfigRaw, err := clusteraccess.WriteOIDCConfigFromRESTConfig(restConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(oidcConfigRaw).ToNot(BeEmpty())

			var oidcConfig map[string]string
			Expect(yaml.Unmarshal(oidcConfigRaw, &oidcConfig)).ToNot(HaveOccurred())
			Expect(oidcConfig).To(HaveKeyWithValue("host", "https://test-server"))
			Expect(oidcConfig)
		})

		It("should create a kubeconfig with token", func() {
			restConfig := readRESTConfigFromKubeconfig("kubeconfig-token.yaml")

			kubeconfigRaw, err := clusteraccess.WriteKubeconfigFromRESTConfig(restConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(kubeconfigRaw).ToNot(BeEmpty())

			config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigRaw)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.Host).To(Equal("https://test-server"))
			Expect(config.TLSClientConfig.CAData).ToNot(BeEmpty())
			Expect(config.BearerToken).To(Equal("dGVzdC10b2tlbg=="))
		})

		It("should create a kubeconfig with basic auth", func() {
			restConfig := readRESTConfigFromKubeconfig("kubeconfig-basicauth.yaml")

			kubeconfigRaw, err := clusteraccess.WriteKubeconfigFromRESTConfig(restConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(kubeconfigRaw).ToNot(BeEmpty())

			config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigRaw)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.Host).To(Equal("https://test-server"))
			Expect(config.TLSClientConfig.CAData).ToNot(BeEmpty())
			Expect(config.Username).To(Equal("foo"))
			Expect(config.Password).To(Equal("bar"))
		})

		It("should create a kubeconfig with client tls", func() {
			restConfig := readRESTConfigFromKubeconfig("kubeconfig-tls.yaml")

			kubeconfigRaw, err := clusteraccess.WriteKubeconfigFromRESTConfig(restConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(kubeconfigRaw).ToNot(BeEmpty())

			config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigRaw)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.Host).To(Equal("https://test-server"))
			Expect(config.TLSClientConfig.CAData).ToNot(BeEmpty())
			Expect(config.TLSClientConfig.CertData).ToNot(BeEmpty())
			Expect(config.TLSClientConfig.KeyData).ToNot(BeEmpty())
		})
	})

})
