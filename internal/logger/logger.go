package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	global       *zap.SugaredLogger
	defaultLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
)

func init() { //nolint: gochecknoinits // If the logging level is not set, the application will have no logs.
	SetLogger(New(defaultLevel))
}

// New creates a new instance of *zap.SugaredLogger with output in simple console format.
// If the logging level is not provided, the default level (zap.ErrorLevel) will be used.
func New(level zapcore.LevelEnabler, options ...zap.Option) *zap.SugaredLogger {
	if level == nil {
		level = defaultLevel
	}

	defaultEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:     "message",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
	core := zapcore.NewCore(
		defaultEncoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	return zap.New(core, options...).Sugar()
}

// Level returns the current logging level of the global logger.
func Level() zapcore.Level {
	return defaultLevel.Level()
}

// Logger returns the global logger.
func Logger() *zap.SugaredLogger {
	return global
}

// SetLogger sets the global logger.
// This function is not thread-safe.
func SetLogger(l *zap.SugaredLogger) {
	global = l
}

// SetLevel sets the log level for the global logger.
func SetLevel(level zapcore.Level) {
	//nolint: errcheck // No need to check the error here.
	defer global.Sync()

	defaultLevel.SetLevel(level)
}

// Debug writes a debug level message using the logger from the context.
func Debug(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Debug(args...)
}

// Debugf writes a formatted debug level message using the logger from the context.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Debugf(format, args...)
}

// DebugKV writes a message and key-value pairs
// at the debug level using the logger from the context.
func DebugKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Debugw(message, kvs...)
}

// Info writes an information level message using the logger from the context.
func Info(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Info(args...)
}

// Infof writes a formatted information level message using the logger from the context.
func Infof(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Infof(format, args...)
}

// InfoKV writes a message and key-value pairs
// at the information level using the logger from the context.
func InfoKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Infow(message, kvs...)
}

// Warn writes a warning level message using the logger from the context.
func Warn(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Warn(args...)
}

// Warnf writes a formatted warning level message using the logger from the context.
func Warnf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Warnf(format, args...)
}

// WarnKV writes a message and key-value pairs
// at the warning level using the logger from the context.
func WarnKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Warnw(message, kvs...)
}

// Error writes an error level message using the logger from the context.
func Error(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Error(args...)
}

// Errorf writes a formatted error level message using the logger from the context.
func Errorf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Errorf(format, args...)
}

// ErrorKV writes a message and key-value pairs
// at the error level using the logger from the context.
func ErrorKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Errorw(message, kvs...)
}

// Fatal writes a fatal error level message
// using the logger from the context and then calls os.Exit(1).
func Fatal(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Fatal(args...)
}

// Fatalf writes a formatted fatal error level message
// using the logger from the context and then calls os.Exit(1).
func Fatalf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Fatalf(format, args...)
}

// FatalKV writes a message and key-value pairs
// at the fatal error level using the logger from the context
// and then calls os.Exit(1).
func FatalKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Fatalw(message, kvs...)
}

// Panic writes a panic level message
// using the logger from the context and then calls panic().
func Panic(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Panic(args...)
}

// Panicf writes a formatted panic level message
// using the logger from the context and then calls panic().
func Panicf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Panicf(format, args...)
}

// PanicKV writes a message and key-value pairs
// at the panic level using the logger from the context
// and then calls panic().
func PanicKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Panicw(message, kvs...)
}
