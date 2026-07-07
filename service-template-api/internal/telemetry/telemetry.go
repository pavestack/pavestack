// Package telemetry wires OpenTelemetry traces and metrics end-to-end.
// Both pipelines share one otlpEndpoint and one Resource (service.name),
// and otelhttp.NewHandler (see internal/server) uses the global
// TracerProvider and MeterProvider this package installs to produce, for
// every request, a trace span AND the semantic-convention
// "http.server.request.duration" histogram - which is why
// platform-config/observability/slo-burn-rate-alerts.yaml and
// service-template-api's Argo Rollouts AnalysisTemplate can both query
// `http_server_request_duration_seconds_count` without this service
// hand-rolling a metrics pipeline. Logs are correlated to the active
// trace/span via internal/logging.TraceContext - see that package for the
// third pillar.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type ShutdownFunc func(context.Context) error

func Init(ctx context.Context, serviceName, otlpEndpoint string) (ShutdownFunc, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	tracerProvider, err := newTracerProvider(ctx, res, otlpEndpoint)
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	meterProvider, err := newMeterProvider(ctx, res, otlpEndpoint)
	if err != nil {
		return nil, err
	}
	otel.SetMeterProvider(meterProvider)

	return func(shutdownCtx context.Context) error {
		traceErr := tracerProvider.Shutdown(shutdownCtx)
		metricErr := meterProvider.Shutdown(shutdownCtx)
		if traceErr != nil {
			return traceErr
		}
		return metricErr
	}, nil
}

func newTracerProvider(ctx context.Context, res *resource.Resource, otlpEndpoint string) (*sdktrace.TracerProvider, error) {
	if otlpEndpoint == "" {
		return sdktrace.NewTracerProvider(sdktrace.WithResource(res)), nil
	}

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(otlpEndpoint))
	if err != nil {
		return nil, fmt.Errorf("create otlp trace exporter: %w", err)
	}
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	), nil
}

func newMeterProvider(ctx context.Context, res *resource.Resource, otlpEndpoint string) (*metric.MeterProvider, error) {
	if otlpEndpoint == "" {
		return metric.NewMeterProvider(metric.WithResource(res)), nil
	}

	exporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(otlpEndpoint))
	if err != nil {
		return nil, fmt.Errorf("create otlp metric exporter: %w", err)
	}
	return metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exporter)),
	), nil
}
