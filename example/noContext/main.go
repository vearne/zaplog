// Non-context logging + OpenTelemetry metrics (log_total counter).
//
//	go run ./example/noContext
//	curl http://localhost:9090/metrics | grep log_total
package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	zlog "github.com/vearne/zaplog"
	"go.opentelemetry.io/otel"
	otelProm "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/zap"
)

var ops1 uint64
var ops2 uint64

func main() {
	initMeterProvider()

	if err := zlog.InitLogger(
		zlog.WithFile("/tmp/zaplog-noContext.log"),
		zlog.WithStdout(),
		zlog.WithLevel("warn"),
		zlog.WithMaxSize(1),
		zlog.WithMaxBackups(3),
		zlog.WithMaxAge(3),
		zlog.WithCompress(true),
	); err != nil {
		log.Fatal(err)
	}
	defer zlog.Sync()

	go func() {
		for {
			atomic.AddUint64(&ops1, 1)
			// Dropped: level is warn.
			zlog.Info("test info1", zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			zlog.Warn("test warn1", zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			zlog.Error("test error1", zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			time.Sleep(200 * time.Millisecond)
		}
	}()
	go func() {
		for {
			atomic.AddUint64(&ops2, 1)
			zlog.Warn("test warn2", zap.Uint64("ops", atomic.LoadUint64(&ops2)))
			zlog.Error("test error2", zap.Uint64("ops", atomic.LoadUint64(&ops2)))
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
