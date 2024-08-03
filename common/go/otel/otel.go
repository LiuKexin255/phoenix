package otel

import (
	"context"
	"log"

	"phoenix/common/go/constdef"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalTracer trace.Tracer
	globalMeter  metric.Meter
	globalLogger *otelzap.Logger

	globalTracerProvider *sdktrace.TracerProvider
	globalMetricProvider *sdkmetric.MeterProvider
	globalLoggerProvider *sdklog.LoggerProvider
	globalResource       *resource.Resource
)

const (
	pkgName = "phoenix/common/go/otel"
)

func Tracer() trace.Tracer { return globalTracer }

func Meter() metric.Meter { return globalMeter }

func Logger(ctx context.Context) otelzap.LoggerWithCtx { return globalLogger.Ctx(ctx) }

func TracerProvider() *sdktrace.TracerProvider { return globalTracerProvider }

func MetricProvider() *sdkmetric.MeterProvider { return globalMetricProvider }

func LoggerProvider() *sdklog.LoggerProvider { return globalLoggerProvider }

func MustInit(ctx context.Context, serviceName string) Shutdown {
	resource, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			attribute.String(constdef.OtlpServiceName, serviceName),
		),
	)

	if err != nil {
		log.Panicln(err)
	}
	globalResource = resource

	tracerProvider, tracerShutdown := MustNewTracerProvider(ctx, resource)
	globalTracerProvider = tracerProvider

	metricProvider, metricShutdown := MustNewMetricProvider(ctx, resource)
	globalMetricProvider = metricProvider

	loggerProvider, loggerShutdown := MustNewLoggerProvider(ctx, resource)
	globalLoggerProvider = loggerProvider

	globalTracer = tracerProvider.Tracer(pkgName)
	globalMeter = metricProvider.Meter(pkgName)
	globalLogger = otelzap.New(zap.NewExample(),
		otelzap.WithLoggerProvider(loggerProvider),
		otelzap.WithMinLevel(zapcore.InfoLevel),
	)

	return func(ctx context.Context) error {
		if err := tracerShutdown(ctx); err != nil {
			log.Println(err)
		}
		if err := metricShutdown(ctx); err != nil {
			log.Println(err)
		}
		if err := loggerShutdown(ctx); err != nil {
			log.Println(err)
		}
		return nil
	}

}
