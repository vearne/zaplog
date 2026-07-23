package zaplog

import "gopkg.in/natefinch/lumberjack.v2"

type Config struct {
	lumberjack.Logger
	WithTraceID    bool
	WithAutoEvents bool
	// CallerSkip is the number of additional stack frames to skip when
	// reporting the caller. Default is 1 (skip zaplog's own Info/Debug/... wrappers).
	// Increase by 1 for each extra wrapper layer around zaplog.
	CallerSkip int

	Level string

	// Output sinks. If none are set, InitLogger defaults to stdout.
	FilePath   string
	AlsoStdout bool
	AlsoStderr bool
	outputSet  bool // true if any WithFile/WithStdout/WithStderr was used

	JSON bool
}

type Option func(c *Config)

// WithLevel sets the minimum log level. Supported: debug, info, warn, error.
// Default is "info".
func WithLevel(level string) Option {
	return func(c *Config) {
		c.Level = level
	}
}

// WithFile enables file output at path (lumberjack rotation).
func WithFile(path string) Option {
	return func(c *Config) {
		c.FilePath = path
		c.Filename = path
		c.outputSet = true
	}
}

// WithStdout enables writing logs to os.Stdout.
func WithStdout() Option {
	return func(c *Config) {
		c.AlsoStdout = true
		c.outputSet = true
	}
}

// WithStderr enables writing logs to os.Stderr.
func WithStderr() Option {
	return func(c *Config) {
		c.AlsoStderr = true
		c.outputSet = true
	}
}

// WithJSON switches the encoder from console to JSON.
func WithJSON() Option {
	return func(c *Config) {
		c.JSON = true
	}
}

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

// WithTraceID configures the logger to add `trace_id` field to structured log messages.
func WithTraceID(on bool) Option {
	return func(c *Config) {
		c.WithTraceID = on
	}
}

// WithAutoEvents configures the logger to automatically add log messages as events to active spans.
// When enabled, each log call will add an event to the current span with the log level and message.
func WithAutoEvents(enable bool) Option {
	return func(c *Config) {
		c.WithAutoEvents = enable
	}
}

// WithCallerSkip sets how many extra stack frames to skip when annotating logs with caller
// file:line. The default is 1 so direct zaplog.Info/Debug/... calls report the application
// call site. Pass 2 if you wrap zaplog in another package (e.g. applog.Info -> zaplog.Info).
func WithCallerSkip(skip int) Option {
	return func(c *Config) {
		c.CallerSkip = skip
	}
}
