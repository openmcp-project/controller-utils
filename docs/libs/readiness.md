# Readiness Checks

The `pkg/readiness` package provides a simple way to check if a kubernetes resource is ready.
The meaning of readiness depends on the resource type.

### Examples

```go
deployment := &appsv1.Deployment{}
err := r.Client.Get(ctx, types.NamespacedName{
  Name:      "my-deployment",
  Namespace: "my-namespace",
}, deployment)

if err != nil {
  return err
}

readiness := readiness.CheckDeployment(deployment)

if readiness.IsReady() {
  fmt.Println("Deployment is ready")
} else {
  fmt.Printf("Deployment is not ready: %s\n", readiness.Message())
}
```
