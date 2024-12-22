package handlers

import (
	"io"

	"go.uber.org/zap"
)

func closeBody(body io.ReadCloser, logger *zap.Logger) {
	err := body.Close()
	if err != nil {
		logger.Error("failed to close body", zap.Error(err))
	}
}
