package tracing

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/kafka-go"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc/metadata"
)

// StartHttpServerTracerSpan starts a new span for HTTP server requests using Fiber
func StartHttpServerTracerSpan(c *fiber.Ctx, operationName string) (context.Context, trace.Span) {
	// Extract trace context from HTTP headers
	headerCarrier := make(propagation.HeaderCarrier)
	c.Request().Header.VisitAll(func(key, value []byte) {
		headerCarrier.Set(string(key), string(value))
	})

	ctx := otel.GetTextMapPropagator().Extract(c.Context(), headerCarrier)

	// Start a new span
	tracer := otel.Tracer("")
	ctx, span := tracer.Start(ctx, operationName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			semconv.HTTPMethod(c.Method()),
			semconv.HTTPURL(c.OriginalURL()),
			semconv.HTTPRoute(c.Route().Path),
		),
	)

	return ctx, span
}

// GetTextMapCarrierFromMetaData extracts trace context from gRPC metadata
func GetTextMapCarrierFromMetaData(ctx context.Context) propagation.MapCarrier {
	carrier := make(propagation.MapCarrier)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for key, values := range md {
			if len(values) > 0 {
				carrier.Set(key, values[0])
			}
		}
	}
	return carrier
}

// StartGrpcServerTracerSpan starts a new span for gRPC server requests
func StartGrpcServerTracerSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	// Extract trace context from gRPC metadata
	carrier := GetTextMapCarrierFromMetaData(ctx)
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	// Start a new span
	tracer := otel.Tracer("")
	ctx, span := tracer.Start(ctx, operationName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			semconv.RPCSystemGRPC,
		),
	)

	return ctx, span
}

// StartKafkaConsumerTracerSpan starts a new span for Kafka consumer operations
func StartKafkaConsumerTracerSpan(ctx context.Context, headers []kafka.Header, operationName string) (context.Context, trace.Span) {
	// Extract trace context from Kafka headers
	carrier := TextMapCarrierFromKafkaMessageHeaders(headers)
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	// Start a new span
	tracer := otel.Tracer("")
	ctx, span := tracer.Start(ctx, operationName,
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
		),
	)

	return ctx, span
}

// TextMapCarrierToKafkaMessageHeaders converts OpenTelemetry carrier to Kafka headers
func TextMapCarrierToKafkaMessageHeaders(carrier propagation.MapCarrier) []kafka.Header {
	headers := make([]kafka.Header, 0, len(carrier))

	for key, value := range carrier {
		headers = append(headers, kafka.Header{
			Key:   key,
			Value: []byte(value),
		})
	}

	return headers
}

// TextMapCarrierFromKafkaMessageHeaders converts Kafka headers to OpenTelemetry carrier
func TextMapCarrierFromKafkaMessageHeaders(headers []kafka.Header) propagation.MapCarrier {
	carrier := make(propagation.MapCarrier, len(headers))
	for _, header := range headers {
		carrier.Set(header.Key, string(header.Value))
	}
	return carrier
}

// InjectTextMapCarrier creates a carrier with injected trace context
func InjectTextMapCarrier(ctx context.Context) propagation.MapCarrier {
	carrier := make(propagation.MapCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier
}

// InjectTextMapCarrierToGrpcMetaData injects trace context into gRPC metadata
func InjectTextMapCarrierToGrpcMetaData(ctx context.Context) context.Context {
	carrier := InjectTextMapCarrier(ctx)
	md := metadata.New(map[string]string(carrier))
	return metadata.NewOutgoingContext(ctx, md)
}

// GetKafkaTracingHeadersFromSpanCtx gets Kafka headers with trace context
func GetKafkaTracingHeadersFromSpanCtx(ctx context.Context) []kafka.Header {
	carrier := InjectTextMapCarrier(ctx)
	return TextMapCarrierToKafkaMessageHeaders(carrier)
}

// StartClientSpan starts a new span for client operations (HTTP, gRPC, Kafka producer)
func StartClientSpan(ctx context.Context, operationName string, spanKind trace.SpanKind) (context.Context, trace.Span) {
	tracer := otel.Tracer("")
	return tracer.Start(ctx, operationName, trace.WithSpanKind(spanKind))
}

// StartHttpClientSpan starts a new span for HTTP client requests
func StartHttpClientSpan(ctx context.Context, method, url string) (context.Context, trace.Span) {
	tracer := otel.Tracer("")
	return tracer.Start(ctx, fmt.Sprintf("HTTP %s", method),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.HTTPMethod(method),
			semconv.HTTPURL(url),
		),
	)
}

// StartKafkaProducerSpan starts a new span for Kafka producer operations
func StartKafkaProducerSpan(ctx context.Context, topic string) (context.Context, trace.Span) {
	tracer := otel.Tracer("")
	return tracer.Start(ctx, fmt.Sprintf("kafka.produce %s", topic),
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination", topic),
		),
	)
}
