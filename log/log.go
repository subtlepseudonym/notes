// Deprecated: log has been replaced by an in-place definition in the main
// package and should not be used. It will be removed is notes/v2
package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(out *os.File, level int) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderConfig)
	allLevels := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return zapcore.Level(int8(level)).Enabled(lvl)
	})

	core := zapcore.NewCore(encoder, out, allLevels)
	return zap.New(core)
}
