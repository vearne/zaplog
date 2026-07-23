package zaplog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const LabelLevel = "level"

var (
	// DefaultLogger is never nil: before InitLogger it discards all logs (zap.NewNop).
	DefaultLogger = otelzap.New(zap.NewNop())
	zapLogConfig  Config
	atomicLevel   zap.AtomicLevel
)

func GetDefaultLogger() *otelzap.Logger {
	return DefaultLogger
}

// InitLogger initializes the default logger.
//
// Defaults when called with no options:
//   - level: info
//   - output: stdout
//
// Examples:
//
//	InitLogger()
//	InitLogger(WithFile("/var/log/app.log"))
//	InitLogger(WithFile(path), WithStdout(), WithLevel("debug"), WithJSON())
func InitLogger(opts ...Option) error {
	zapLogConfig = Config{
		Logger: lumberjack.Logger{
			MaxSize:    1024, // megabytes
			MaxBackups: 3,
			MaxAge:     7, // days
			Compress:   false,
		},
		WithTraceID:    false,
		WithAutoEvents: true,
		CallerSkip:     1,
		Level:          "info",
	}

	for _, opt := range opts {
		opt(&zapLogConfig)
	}

	if !zapLogConfig.outputSet {
		zapLogConfig.AlsoStdout = true
	}

	if zapLogConfig.FilePath != "" {
		if err := ensureLogDir(zapLogConfig.FilePath); err != nil {
			return err
		}
		zapLogConfig.Filename = zapLogConfig.FilePath
	}

	if zapLogConfig.FilePath == "" && !zapLogConfig.AlsoStdout && !zapLogConfig.AlsoStderr {
		return fmt.Errorf("zaplog: no output configured")
	}

	atomicLevel = zap.NewAtomicLevel()
	if err := applyLevel(zapLogConfig.Level); err != nil {
		return err
	}

	callerSkip := zapLogConfig.CallerSkip
	if callerSkip < 0 {
		callerSkip = 0
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	encoderConfig.ConsoleSeparator = " | "

	var encoder zapcore.Encoder
	if zapLogConfig.JSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var cores []zapcore.Core
	if zapLogConfig.FilePath != "" {
		cores = append(cores, zapcore.NewCore(encoder.Clone(), zapcore.AddSync(&zapLogConfig), atomicLevel))
	}
	if zapLogConfig.AlsoStdout {
		cores = append(cores, zapcore.NewCore(encoder.Clone(), zapcore.AddSync(os.Stdout), atomicLevel))
	}
	if zapLogConfig.AlsoStderr {
		cores = append(cores, zapcore.NewCore(encoder.Clone(), zapcore.AddSync(os.Stderr), atomicLevel))
	}

	logger := zap.New(zapcore.NewTee(cores...))
	logger = logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(callerSkip))

	DefaultLogger = otelzap.New(logger,
		otelzap.WithMinLevel(zap.DebugLevel),
		otelzap.WithCallerDepth(callerSkip),
		otelzap.WithErrorStatusLevel(zap.ErrorLevel),
	)
	return nil
}

// SetLevel changes the minimum log level at runtime.
// Supported: debug, info, warn, error. Must be called after InitLogger.
func SetLevel(level string) error {
	return applyLevel(level)
}

// Level returns the current minimum log level.
func Level() zapcore.Level {
	return atomicLevel.Level()
}

// AtomicLevel returns the underlying zap.AtomicLevel for advanced use.
func AtomicLevel() zap.AtomicLevel {
	return atomicLevel
}

// Sync flushes any buffered log entries. Call before process exit.
func Sync() error {
	return DefaultLogger.Sync()
}

func applyLevel(level string) error {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "", "info":
		atomicLevel.SetLevel(zap.InfoLevel)
	case "debug":
		atomicLevel.SetLevel(zap.DebugLevel)
	case "warn", "warning":
		atomicLevel.SetLevel(zap.WarnLevel)
	case "error":
		atomicLevel.SetLevel(zap.ErrorLevel)
	default:
		return fmt.Errorf("zaplog: unknown level %q", level)
	}
	return nil
}

