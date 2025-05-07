package readiness

import (
	"fmt"
	"slices"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	networkingv1 "k8s.io/api/networking/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type CheckResult []string

func (r CheckResult) IsReady() bool {
	return len(r) == 0
}

func (r CheckResult) Message() string {
	return strings.Join(r, ", ")
}

func NewReadyResult() CheckResult {
	return CheckResult{}
}

func NewNotReadyResult(message string) CheckResult {
	return CheckResult{message}
}

func NewFailedResult(err error) CheckResult {
	return NewNotReadyResult(fmt.Sprintf("readiness check failed: %v", err))
}

func Aggregate(results ...CheckResult) CheckResult {
	return slices.Concat(results...)
}

// CheckDeployment checks the readiness of a deployment.
func CheckDeployment(dp *appsv1.Deployment) CheckResult {
	if dp.Status.ObservedGeneration < dp.Generation {
		return NewNotReadyResult(fmt.Sprintf("deployment %s/%s not ready: observed generation outdated", dp.Namespace, dp.Name))
	}

	var specReplicas int32 = 0
	if dp.Spec.Replicas != nil {
		specReplicas = *dp.Spec.Replicas
	}

	if dp.Generation != dp.Status.ObservedGeneration ||
		specReplicas != dp.Status.Replicas ||
		specReplicas != dp.Status.UpdatedReplicas ||
		specReplicas != dp.Status.AvailableReplicas {
		return NewNotReadyResult(fmt.Sprintf("deployment %s/%s is not ready", dp.Namespace, dp.Name))
	}

	return NewReadyResult()
}

// CheckJob checks the completion status of a Kubernetes Job.
func CheckJob(job *batchv1.Job) CheckResult {
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == "True" {
			return NewReadyResult()
		}
		if condition.Type == batchv1.JobFailed && condition.Status == "True" {
			return NewNotReadyResult(fmt.Sprintf("job %s/%s failed: %s", job.Namespace, job.Name, condition.Message))
		}
	}

	return NewNotReadyResult(fmt.Sprintf("job %s/%s is not completed", job.Namespace, job.Name))
}

// CheckJobFailed checks if a Kubernetes Job has failed.
func CheckJobFailed(job *batchv1.Job) bool {
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobFailed && condition.Status == "True" {
			return true
		}
	}
	return false
}

// CheckIngress checks the readiness of a Kubernetes Ingress.
func CheckIngress(ingress *networkingv1.Ingress) CheckResult {
	if len(ingress.Status.LoadBalancer.Ingress) > 0 {
		return NewReadyResult()
	}
	return NewNotReadyResult(fmt.Sprintf("ingress %s/%s is not ready: no load balancer ingress entries", ingress.Namespace, ingress.Name))
}

// CheckGateway checks the readiness of a Kubernetes Gateway.
func CheckGateway(gateway *gatewayv1.Gateway) CheckResult {
	for _, condition := range gateway.Status.Conditions {
		if condition.Type == string(gatewayv1.GatewayConditionReady) && condition.Status == "True" {
			return NewReadyResult()
		}
	}
	return NewNotReadyResult(fmt.Sprintf("gateway %s/%s is not ready: no ready condition with status True", gateway.Namespace, gateway.Name))
}
