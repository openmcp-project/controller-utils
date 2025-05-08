# Webhooks

The `pkg/init/webhooks` provides easy tools to deploy webhook configuration and certificates on a target cluster.

### Noteworthy Functions
- `GenerateCertificate` generates and deploy webhook certificates to the target cluster.
- `Install` deploys mutating/validating webhook configuration on a target cluster.
