# AGENTS.md - zaplog Development Guide

This is a Go library that wraps [zap](https://github.com/uber-go/zap) with OpenTelemetry integration and structured logging capabilities.

## Build & Test Commands

```bash
# Build the entire package
go build ./...

# Run tests (when available)
go test ./...

# Run specific test file
go test -v ./path/to/file_test.go -run TestFunctionName

# Run examples
go run example/withContext/main.go
go run example/noContext/main.go

# Lint
go vet ./...

# Format code
go fmt ./...

# Update dependencies
go mod tidy
go mod download
```

## Code Style Guidelines

### Import Order
Standard library first, then third-party imports. Group with blank lines:
```go
import (
    "context"
    "time"

    "github.com/uptrace/opentelemetry-go-extra/otelzap"
    "go.uber.org/zap"
)
```

### Naming Conventions
- **Exported functions**: PascalCase (`InitLogger`, `GetDefaultLogger`)
- **Exported variables**: PascalCase (`DefaultLogger`, `LabelLevel`)
- **Internal variables**: camelCase (`myMeter`, `zapLogConfig`)
- **Constants**: PascalCase
- **Package name**: lowercase, matches directory name (`zaplog`)

### File Organization
- Package-level file: `zaplog.go` (main exports, initialization)
- Configuration/options: `option.go`
- Metrics: `otel_metrics.go`
- Examples: `example/` directory

### Options Pattern
Use functional options for configuration:
```go
type Option func(c *Config)

func WithMaxSize(maxSize int) Option {
    return func(c *Config) {
        c.MaxSize = maxSize
    }
}
```

### Logging Functions
Provide both context and non-context variants:
```go
func DebugContext(ctx context.Context, msg string, fields ...zapcore.Field)
func Debug(msg string, fields ...zapcore.Field)
```

### OpenTelemetry Integration
- Use `otelzap.Logger` for tracing-aware logging
- Initialize metrics in `init()` function: `otel.GetMeterProvider().Meter("zaplog")`
- Use attribute.String for metric labels

### Error Handling
- Use `panic()` for unrecoverable initialization errors in examples
- Use `log.Fatalf()` for fatal errors
- Keep error handling simple - this is a logging library, errors are not extensively wrapped

### GoDoc Comments
Exported functions must have GoDoc comments:
```go
// InitLogger initializes the default logger with the specified configuration.
func InitLogger(logPath string, level string, opts ...Option)
```

### Constants & Globals
Group package-level variables at top:
```go
const LabelLevel = "level"

var (
    DefaultLogger *otelzap.Logger
    zapLogConfig  Config
)
```

### Testing (When Adding)
- Test files: `*_test.go` in same package
- Test functions: `Test<FunctionName>`
- Use `testing.T` and testify assertions if needed

## Project Structure
```
zaplog/
├── zaplog.go           # Main package exports, logging functions
├── option.go           # Functional options for configuration
├── otel_metrics.go     # OpenTelemetry metrics initialization
├── example/            # Usage examples
│   ├── withContext/main.go
│   └── noContext/main.go
└── go.mod              # Go module definition (Go 1.24)
```

## Key Dependencies
- `go.uber.org/zap` - Core logging library
- `github.com/uptrace/opentelemetry-go-extra/otelzap` - OpenTelemetry integration
- `go.opentelemetry.io/otel/*` - OpenTelemetry SDK
- `gopkg.in/natefinch/lumberjack.v2` - Log rotation

## Notes
- No build scripts (Makefile, shell scripts) - use standard Go toolchain
- No CI/CD configuration files present
- Package uses Go 1.24+ features
- Vendor directory present for dependency management
