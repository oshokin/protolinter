package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type coreWithLevel struct {
	zapcore.Core
	level zapcore.Level
}

// Enabled возвращает true, если предоставленный уровень включен для логирования ядром.
// Он вызывает метод Enabled у обернутого zapcore.Level.
func (c *coreWithLevel) Enabled(l zapcore.Level) bool {
	return c.level.Enabled(l)
}

// Check добавляет ядро к проверенной записи, если уровень записи включен для логирования.
// Он возвращает проверенную запись с добавленным ядром или исходную проверенную запись,
// если уровень отключен.
func (c *coreWithLevel) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}

	return ce
}

// With возвращает новое ядро с добавленными полями к обернутому ядру.
// Он возвращает новый coreWithLevel с тем же уровнем, что и у исходного ядра.
func (c *coreWithLevel) With(fields []zapcore.Field) zapcore.Core {
	return &coreWithLevel{
		c.Core.With(fields),
		c.level,
	}
}

// WithLevel представляет собой опцию, которая создает логгер
// с указанным уровнем логирования на основе существующего логгера.
// Он возвращает zap.Option, который оборачивает существующее ядро
// в coreWithLevel с указанным уровнем.
func WithLevel(lvl zapcore.Level) zap.Option {
	return zap.WrapCore(
		func(core zapcore.Core) zapcore.Core {
			return &coreWithLevel{core, lvl}
		})
}
