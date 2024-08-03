package otel

import (
	"context"
	"log"
	"time"

	"phoenix/common/go/constdef"
	"phoenix/common/go/env"

	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/encoding/gzip"
)

type Shutdown func(ctx context.Context) error

// MustNewTracerProvider 创建 OTLP tracer provider
func MustNewTracerProvider(ctx context.Context, res *resource.Resource) (*sdktrace.TracerProvider, Shutdown) {
	exporter, err := otlptracegrpc.New(
		ctx,
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
		sdktrace.WithMaxExportBatchSize(10_000),
	)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithIDGenerator(xray.NewIDGenerator()),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	tracerProvider.RegisterSpanProcessor(bsp)

	return tracerProvider, func(ctx context.Context) error {
		return tracerProvider.Shutdown(ctx)
	}
}

// MustNewMetricProvider 创建 OTLP metices provider
func MustNewMetricProvider(ctx context.Context, res *resource.Resource) (*sdkmetric.MeterProvider, func(ctx context.Context) error) {
	exporter, err := otlpmetricgrpc.New(ctx,
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
	if err != nil {
		log.Panicln(err)
	}

	reader := sdkmetric.NewPeriodicReader(
		exporter,
		sdkmetric.WithInterval(30*time.Second),
	)

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)

	return provider, func(ctx context.Context) error {
		return provider.Shutdown(ctx)
	}
}

// MustNewLoggerProvider 创建 OTLP logger provider
func MustNewLoggerProvider(ctx context.Context, res *resource.Resource) (*sdklog.LoggerProvider, func(ctx context.Context) error) {
	exporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithEndpoint(env.GetUptraceEndpoint()),
		otlploggrpc.WithHeaders(map[string]string{
			constdef.UptraceDSNHeader: env.GetUptraceDSN(),
		}),
		otlploggrpc.WithCompressor(gzip.Name),
	)
	if err != nil {
		log.Panicln(err)
	}

	bsp := sdklog.NewBatchProcessor(exporter,
		sdklog.WithMaxQueueSize(10_000),
		sdklog.WithExportMaxBatchSize(10_000),
		sdklog.WithExportInterval(10*time.Second),
		sdklog.WithExportTimeout(10*time.Second),
	)

	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(bsp),
		sdklog.WithResource(res),
	)

	return provider, func(ctx context.Context) error {
		return provider.Shutdown(ctx)
	}
}
