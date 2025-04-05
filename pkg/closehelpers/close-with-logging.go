package closehelpers

import (
	"io"

	"go.uber.org/zap"
)

func CloseWithErrorLogging(closer io.Closer, entityName string, logger *zap.Logger) {
	err := closer.Close()
	if err != nil {
		logger.Error("failed to close", zap.String("entityName", entityName), zap.Error(err))
	}
}
