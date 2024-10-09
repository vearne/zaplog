package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	zlog "github.com/vearne/zaplog"
	"go.uber.org/zap"
	"net/http"
	"sync/atomic"
	"time"
)

var ops1 uint64
var ops2 uint64

func main() {
	zlog.InitLogger("/tmp/withContext.log", "warn")
	go func() {
		logger1 := zlog.Named("worker1")
		ctx := context.Background()
		for {
			atomic.AddUint64(&ops1, 1)

			logger1.InfoContext(ctx, "test info1", zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			logger1.WarnContext(ctx, "test warn1", zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			logger1.ErrorContext(ctx, "test error1", zap.Uint64("ops", atomic.LoadUint64(&ops1)))
			time.Sleep(200 * time.Millisecond)
		}

	}()
	go func() {
		logger2 := zlog.Named("worker2")
		ctx := context.Background()
		for {
			atomic.AddUint64(&ops2, 1)
			logger2.InfoContext(ctx, "test info2", zap.Uint64("ops", atomic.LoadUint64(&ops2)))
			logger2.WarnContext(ctx, "test warn2", zap.Uint64("ops", atomic.LoadUint64(&ops2)))
			logger2.ErrorContext(ctx, "test error2", zap.Uint64("ops", atomic.LoadUint64(&ops2)))
			time.Sleep(200 * time.Millisecond)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("starting...")
	// http://localhost:9090/metrics
	http.ListenAndServe(":9090", nil)
}
