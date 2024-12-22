package handlers

import (
	"io"

	"go.uber.org/zap"
)

func closeBody(Body io.ReadCloser, logger *zap.Logger) {
	err := Body.Close()
	if err != nil {
		logger.Error("failed to close body", zap.Error(err))
	}
}
