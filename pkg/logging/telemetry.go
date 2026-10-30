// SPDX-FileCopyrightText: Copyright OpenControlPlane contributors.
//
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	otlploggrpc "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

const instrumentationScope = "github.com/openmcp-project/controller-utils/pkg/logging"

// ShutdownFunc flushes and shuts down the log provider.
type ShutdownFunc func(context.Context) error

var noopShutdown ShutdownFunc = func(context.Context) error { return nil }

// Resource returns the SDK's process-wide default resource.
func Resource() *resource.Resource {
	return resource.Default()
}

// RegisterSDKInfoMetric exposes OTel SDK identity to Prometheus receivers via
// the standard target_info series.
func RegisterSDKInfoMetric(reg prometheus.Registerer) error {
	info := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "target_info",
		Help: "Telemetry SDK metadata for this scrape target.",
		ConstLabels: prometheus.Labels{
			"telemetry_sdk_name":     "opentelemetry",
			"telemetry_sdk_language": "go",
			"telemetry_sdk_version":  sdk.Version(),
		},
	})
	info.Set(1)
	var alreadyRegistered prometheus.AlreadyRegisteredError
	if err := reg.Register(info); err != nil && !errors.As(err, &alreadyRegistered) {
		return err
	}
	return nil
}

// SetupLogs enables OTLP/gRPC log export when standard OTel environment
// variables configure an exporter endpoint.
func SetupLogs(ctx context.Context) (ShutdownFunc, error) {
	if !logsExportConfigured() {
		return noopShutdown, nil
	}

	exporter, err := otlploggrpc.New(ctx)
	if err != nil {
		return noopShutdown, err
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(Resource()),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)
	otel.SetLoggerProvider(provider)

	return provider.Shutdown, nil
}

func logsExportConfigured() bool {
	exporter := strings.ToLower(strings.TrimSpace(os.Getenv("OTEL_LOGS_EXPORTER")))
	if exporter == "none" {
		return false
	}
	return exporter == "otlp" || os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT") != "" || os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != ""
}
