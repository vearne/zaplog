package zaplog

import (
	"context"
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
	DefaultLogger *otelzap.Logger
	zapLogConfig  Config
)

func GetDefaultLogger() *otelzap.Logger {
	return DefaultLogger
}

func InitLogger(logPath string, level string, opts ...Option) {
	alevel := zap.NewAtomicLevel()

	zapLogConfig = Config{
		Logger: lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    1024, // megabytes
			MaxBackups: 3,
			MaxAge:     7,     //days
			Compress:   false, // disabled by default
		},
		WithTraceID: false,
	}

	for _, opt := range opts {
		opt(&zapLogConfig)
	}

	w := zapcore.AddSync(&zapLogConfig)

	switch level {
	case "debug":
		alevel.SetLevel(zap.DebugLevel)
	case "info":
		alevel.SetLevel(zap.InfoLevel)
	case "warn":
		alevel.SetLevel(zap.WarnLevel)
	case "error":
		alevel.SetLevel(zap.ErrorLevel)
	default:
		alevel.SetLevel(zap.InfoLevel)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	encoderConfig.ConsoleSeparator = " | "

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		alevel,
	)

	logger := zap.New(core)
	logger = logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1))

	DefaultLogger = otelzap.New(logger,
		otelzap.WithMinLevel(zap.DebugLevel),
		otelzap.WithCallerDepth(1),
	)
}

func statistics(level zapcore.Level) {
	if DefaultLogger.Level() >= level {
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

func Named(s string) *otelzap.Logger {
	l := DefaultLogger.Clone()
	l.Logger = l.Logger.Named(s)
	return l
}

func DebugContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	statistics(zap.DebugLevel)
	fields = tryAddFields(ctx, fields)
	DefaultLogger.DebugContext(ctx, msg, fields...)
}

func InfoContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	statistics(zap.InfoLevel)
	fields = tryAddFields(ctx, fields)
	DefaultLogger.InfoContext(ctx, msg, fields...)
}

func WarnContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	statistics(zap.WarnLevel)
	fields = tryAddFields(ctx, fields)
	DefaultLogger.WarnContext(ctx, msg, fields...)
}

func ErrorContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	statistics(zap.ErrorLevel)
	fields = tryAddFields(ctx, fields)
	DefaultLogger.ErrorContext(ctx, msg, fields...)
}

func FatalContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	statistics(zap.FatalLevel)
	fields = tryAddFields(ctx, fields)
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
