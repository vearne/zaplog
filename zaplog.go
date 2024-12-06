package zaplog

import (
	"context"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

const LabelLevel = "level"

var (
	DefaultLogger *otelzap.Logger
)

func GetDefaultLogger() *otelzap.Logger {
	return DefaultLogger
}

func InitLogger(logPath string, level string) {
	alevel := zap.NewAtomicLevel()

	hook := lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    1024, // megabytes
		MaxBackups: 10,
		MaxAge:     7,    //days
		Compress:   true, // disabled by default
	}
	w := zapcore.AddSync(&hook)

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
		otelzap.WithTraceIDField(true),
		otelzap.WithMinLevel(zap.DebugLevel),
		otelzap.WithCallerDepth(1),
	)
}

func Named(s string) *otelzap.Logger {
	l := DefaultLogger.Clone()
	l.Logger = l.Logger.Named(s)
	return l
}

func DebugContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	if DefaultLogger.Level() <= zap.DebugLevel {
		logTotal.Add(context.Background(),
			1,
			metric.WithAttributes(
				attribute.String(LabelLevel, "debug"),
			),
		)
	}
	DefaultLogger.DebugContext(ctx, msg, fields...)
}

func InfoContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	if DefaultLogger.Level() <= zap.InfoLevel {
		logTotal.Add(context.Background(),
			1,
			metric.WithAttributes(
				attribute.String(LabelLevel, "info"),
			),
		)
	}
	DefaultLogger.InfoContext(ctx, msg, fields...)
}

func WarnContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	if DefaultLogger.Level() <= zap.WarnLevel {
		logTotal.Add(context.Background(),
			1,
			metric.WithAttributes(
				attribute.String(LabelLevel, "warn"),
			),
		)
	}
	DefaultLogger.WarnContext(ctx, msg, fields...)
}

func ErrorContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	if DefaultLogger.Level() <= zap.ErrorLevel {
		logTotal.Add(context.Background(),
			1,
			metric.WithAttributes(
				attribute.String(LabelLevel, "error"),
			),
		)
	}
	DefaultLogger.ErrorContext(ctx, msg, fields...)
}

func FatalContext(ctx context.Context, msg string, fields ...zapcore.Field) {
	logTotal.Add(context.Background(),
		1,
		metric.WithAttributes(
			attribute.String(LabelLevel, "fatal"),
		),
	)
	DefaultLogger.FatalContext(ctx, msg, fields...)
}

func Debug(msg string, fields ...zapcore.Field) {
	if DefaultLogger.Level() <= zap.DebugLevel {
		logTotal.Add(context.Background(),
			1,
			metric.WithAttributes(
				attribute.String(LabelLevel, "debug"),
			),
		)
	}
	DefaultLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zapcore.Field) {
	if DefaultLogger.Level() <= zap.InfoLevel {
		logTotal.Add(context.Background(),
			1,
			metric.WithAttributes(
				attribute.String(LabelLevel, "info"),
			),
		)
	}
	DefaultLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zapcore.Field) {
	if DefaultLogger.Level() <= zap.WarnLevel {
		logTotal.Add(context.Background(),
			1,
			metric.WithAttributes(
				attribute.String(LabelLevel, "warn"),
			),
		)
	}
	DefaultLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	if DefaultLogger.Level() <= zap.ErrorLevel {
		logTotal.Add(context.Background(),
			1,
			metric.WithAttributes(
				attribute.String(LabelLevel, "error"),
			),
		)
	}
	DefaultLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zapcore.Field) {
	logTotal.Add(context.Background(),
		1,
		metric.WithAttributes(
			attribute.String(LabelLevel, "fatal"),
		),
	)
	DefaultLogger.Fatal(msg, fields...)
}
