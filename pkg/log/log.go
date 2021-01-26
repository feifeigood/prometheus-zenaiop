package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	lg      *zap.Logger
	encoder zapcore.Encoder
)

// Options for the creation of a LOG object.
type Options struct {
	Level string
}

// Init initialize LOG object at once.
func Init(opts Options) (err error) {
	lvl := new(zapcore.Level)
	if err := lvl.UnmarshalText([]byte(opts.Level)); err != nil {
		return err
	}

	encodercfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		NameKey:        "logger",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 短路径编码器(相对路径+行号)
		EncodeName:     zapcore.FullNameEncoder,
	}

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(*lvl),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "console",
		EncoderConfig:    encodercfg,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if lg, err = cfg.Build(zap.AddCaller()); err != nil {
		return err
	}

	// replaced global logger instance
	zap.ReplaceGlobals(lg)

	return
}
