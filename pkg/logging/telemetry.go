// SPDX-FileCopyrightText: Copyright OpenControlPlane contributors.
//
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	otlploggrpc "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
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