// ensureLogDir validates logPath and creates its parent directory when missing.
func ensureLogDir(logPath string) error {
	if strings.TrimSpace(logPath) == "" {
		return fmt.Errorf("zaplog: log path is empty")
	}
	dir := filepath.Dir(logPath)
	fi, err := os.Stat(dir)
	if err == nil {
		if !fi.IsDir() {
			return fmt.Errorf("zaplog: log path parent is not a directory: %s", dir)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("zaplog: stat log dir: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("zaplog: create log dir: %w", err)
	}
	return nil
}

func statistics(level zapcore.Level) {
	if DefaultLogger.Core().Enabled(level) {
		logTotal.Add(context.Background(),
			1,
			metric.WithAttributes(
				attribute.String(LabelLevel, level.String()),
			),
		)
	}
}

func tryAddFields(ctx context.Context, fields []zapcore.Field) []zapcore.Field {
	if zapLogConfig.WithTraceID {
		traceID := GetTraceID(ctx)
		if len(traceID) > 0 {
			fields = append(fields, zap.String("trace_id", traceID))
		}
	}
	return fields
}

// Named returns a child logger with the given name.
//
// Logs written through the returned *otelzap.Logger bypass zaplog package
// helpers: they do NOT go through statistics / trace_id injection / auto span
// events. Use package-level Info/InfoContext/... when you need those features.
func Named(s string) *otelzap.Logger {
	l := DefaultLogger.Clone()
	l.Logger = l.Logger.Named(s)
	return l
}

func zapFieldsToAttributes(fields []zapcore.Field) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(fields))
	for _, f := range fields {
		switch f.Type {
		case zapcore.StringType:
			attrs = append(attrs, attribute.String(f.Key, f.String))
		case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
			attrs = append(attrs, attribute.Int64(f.Key, f.Integer))
		case zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type:
			attrs = append(attrs, attribute.Int64(f.Key, int64(f.Integer)))
		case zapcore.Float64Type:
			attrs = append(attrs, attribute.Float64(f.Key, f.Interface.(float64)))
		case zapcore.BoolType:
			attrs = append(attrs, attribute.Bool(f.Key, f.Integer == 1))
		case zapcore.TimeType:
			attrs = append(attrs, attribute.String(f.Key, f.Interface.(time.Time).Format(time.RFC3339)))
		case zapcore.StringerType:
			attrs = append(attrs, attribute.String(f.Key, f.Interface.(fmt.Stringer).String()))
		case zapcore.ReflectType, zapcore.ArrayMarshalerType, zapcore.ObjectMarshalerType:
			if jsonBytes, err := json.Marshal(f.Interface); err == nil {
				attrs = append(attrs, attribute.String(f.Key, string(jsonBytes)))
			} else {
				attrs = append(attrs, attribute.String(f.Key, fmt.Sprintf("%v", f.Interface)))
			}
		default:
			attrs = append(attrs, attribute.String(f.Key, fmt.Sprintf("%v", f.Interface)))
		}
	}
	return attrs
}

func addEventIfEnabled(ctx context.Context, level, msg string, fields ...zapcore.Field) {
	if zapLogConfig.WithAutoEvents {
		span := trace.SpanFromContext(ctx)
		if span.SpanContext().IsValid() {
			attrs := []attribute.KeyValue{
				attribute.String("level", level),
				attribute.String("message", msg),
			}
			attrs = append(attrs, zapFieldsToAttributes(fields)...)
			span.AddEvent("log", trace.WithAttributes(attrs...))
		}
	}
}

func DebugContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	statistics(zap.DebugLevel)
	fields = tryAddFields(ctx, fields)
	addEventIfEnabled(ctx, "debug", msg, fields...)
	DefaultLogger.DebugContext(ctx, msg, fields...)
}

func InfoContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	statistics(zap.InfoLevel)
	fields = tryAddFields(ctx, fields)
	addEventIfEnabled(ctx, "info", msg, fields...)
	DefaultLogger.InfoContext(ctx, msg, fields...)
}

func WarnContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	statistics(zap.WarnLevel)
	fields = tryAddFields(ctx, fields)
	addEventIfEnabled(ctx, "warn", msg, fields...)
	DefaultLogger.WarnContext(ctx, msg, fields...)
}

func ErrorContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	statistics(zap.ErrorLevel)
	fields = tryAddFields(ctx, fields)
	addEventIfEnabled(ctx, "error", msg, fields...)
	DefaultLogger.ErrorContext(ctx, msg, fields...)
}

func FatalContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	statistics(zap.FatalLevel)
	fields = tryAddFields(ctx, fields)
	addEventIfEnabled(ctx, "fatal", msg, fields...)
	DefaultLogger.FatalContext(ctx, msg, fields...)
}

func Debug(msg string, fields ...zapcore.Field) {
	statistics(zap.DebugLevel)
	DefaultLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zapcore.Field) {
	statistics(zap.InfoLevel)
	DefaultLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zapcore.Field) {
	statistics(zap.WarnLevel)
	DefaultLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	statistics(zap.ErrorLevel)
	DefaultLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zapcore.Field) {
	statistics(zap.FatalLevel)
	DefaultLogger.Fatal(msg, fields...)
}

// GetTraceID returns the 32-character lowercase hexadecimal string of the TraceID from the current context.
// Returns an empty string if there is no Span in the ctx.
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}
	return span.SpanContext().TraceID().String()
}
