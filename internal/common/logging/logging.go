// Package logging
package logging

import (
	"os"

	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"
)

func CreateZapLogger(development bool) *zap.Logger {
	var encoderConfig zapcore.EncoderConfig

	if development {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
	}

	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	level := zap.InfoLevel
	if development {
		level = zap.DebugLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
	)

	return zap.New(core)
}
