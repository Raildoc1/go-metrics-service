// Package compression contains gzip compression helpers
package compression

import (
	"compress/gzip"
	"fmt"
	"go-metrics-service/pkg/closehelpers"
	"io"

	"go.uber.org/zap"
)

type Encoder interface {
	Encode(v any) error
}

type Decoder interface {
	Decode(v any) error
}

// GzipCompress encodes item with encoder created with newEncoder function
// then compresses with gzip of provided level
// then writes to writer
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
	defer closehelpers.CloseWithErrorLogging(gzipWriter, "gzip writer", logger)

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

// GzipDecompress decompress input from reader
// then decodes it to item
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
	defer closehelpers.CloseWithErrorLogging(gzipReader, "gzip reader", logger)

	decoder := newDecoder(gzipReader)
	err = decoder.Decode(item)
	if err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}
	return nil
}
