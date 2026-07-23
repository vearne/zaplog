# Examples

Run from the module root (`github.com/vearne/zaplog`).

| Dir | What it shows | Command |
|-----|---------------|---------|
| [`basic`](./basic) | `InitLogger()` defaults, `SetLevel`, `Sync`, `Named` | `go run ./example/basic` |
| [`file`](./file) | `WithFile` + rotation; auto-create parent dirs | `go run ./example/file` |
| [`teeJSON`](./teeJSON) | File + stdout tee, JSON encoder (container-friendly) | `go run ./example/teeJSON` |
| [`noContext`](./noContext) | Non-context logs + Prometheus `log_total` metrics | `go run ./example/noContext` |
| [`withContext`](./withContext) | `InfoContext` + `WithTraceID` / `WithAutoEvents` | `go run ./example/withContext` |

## Quick start

```bash
# zero-config: info → stdout
go run ./example/basic

# file only
go run ./example/file

# JSON to stdout and file
go run ./example/teeJSON
```

## Init cheat sheet

```go
zlog.InitLogger() // level=info, stdout

zlog.InitLogger(zlog.WithFile("/var/log/app.log"), zlog.WithLevel("debug"))

zlog.InitLogger(
    zlog.WithFile("/var/log/app.log"),
    zlog.WithStdout(),
    zlog.WithJSON(),
)

zlog.SetLevel("debug")
defer zlog.Sync()
```

**Output rule:** if you pass no `WithFile` / `WithStdout` / `WithStderr`, stdout is used.
As soon as any of those is set, only the sinks you listed are enabled.
