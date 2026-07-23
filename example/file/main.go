// File-only logging with rotation options.
// Parent directories are created automatically when missing.
//
//	go run ./example/file
package main

import (
	"os"
	"path/filepath"

	zlog "github.com/vearne/zaplog"
	"go.uber.org/zap"
)

func main() {
	logPath := filepath.Join(os.TempDir(), "zaplog-example", "app.log")

	// WithFile alone disables the default stdout sink.
	if err := zlog.InitLogger(
		zlog.WithFile(logPath),
		zlog.WithLevel("debug"),
		zlog.WithMaxSize(1),    // MB
		zlog.WithMaxBackups(3),
		zlog.WithMaxAge(7),     // days
		zlog.WithCompress(false),
	); err != nil {
		panic(err)
	}
	defer zlog.Sync()

	zlog.Debug("debug to file", zap.String("path", logPath))
	zlog.Info("info to file")
	zlog.Warn("warn to file")
	zlog.Error("error to file", zap.String("hint", "check the log file"))

	println("wrote logs to", logPath)
}
