package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"phoenix/common/go/constdef"
	"phoenix/common/go/env"
	"phoenix/project/infra/object_storage/storage/proto"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
)

var (
	kacp = keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}
)

func main() {
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
		log.Panic(err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter,
		sdktrace.WithMaxQueueSize(10_000),
		sdktrace.WithMaxExportBatchSize(10_000))
	sdktrace.WithSampler(sdktrace.AlwaysSample())
	// Call shutdown to flush the buffers when program exits.
	defer bsp.Shutdown(context.Background())

	resource, err := resource.New(context.Background(),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			attribute.String(constdef.OtlpServiceName, "object_storage.gateway.gateway"),
		),
	)
	if err != nil {
		log.Panic(err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithIDGenerator(xray.NewIDGenerator()),
	)
	tracerProvider.RegisterSpanProcessor(bsp)

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.NewClient("object-storage--storage-service:8888",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(
			otelgrpc.WithTracerProvider(tracerProvider),
			otelgrpc.WithPropagators(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{})),
		)),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	err = proto.RegisterStoragerHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr: ":443",
		Handler: otelhttp.NewHandler(gwmux, "grpc-gateway",
			otelhttp.WithTracerProvider(tracerProvider),
			otelhttp.WithPropagators(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{})),
		),
	}
	log.Fatalln(gwServer.ListenAndServe())
}
