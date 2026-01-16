package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	zlog "github.com/vearne/zaplog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelProm "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
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
		zlog.WithAutoEvents(true),
	)

	go func() {
		//logger1 := zlog.Named("worker1")
		for {
			atomic.AddUint64(&ops1, 1)
			ctx, span := otel.GetTracerProvider().Tracer("test").Start(context.Background(),
				fmt.Sprintf("test:%v", atomic.LoadUint64(&ops1)))
			zlog.InfoContext(ctx, "test info1", zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			zlog.InfoContext(ctx, "test info2",
				zap.Any("testAny", struct {
					Age  int
					Name string
				}{Age: 12, Name: "testName"}))
			zlog.WarnContext(ctx, "test warn1", zap.Strings("strs", []string{"aaa", "bbb", "ccc"}),
				zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			//zlog.ErrorContext(ctx, "test error1", zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			span.End()

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

	// Create stdout exporter for debugging
	stdoutExporter, err := stdouttrace.New(stdouttrace.WithWriter(os.Stderr))
	if err != nil {
		log.Fatalf("failed to create stdout exporter: %v", err)
	}

	// Create OTLP HTTP exporter for production
	httpExporter, err := otlptracehttp.New(ctx, otlptracehttp.WithInsecure())
	if err != nil {
		log.Printf("failed to create OTLP HTTP exporter: %v (continuing with stdout only)", err)
		// Use only stdout exporter if OTLP fails
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSyncer(stdoutExporter),
		)
		otel.SetTracerProvider(tp)
		return tp
	}

	// Use both exporters: stdout (sync) and OTLP (batched)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(stdoutExporter),
		sdktrace.WithBatcher(httpExporter),
	)
	otel.SetTracerProvider(tp)
	return tp
}
