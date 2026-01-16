package zaplog

import (
	"context"
	"encoding/json"
	"fmt"
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
		WithTraceID:    false,
		WithAutoEvents: true,
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
		otelzap.WithErrorStatusLevel(zap.ErrorLevel), // Error 日志设置 span 状态为 Error
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
