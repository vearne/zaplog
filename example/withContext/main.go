package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	zlog "github.com/vearne/zaplog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelProm "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

var ops1 uint64

//var ops2 uint64

func main() {
	InitTracerProvider()
	InitMeterProvider()

	zlog.InitLogger("/tmp/withContext.log", "debug",
		zlog.WithCompress(false),
		zlog.WithMaxAge(3),
		zlog.WithMaxBackups(3),
		zlog.WithMaxSize(1),
		zlog.WithTraceID(true),
	)

	go func() {
		//logger1 := zlog.Named("worker1")
		for {
			atomic.AddUint64(&ops1, 1)
			ctx, span := otel.GetTracerProvider().Tracer("test").Start(context.Background(),
				fmt.Sprintf("test:%v", atomic.LoadUint64(&ops1)))
			defer span.End()

			zlog.InfoContext(ctx, "test info1", zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			zlog.WarnContext(ctx, "test warn1", zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			zlog.ErrorContext(ctx, "test error1", zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			time.Sleep(200 * time.Millisecond)
		}

	}()
	go func() {
		//logger2 := zlog.Named("worker2")
		ctx := context.Background()
		for {
			atomic.AddUint64(&ops1, 1)
			zlog.InfoContext(ctx, "test info2", zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			zlog.WarnContext(ctx, "test warn2", zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			zlog.ErrorContext(ctx, "test error2", zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			time.Sleep(200 * time.Millisecond)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("starting...")
	// http://localhost:9090/metrics
	http.ListenAndServe(":9090", nil)
}

func InitMeterProvider() *sdkmetric.MeterProvider {
	promExporter, err := otelProm.New(otelProm.WithNamespace("otel-metrics"),
		otelProm.WithRegisterer(prometheus.DefaultRegisterer),
	)
	if err != nil {
		panic(err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(promExporter),
	)
	otel.SetMeterProvider(mp)

	return mp
}

func InitTracerProvider() *sdktrace.TracerProvider {
	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		log.Fatalf("new otlp trace grpc exporter failed: %v", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}
