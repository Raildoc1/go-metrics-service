package logger

import (
	"fmt"

	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

func CreateLogger(development bool) error {
	var lgr *zap.Logger
	var err error

	if development {
		lgr, err = zap.NewDevelopment()
	} else {
		lgr, err = zap.NewProduction()
	}

	if err != nil {
		return fmt.Errorf("logger init: %w", err)
	}

	Log = lgr.Sugar()
	return nil
}

func Sync() error {
	err := Log.Sync()
	if err != nil {
		return fmt.Errorf("logger sync: %w", err)
	}
	return nil
}
