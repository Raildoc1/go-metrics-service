package logging

import (
	"fmt"

	"go.uber.org/zap"
)

func CreateLogger(development bool) (*zap.SugaredLogger, error) {
	var lgr *zap.Logger
	var err error

	if development {
		lgr, err = zap.NewDevelopment()
	} else {
		lgr, err = zap.NewProduction()
	}

	if err != nil {
		return nil, fmt.Errorf("logger init: %w", err)
	}

	return lgr.Sugar(), nil
}
