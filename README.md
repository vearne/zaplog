# zaplog

Simple packaging for [zap](https://github.com/uber-go/zap) with OpenTelemetry hooks.

## Install

```bash
go get github.com/vearne/zaplog
```

## Init

```go
// defaults: level=info, output=stdout
zlog.InitLogger()

// file only
zlog.InitLogger(zlog.WithFile("/var/log/app.log"), zlog.WithLevel("debug"))

// file + stdout, JSON (typical container setup)
zlog.InitLogger(
    zlog.WithFile("/var/log/app.log"),
    zlog.WithStdout(),
    zlog.WithJSON(),
    zlog.WithLevel("info"),
)

defer zlog.Sync()
```

### Options

| Option | Meaning |
|--------|---------|
| `WithLevel(string)` | debug / info / warn / error (default info) |
| `WithFile(path)` | lumberjack file output |
| `WithStdout()` / `WithStderr()` | std sinks |
| `WithJSON()` | JSON encoder instead of console |
| `WithCallerSkip(n)` | extra caller frames to skip (default 1) |
| `WithTraceID` / `WithAutoEvents` | OTEL helpers |
| `WithMaxSize` / `WithMaxAge` / `WithMaxBackups` / `WithCompress` | rotation |

**Output rule:** if no output option is set, **stdout** is used. Once any of
`WithFile` / `WithStdout` / `WithStderr` is set, only those sinks are enabled.

### Runtime

```go
zlog.SetLevel("debug")
_ = zlog.Sync() // flush before exit
```

### Named

`Named()` returns `*otelzap.Logger` and **bypasses** package helpers
(`statistics` / `trace_id` / auto span events). Use package-level
`Info` / `InfoContext` when you need those.

## Examples

Runnable demos live under [`example/`](./example). See [`example/README.md`](./example/README.md).

```bash
go run ./example/basic
go run ./example/file
go run ./example/teeJSON
go run ./example/noContext      # + Prometheus on :9090
go run ./example/withContext   # + trace_id / span events
```

## Minimal snippet

```go
package main

import (
	zlog "github.com/vearne/zaplog"
	"go.uber.org/zap"
)

func main() {
	if err := zlog.InitLogger(); err != nil {
		panic(err)
	}
	defer zlog.Sync()

	zlog.Info("hello", zap.String("from", "zaplog"))
}
```
