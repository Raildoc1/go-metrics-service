package compression

import (
	"compress/gzip"
	"fmt"
	"io"

	"go.uber.org/zap"
)

func GzipCompress(source io.Reader, target io.Writer, logger *zap.Logger) error {
	gzipWriter, err := gzip.NewWriterLevel(target, gzip.BestSpeed)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}
	defer func(gzipWriter *gzip.Writer) {
		err := gzipWriter.Close()
		if err != nil {
			logger.Error("failed to close gzip writer", zap.Error(err))
		}
	}(gzipWriter)

	_, err = io.Copy(gzipWriter, source)
	if err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}
	return nil
}
