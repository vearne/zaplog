package zaplog

import "gopkg.in/natefinch/lumberjack.v2"

type Option func(l *lumberjack.Logger)

// megabytes
func WithMaxSize(maxSize int) Option {
	return func(l *lumberjack.Logger) {
		l.MaxSize = maxSize
	}
}

// days
func WithMaxAge(days int) Option {
	return func(l *lumberjack.Logger) {
		l.MaxAge = days
	}
}

func WithMaxBackups(maxBackups int) Option {
	return func(l *lumberjack.Logger) {
		l.MaxBackups = maxBackups
	}
}

func WithCompress(compress bool) Option {
	return func(l *lumberjack.Logger) {
		l.Compress = compress
	}
}
