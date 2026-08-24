package tracing

import (
	"context"
	"errors"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const defaultServiceName = "skycoin4444"

func Tracer(serviceName string) trace.Tracer {
	if serviceName == "" {
		serviceName = os.Getenv("OTEL_SERVICE_NAME")
	}
	if serviceName == "" {
		serviceName = defaultServiceName
	}
	return otel.Tracer(serviceName)
}

func Start(ctx context.Context, tracer trace.Tracer, spanName string) (context.Context, trace.Span, error) {
	if tracer == nil {
		return ctx, trace.SpanFromContext(ctx), errors.New("tracer is required")
	}
	if spanName == "" {
		return ctx, trace.SpanFromContext(ctx), errors.New("span name is required")
	}
	spanCtx, span := tracer.Start(ctx, spanName)
	return spanCtx, span, nil
}

func HTTPHandler(handler http.Handler, operation string) http.Handler {
	if operation == "" {
		operation = "http.request"
	}
	return otelhttp.NewHandler(handler, operation)
}
