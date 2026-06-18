package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("demo-service"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("demo-tracer")
	ctx, span := tracer.Start(context.Background(), "main-operation")
	
	log.Println("Doing work...")
	time.Sleep(500 * time.Millisecond)
	
	_, childSpan := tracer.Start(ctx, "child-operation")
	time.Sleep(200 * time.Millisecond)
	childSpan.End()

	span.End()
	log.Println("Work complete.")
}
