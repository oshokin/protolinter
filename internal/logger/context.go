package logger

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const loggerContextKey contextKey = "logger"

// ToContext создает контекст с переданным логгером внутри.
func ToContext(ctx context.Context, l *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerContextKey, l)
}

// FromContext достает логгер из контекста. Если в контексте логгер не
// обнаруживается - возвращает глобальный логгер.
func FromContext(ctx context.Context) *zap.SugaredLogger {
	return getLogger(ctx)
}

// WithName создает именованный логгер из уже имеющегося в контексте.
// Дочерние логгеры будут наследовать имена (см пример).
func WithName(ctx context.Context, name string) context.Context {
	log := FromContext(ctx).Named(name)
	return ToContext(ctx, log)
}

// WithKV создает логгер из уже имеющегося в контексте и устанавливает метаданные.
// Принимает ключ и значение, которые будут наследоваться дочерними логгерами.
func WithKV(ctx context.Context, key string, value interface{}) context.Context {
	log := FromContext(ctx).With(key, value)
	return ToContext(ctx, log)
}

// WithFields создает логгер из уже имеющегося в контексте и устанавливает метаданные,
// используя типизированные поля.
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	log := FromContext(ctx).
		Desugar().
		With(fields...).
		Sugar()

	return ToContext(ctx, log)
}

func getLogger(ctx context.Context) *zap.SugaredLogger {
	l := global
	if logger, ok := ctx.Value(loggerContextKey).(*zap.SugaredLogger); ok {
		l = logger
	}

	return l
}
