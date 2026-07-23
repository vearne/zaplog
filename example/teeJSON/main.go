// Container-style setup: JSON logs to stdout, optionally also to a file (tee).
//
//	go run ./example/teeJSON
package main

import (
	"os"
	"path/filepath"

	zlog "github.com/vearne/zaplog"
	"go.uber.org/zap"
)

func main() {
	logPath := filepath.Join(os.TempDir(), "zaplog-example", "tee.log")

	// Explicit sinks: once any WithFile/WithStdout/WithStderr is set,
	// only those sinks are used (no implicit default stdout).
	if err := zlog.InitLogger(
		zlog.WithFile(logPath),
		zlog.WithStdout(),
		zlog.WithJSON(),
		zlog.WithLevel("info"),
	); err != nil {
		panic(err)
	}
	defer zlog.Sync()

	zlog.Info("json log to file and stdout",
		zap.String("service", "api"),
		zap.Int("port", 8080),
	)
	zlog.Warn("json warn", zap.Strings("tags", []string{"tee", "json"}))

	println("also wrote to", logPath)
}
