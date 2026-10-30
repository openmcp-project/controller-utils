// SPDX-FileCopyrightText: Copyright OpenControlPlane contributors.
//
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	collector "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/grpc"
)

// A real receiver exercises endpoint selection, batching, and shutdown flushing.
type logReceiver struct {
	collector.UnimplementedLogsServiceServer
	requests chan *collector.ExportLogsServiceRequest
}

func (r *logReceiver) Export(_ context.Context, request *collector.ExportLogsServiceRequest) (*collector.ExportLogsServiceResponse, error) {
	r.requests <- request
	return &collector.ExportLogsServiceResponse{}, nil
}

func TestOTLPLogExport(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	server := grpc.NewServer()
	receiver := &logReceiver{requests: make(chan *collector.ExportLogsServiceRequest, 10)}
	collector.RegisterLogsServiceServer(server, receiver)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	t.Setenv("OTEL_LOGS_EXPORTER", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:1")
	t.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "http://"+listener.Addr().String())
	previous := otel.GetLoggerProvider()
	t.Cleanup(func() { otel.SetLoggerProvider(previous) })
	shutdown, err := SetupLogs(context.Background())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	t.Cleanup(func() { _ = shutdown(ctx) })
	logger, err := New(&Config{})
	require.NoError(t, err)
	logger.Debug("filtered")
	logger.WithName("controller").WithValues("object", "sample").Info("exported")
	require.NoError(t, shutdown(ctx))
	select {
	case request := <-receiver.requests:
		require.Len(t, request.ResourceLogs, 1)
		resource := request.ResourceLogs[0]
		attrs := map[string]string{}
		for _, attr := range resource.Resource.Attributes {
			attrs[attr.Key] = attr.Value.GetStringValue()
		}
		require.Equal(t, "opentelemetry", attrs["telemetry.sdk.name"])
		require.Equal(t, "go", attrs["telemetry.sdk.language"])
		require.NotEmpty(t, attrs["telemetry.sdk.version"])
		require.NotEmpty(t, attrs["service.name"])
		require.Len(t, resource.ScopeLogs, 1)
		require.Len(t, resource.ScopeLogs[0].LogRecords, 1)
		record := resource.ScopeLogs[0].LogRecords[0]
		require.Equal(t, "exported", record.Body.GetStringValue())
		fields := map[string]string{}
		for _, attr := range record.Attributes {
			fields[attr.Key] = attr.Value.GetStringValue()
		}
		require.Equal(t, "sample", fields["object"])
	case <-ctx.Done():
		t.Fatal("OTLP receiver did not receive the flushed log")
	}
}

func TestLogsExportConfigured(t *testing.T) {
	for _, tc := range []struct {
		name, exporter, common, logs string
		enabled                      bool
	}{
		{name: "disabled"},
		{name: "common endpoint", common: "http://collector:4317", enabled: true},
		{name: "logs endpoint", logs: "http://collector:4317", enabled: true},
		{name: "explicit exporter", exporter: "otlp", enabled: true},
		{name: "none takes precedence", exporter: " NONE ", common: "http://collector:4317", logs: "http://collector:4317"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("OTEL_LOGS_EXPORTER", tc.exporter)
			t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", tc.common)
			t.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", tc.logs)
			require.Equal(t, tc.enabled, logsExportConfigured())
		})
	}
}
