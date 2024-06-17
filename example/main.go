package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	zlog "github.com/vearne/zaplog"
	"go.uber.org/zap"
	"net/http"
	"sync/atomic"
	"time"
)

var ops uint64

func main() {
	zlog.InitLogger("/tmp/otel.log", "warn")
	go func() {
		for {
			atomic.AddUint64(&ops, 1)

			zlog.Debug("test debug", zap.Uint64("ops", atomic.LoadUint64(&ops)))
			zlog.Debug("test debug", zap.Uint64("ops", atomic.LoadUint64(&ops)))
			zlog.Debug("test debug", zap.Uint64("ops", atomic.LoadUint64(&ops)))

			zlog.Info("test info", zap.Uint64("ops", atomic.LoadUint64(&ops)))
			zlog.Info("test info", zap.Uint64("ops", atomic.LoadUint64(&ops)))
			zlog.Warn("test warn", zap.Uint64("ops", atomic.LoadUint64(&ops)))
			zlog.Error("test error", zap.Uint64("ops", atomic.LoadUint64(&ops)))
			time.Sleep(200 * time.Millisecond)
		}

	}()

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("starting...")
	// http://localhost:9090/metrics
	http.ListenAndServe(":9090", nil)
}
