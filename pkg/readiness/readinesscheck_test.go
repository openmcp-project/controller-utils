package readiness_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/openmcp-project/controller-utils/pkg/readiness"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Readiness Check Test Suite")
}

var _ = Describe("Readiness Check", func() {

	It("should return true when the readiness check is ready", func() {
		result := readiness.NewReadyResult()
		Expect(result.IsReady()).To(BeTrue())
	})

	It("should return false when the readiness check is not ready", func() {
		result := readiness.NewNotReadyResult("test message")
		Expect(result.IsReady()).To(BeFalse())
	})

	It("should return false when the readiness check is failed", func() {
		result := readiness.NewFailedResult(nil)
		Expect(result.IsReady()).To(BeFalse())
	})

	It("should return the message", func() {
		result := readiness.NewNotReadyResult("test message")
		Expect(result.Message()).To(Equal("test message"))
	})

	It("should return the message with multiple messages", func() {
		result := readiness.Aggregate(
			readiness.NewNotReadyResult("test message 1"),
			readiness.NewNotReadyResult("test message 2"),
		)
		Expect(result.Message()).To(Equal("test message 1, test message 2"))
	})

	It("should return true when a deployment is ready", func() {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 1,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](1),
			},
			Status: appsv1.DeploymentStatus{
				ObservedGeneration: 1,
				Replicas:           1,
				UpdatedReplicas:    1,
				AvailableReplicas:  1,
			},
		}
		result := readiness.CheckDeployment(deployment)
		Expect(result.IsReady()).To(BeTrue())
	})

	It("should return false when a deployment is not ready", func() {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 1,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](1),
			},
			Status: appsv1.DeploymentStatus{
				ObservedGeneration: 1,
				Replicas:           1,
				UpdatedReplicas:    0,
				AvailableReplicas:  0,
			},
		}
		result := readiness.CheckDeployment(deployment)
		Expect(result.IsReady()).To(BeFalse())
	})

	It("should return true when a job is completed", func() {
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "test-job",
			},
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{
					{
						Type:   batchv1.JobComplete,
						Status: "True",
					},
				},
			},
		}
		result := readiness.CheckJob(job)
		Expect(result.IsReady()).To(BeTrue())
	})

	It("should return false when a job has failed", func() {
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "test-job",
			},
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{
					{
						Type:    batchv1.JobFailed,
						Status:  "True",
						Message: "Job failed due to an error",
					},
				},
			},
		}
		result := readiness.CheckJob(job)
		Expect(result.IsReady()).To(BeFalse())
		Expect(result.Message()).To(Equal("job default/test-job failed: Job failed due to an error"))
	})

	It("should return false when a job is not completed", func() {
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "test-job",
			},
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{},
			},
		}
		result := readiness.CheckJob(job)
		Expect(result.IsReady()).To(BeFalse())
		Expect(result.Message()).To(Equal("job default/test-job is not completed"))
	})

	It("should return true when a job has failed using CheckJobFailed", func() {
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "test-job",
			},
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{
					{
						Type:   batchv1.JobFailed,
						Status: "True",
					},
				},
			},
		}
		Expect(readiness.CheckJobFailed(job)).To(BeTrue())
	})

	It("should return false when a job has not failed using CheckJobFailed", func() {
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "test-job",
			},
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{},
			},
		}
		Expect(readiness.CheckJobFailed(job)).To(BeFalse())
	})

	It("should return true when an ingress is ready", func() {
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "test-ingress",
			},
			Status: networkingv1.IngressStatus{
				LoadBalancer: networkingv1.IngressLoadBalancerStatus{
					Ingress: []networkingv1.IngressLoadBalancerIngress{
						{Hostname: "example.com"},
					},
				},
			},
		}
		result := readiness.CheckIngress(ingress)
		Expect(result.IsReady()).To(BeTrue())
	})

	It("should return false when an ingress is not ready", func() {
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "test-ingress",
			},
			Status: networkingv1.IngressStatus{
				LoadBalancer: networkingv1.IngressLoadBalancerStatus{
					Ingress: []networkingv1.IngressLoadBalancerIngress{},
				},
			},
		}
		result := readiness.CheckIngress(ingress)
		Expect(result.IsReady()).To(BeFalse())
		Expect(result.Message()).To(Equal("ingress default/test-ingress is not ready: no load balancer ingress entries"))
	})

	It("should return true when a gateway is ready", func() {
		gateway := &gatewayv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "test-gateway",
			},
			Status: gatewayv1.GatewayStatus{
				Conditions: []metav1.Condition{
					{
						Type:   string(gatewayv1.GatewayConditionReady),
						Status: "True",
					},
				},
			},
		}
		result := readiness.CheckGateway(gateway)
		Expect(result.IsReady()).To(BeTrue())
	})

	It("should return false when a gateway is not ready", func() {
		gateway := &gatewayv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "test-gateway",
			},
			Status: gatewayv1.GatewayStatus{
				Conditions: []metav1.Condition{},
			},
		}
		result := readiness.CheckGateway(gateway)
		Expect(result.IsReady()).To(BeFalse())
		Expect(result.Message()).To(Equal("gateway default/test-gateway is not ready: no ready condition with status True"))
	})
})
