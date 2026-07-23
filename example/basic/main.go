// Basic usage: zero-config InitLogger (default level=info, output=stdout),
// runtime SetLevel, and Sync on exit.
//
//	go run ./example/basic
package main

import (
	"fmt"
	"time"

	zlog "github.com/vearne/zaplog"
	"go.uber.org/zap"
)

func main() {
	// No options → level=info, write to stdout.
	if err := zlog.InitLogger(); err != nil {
		panic(err)
	}
	defer zlog.Sync()

	zlog.Info("app started", zap.String("env", "dev"))
	zlog.Debug("this debug line is dropped (level=info)")
	zlog.Warn("disk almost full", zap.Float64("used_pct", 92.5))

	// Raise verbosity at runtime.
	if err := zlog.SetLevel("debug"); err != nil {
		panic(err)
	}
	zlog.Debug("debug enabled after SetLevel")

	// Named() returns *otelzap.Logger and bypasses package helpers
	// (statistics / trace_id / auto span events). Prefer package-level
	// Info/InfoContext when you need those features.
	worker := zlog.Named("worker")
	worker.Info("named logger still works", zap.Int("n", 1))

	fmt.Println("--- done ---")
	time.Sleep(10 * time.Millisecond)
}
