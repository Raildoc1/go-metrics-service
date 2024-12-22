package logging

import (
	"os"

	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"
)

func CreateZapLogger(development bool, f *os.File) *zap.Logger {
	var encoderConfig zapcore.EncoderConfig

	if development {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
	}

	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)

	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	level := zap.InfoLevel
	if development {
		level = zap.DebugLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(f), level),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
	)

	return zap.New(core)
}
