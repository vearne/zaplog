// Context-aware logging with trace_id injection and auto span events.
//
//	go run ./example/withContext
//	curl http://localhost:9090/metrics | grep log_total
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

var ops uint64

func main() {
	initTracerProvider()
	initMeterProvider()

	if err := zlog.InitLogger(
		zlog.WithFile("/tmp/zaplog-withContext.log"),
		zlog.WithStdout(),
		zlog.WithLevel("debug"),
		zlog.WithTraceID(true),
		zlog.WithAutoEvents(true),
		zlog.WithMaxSize(1),
		zlog.WithMaxBackups(3),
		zlog.WithMaxAge(3),
	); err != nil {
		log.Fatal(err)
	}
	defer zlog.Sync()

	go func() {
		tracer := otel.GetTracerProvider().Tracer("zaplog-example")
		for {
			n := atomic.AddUint64(&ops, 1)
			ctx, span := tracer.Start(context.Background(), fmt.Sprintf("work:%d", n))

			zlog.InfoContext(ctx, "traced info",
				zap.Uint64("ops", n),
				zap.Any("user", struct {
					Age  int
					Name string
				}{Age: 12, Name: "alice"}),
			)
			zlog.WarnContext(ctx, "traced warn",
				zap.Strings("tags", []string{"otel", "trace"}),
				zap.Uint64("ops", n),
			)
			span.End()
			time.Sleep(200 * time.Millisecond)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("metrics: http://localhost:9090/metrics")
	log.Fatal(http.ListenAndServe(":9090", nil))
}

func initMeterProvider() *sdkmetric.MeterProvider {
	promExporter, err := otelProm.New(
		otelProm.WithNamespace("otel-metrics"),
		otelProm.WithRegisterer(prometheus.DefaultRegisterer),
	)
	if err != nil {
		panic(err)
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(promExporter))
	otel.SetMeterProvider(mp)
	return mp
}

func initTracerProvider() *sdktrace.TracerProvider {
	ctx := context.Background()

	stdoutExporter, err := stdouttrace.New(stdouttrace.WithWriter(os.Stderr))
	if err != nil {
		log.Fatalf("stdout trace exporter: %v", err)
	}

	httpExporter, err := otlptracehttp.New(ctx, otlptracehttp.WithInsecure())
	if err != nil {
		log.Printf("OTLP exporter unavailable (%v); using stdout only", err)
		tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(stdoutExporter))
		otel.SetTracerProvider(tp)
		return tp
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(stdoutExporter),
		sdktrace.WithBatcher(httpExporter),
	)
	otel.SetTracerProvider(tp)
	return tp
}
