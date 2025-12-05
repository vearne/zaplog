package zaplog

import "gopkg.in/natefinch/lumberjack.v2"

type Config struct {
	lumberjack.Logger
	WithTraceID bool
}
type Option func(c *Config)

// megabytes
func WithMaxSize(maxSize int) Option {
	return func(c *Config) {
		c.MaxSize = maxSize
	}
}

// days
func WithMaxAge(days int) Option {
	return func(c *Config) {
		c.MaxAge = days
	}
}

func WithMaxBackups(maxBackups int) Option {
	return func(c *Config) {
		c.MaxBackups = maxBackups
	}
}

func WithCompress(compress bool) Option {
	return func(c *Config) {
		c.Compress = compress
	}
}

// WithTraceIDField configures the logger to add `trace_id` field to structured log messages.
func WithTraceID(on bool) Option {
	return func(c *Config) {
		c.WithTraceID = on
	}
}
