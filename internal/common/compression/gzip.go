package compression

import (
	"compress/gzip"
	"fmt"
	"io"

	"go.uber.org/zap"
)

type Encoder interface {
	Encode(v any) error
}

type Decoder interface {
	Decode(v any) error
}

func GzipCompress(
	item any,
	newEncoder func(writer io.Writer) Encoder,
	target io.Writer,
	level int,
	logger *zap.Logger,
) error {
	gzipWriter, err := gzip.NewWriterLevel(target, level)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}
	defer func(gzipWriter *gzip.Writer) {
		err := gzipWriter.Close()
		if err != nil {
			logger.Error("failed to close gzip writer", zap.Error(err))
		}
	}(gzipWriter)

	encoder := newEncoder(gzipWriter)
	err = encoder.Encode(item)
	if err != nil {
		return fmt.Errorf("failed to encode item: %w", err)
	}

	err = gzipWriter.Close()
	if err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}
	return nil
}

func GzipDecompress(
	item any,
	newDecoder func(reader io.Reader) Decoder,
	reader io.Reader,
	logger *zap.Logger,
) error {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func(gzipReader *gzip.Reader) {
		err := gzipReader.Close()
		if err != nil {
			logger.Error("failed to close gzip reader", zap.Error(err))
		}
	}(gzipReader)

	decoder := newDecoder(gzipReader)
	err = decoder.Decode(item)
	if err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}
	return nil
}
