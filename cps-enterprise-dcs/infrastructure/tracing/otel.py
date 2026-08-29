"""
OpenTelemetry Distributed Tracing for CP'S Enterprise DCS
Provides tracing instrumentation for Python services.
"""

from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.semconv.resource import ResourceAttributes
import os


def init_tracer(service_name: str, service_version: str = "1.0.0") -> TracerProvider:
    """Initialize OpenTelemetry tracer for a Python service."""
    otlp_endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
    environment = os.getenv("DCS_ENV", "development")

    resource = Resource.create({
        ResourceAttributes.SERVICE_NAME: service_name,
        ResourceAttributes.SERVICE_VERSION: service_version,
        ResourceAttributes.DEPLOYMENT_ENVIRONMENT: environment,
    })

    exporter = OTLPSpanExporter(endpoint=otlp_endpoint, insecure=True)
    processor = BatchSpanProcessor(exporter)

    provider = TracerProvider(resource=resource)
    provider.add_span_processor(processor)

    trace.set_tracer_provider(provider)
    return provider


def get_tracer(module_name: str) -> trace.Tracer:
    """Get a tracer for a specific module."""
    return trace.get_tracer(module_name)


class TraceContext:
    """Context manager for creating spans."""

    def __init__(self, name: str, tracer: trace.Tracer = None):
        self.name = name
        self.tracer = tracer or get_tracer(__name__)
        self.span = None

    def __enter__(self):
        self.span = self.tracer.start_span(self.name)
        self.span.__enter__()
        return self.span

    def __exit__(self, exc_type, exc_val, exc_tb):
        if exc_type is not None:
            self.span.record_exception(exc_val)
            self.span.set_status(trace.Status(trace.StatusCode.ERROR, str(exc_val)))
        self.span.__exit__(exc_type, exc_val, exc_tb)
