package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	zlog "github.com/vearne/zaplog"
	"go.opentelemetry.io/otel"
	otelProm "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/zap"
	"net/http"
	"sync/atomic"
	"time"
)

var ops1 uint64
var ops2 uint64

func main() {
	InitMeterProvider()

	zlog.InitLogger("/tmp/withContext.log", "debug")
	go func() {
		//logger1 := zlog.Named("worker1")
		ctx := context.Background()
		for {
			atomic.AddUint64(&ops1, 1)

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
			atomic.AddUint64(&ops2, 1)

			zlog.InfoContext(ctx, "test info2", zap.Uint64("ops", atomic.LoadUint64(&ops2)))
			zlog.WarnContext(ctx, "test warn2", zap.Uint64("ops", atomic.LoadUint64(&ops2)))
			zlog.ErrorContext(ctx, "test error2", zap.Uint64("ops", atomic.LoadUint64(&ops2)))
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
		//sdkmetric.WithResource(initResource()),
	)
	otel.SetMeterProvider(mp)

	return mp
}
