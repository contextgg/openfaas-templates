package srv

import (
	"context"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"

	"github.com/contextcloud/graceful/config"
)

type tracer struct {
	tp *trace.TracerProvider
}

func (t *tracer) Start(ctx context.Context) error {
	return nil
}

func (t *tracer) Shutdown(ctx context.Context) error {
	t.tp.ForceFlush(ctx)
	return t.tp.Shutdown(ctx)
}

func newTracer(exporter trace.SpanExporter, res *resource.Resource) (Startable, error) {
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return &tracer{
		tp: tp,
	}, nil
}

func NewGcpTracer(projectId string, res *resource.Resource) (Startable, error) {
	exporter, err := texporter.New(texporter.WithProjectID(projectId))
	if err != nil {
		return nil, err
	}
	return newTracer(exporter, res)
}

func NewZipkin(url string, res *resource.Resource) (Startable, error) {
	exporter, err := zipkin.New(url)
	if err != nil {
		return nil, err
	}
	return newTracer(exporter, res)
}

func NewTracer(ctx context.Context, cfg *config.Config) (Startable, error) {
	if !cfg.Tracing.Enabled {
		return NewNoop(), nil
	}

	res, err := resource.New(ctx,
		// Use the GCP resource detector to detect information about the GCP platform
		resource.WithDetectors(gcp.NewDetector()),
		// Keep the default detectors
		resource.WithTelemetrySDK(),
		// Add your own custom attributes to identify your application
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.Version),
		),
	)
	if err != nil {
		return nil, err
	}

	switch cfg.Tracing.Type {
	case "zipkin":
		return NewZipkin(cfg.Tracing.Url, res)
	case "gcp":
		return NewGcpTracer(cfg.Tracing.Url, res)
	default:
		return NewNoop(), nil
	}
}
