package tracing

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig holds OpenTelemetry configuration
type TracingConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	SampleRate     float64
}

// InitTracer initializes OpenTelemetry tracing
func InitTracer(cfg *TracingConfig) (*sdktrace.TracerProvider, error) {
	if cfg == nil {
		cfg = &TracingConfig{
			ServiceName:    "dcs-service",
			ServiceVersion: "1.0.0",
			Environment:    os.Getenv("DCS_ENV"),
			OTLPEndpoint:   os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
			SampleRate:     1.0,
		}
	}

	ctx := context.Background()

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
		)),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRate)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTracerName(cfg.ServiceName)

	return tp, nil
}

// StartSpan starts a new span
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer("dcs").Start(ctx, name)
}

// AddEvent adds an event to the current span
func AddEvent(ctx context.Context, name string, attributes ...trace.EventOption) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, attributes...)
}

// SetAttributes sets attributes on the current span
func SetAttributes(ctx context.Context, attributes ...trace.Attribute) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attributes...)
}

// RecordError records an error on the current span
func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}
