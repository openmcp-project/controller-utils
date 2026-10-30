# Logging

This package contains the logging library from the [Landscaper controller-utils module](https://github.com/gardener/landscaper/tree/master/controller-utils/pkg/logging).

The library provides a wrapper around `logr.Logger`, exposing additional helper functions. The original `logr.Logger` can be retrieved by using the `Logr()` method. Also, it notices when multiple values are added to the logger with the same key - instead of simply overwriting the previous ones (like `logr.Logger` does it), it appends the key with a `_conflict(x)` suffix, where `x` is the number of times this conflict has occurred.

### Noteworthy Functions

- `GetLogger()` is a singleton-style getter function for a logger.
- There are several `FromContext...` functions for retrieving a logger from a `context.Context` object.
- `InitFlags(...)` can be used to add the configuration flags for this logger to a cobra `FlagSet`.

### OpenTelemetry logs

Call `logging.SetupLogs(ctx)` once at application startup to enable OTLP/gRPC export. Loggers created by `New`, `GetLogger`, and `NewCliLogger` additionally export to the global OpenTelemetry logger provider when export is configured. Existing stderr output, formatting, verbosity, and key-conflict handling remain unchanged.

```go
shutdown, err := logging.SetupLogs(ctx)
if err != nil {
    // Report the setup error through the application logger; stderr remains available.
    log.Error(err, "unable to configure OpenTelemetry logs")
}
defer func() {
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := shutdown(shutdownCtx); err != nil {
        log.Error(err, "unable to flush OpenTelemetry logs")
    }
}()
```

Set `OTEL_EXPORTER_OTLP_ENDPOINT` or `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT` (for example, `http://otel-collector.monitoring.svc:4317`), or explicitly select `OTEL_LOGS_EXPORTER=otlp`. Without these, export stays disabled. `OTEL_LOGS_EXPORTER=none` overrides configured endpoints. Transport is gRPC; the exporter handles standard OTLP endpoint, TLS, header, and timeout variables, with logs-specific settings taking precedence.

`SetupLogs` installs a process-wide provider; the application owns its lifecycle. Configure the environment before creating loggers, call setup only once, and flush with a fresh, bounded context after stopping application work. The returned shutdown function is safe to call even when setup fails or export is disabled. Exported logs use `Resource()`, the SDK default resource, including `service.name` and `telemetry.sdk.*`; configure service identity with `OTEL_SERVICE_NAME` and `OTEL_RESOURCE_ATTRIBUTES` before setup. Existing custom global providers are replaced only when export is enabled.

Applications exposing Prometheus metrics can explicitly call `logging.RegisterSDKInfoMetric(registry)` to publish SDK identity on `target_info`. No metric is registered automatically. Collectors can promote the `telemetry_sdk_name`, `telemetry_sdk_language`, and `telemetry_sdk_version` labels to resource attributes. Repeated registration is supported; conflicting descriptors return an error.

When exporting both stderr-collected logs and OTLP logs, configure collector-side deduplication. Collector deployment and operator chart configuration remain application responsibilities.
