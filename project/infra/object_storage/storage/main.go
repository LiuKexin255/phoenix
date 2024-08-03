package main

import (
	"context"
	"log"
	"net"
	"time"

	"phoenix/common/go/constdef"
	"phoenix/common/go/env"
	"phoenix/project/infra/object_storage/storage/proto"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
)

var (
	kasp = keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
		MaxConnectionAge:      60 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
		MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
		Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
		Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
	}
)

const (
	pkgName = "phoenix/project/infra/object_storage/storage"
)

func main() {
	resource, err := resource.New(context.Background(),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			attribute.String(constdef.OtlpServiceName, "object_storage.storage.storage"),
		),
	)
	if err != nil {
		log.Panicln(err)
	}

	tracerProvider := tracerProvider(resource)
	tracer := tracerProvider.Tracer(pkgName)

	metricProvider := metricProvider(resource)
	meter := metricProvider.Meter(pkgName)

	lis, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.KeepaliveParams(kasp),
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithTracerProvider(tracerProvider),
			otelgrpc.WithMeterProvider(metricProvider),
			otelgrpc.WithPropagators(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{})),
		)),
	)

	serverImpl, err := newServer(tracer, meter)
	if err != nil {
		log.Panicln(err)
	}
	proto.RegisterStoragerServer(s, serverImpl)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func tracerProvider(res *resource.Resource) *sdktrace.TracerProvider {
	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithCompressor(gzip.Name),
		otlptracegrpc.WithEndpoint(env.GetUptraceEndpoint()),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithHeaders(map[string]string{
			// Set the Uptrace DSN here or use UPTRACE_DSN env var.
			constdef.UptraceDSNHeader: env.GetUptraceDSN(),
		}),
	)
	if err != nil {
		log.Panicln(err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter,
		sdktrace.WithMaxQueueSize(10_000),
		sdktrace.WithMaxExportBatchSize(10_000))
	sdktrace.WithSampler(sdktrace.AlwaysSample())
	// Call shutdown to flush the buffers when program exits.
	defer bsp.Shutdown(context.Background())

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithIDGenerator(xray.NewIDGenerator()),
	)
	tracerProvider.RegisterSpanProcessor(bsp)

	return tracerProvider
}

func metricProvider(res *resource.Resource) *sdkmetric.MeterProvider {
	exporter, err := otlpmetricgrpc.New(context.Background(),
		otlpmetricgrpc.WithEndpoint(env.GetUptraceEndpoint()),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithHeaders(map[string]string{
			// Set the Uptrace DSN here or use UPTRACE_DSN env var.
			constdef.UptraceDSNHeader: env.GetUptraceDSN(),
		}),
		otlpmetricgrpc.WithCompressor(gzip.Name),
		otlpmetricgrpc.WithTemporalitySelector(func(kind sdkmetric.InstrumentKind) metricdata.Temporality {
			switch kind {
			case sdkmetric.InstrumentKindCounter,
				sdkmetric.InstrumentKindObservableCounter,
				sdkmetric.InstrumentKindHistogram:
				return metricdata.DeltaTemporality
			default:
				return metricdata.CumulativeTemporality
			}
		}),
	)

	log.Panicln(err)

	reader := sdkmetric.NewPeriodicReader(
		exporter,
		sdkmetric.WithInterval(30*time.Second),
	)

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)

}
