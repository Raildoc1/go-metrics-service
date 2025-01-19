package memstorage

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"go-metrics-service/internal/common/compression"
	"io"

	"go.uber.org/zap"
)

type MemStorage struct {
	data   rawData
	logger *zap.Logger
}

type rawData struct {
	Values map[string]any
}

func New(logger *zap.Logger) *MemStorage {
	return &MemStorage{
		data: rawData{
			Values: make(map[string]any),
		},
		logger: logger,
	}
}

func LoadFrom(reader io.Reader, logger *zap.Logger) (*MemStorage, error) {
	var readData rawData
	err := compression.GzipDecompress(
		&readData,
		func(reader io.Reader) compression.Decoder {
			return gob.NewDecoder(reader)
		},
		reader,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}
	return &MemStorage{
		data: rawData{
			Values: readData.Values,
		},
		logger: logger,
	}, nil
}

func (s *MemStorage) SaveTo(writer io.Writer) error {
	err := compression.GzipCompress(
		rawData{
			Values: s.data.Values,
		},
		func(writer io.Writer) compression.Encoder {
			return gob.NewEncoder(writer)
		},
		writer,
		gzip.BestCompression,
		s.logger,
	)
	if err != nil {
		return fmt.Errorf("failed not compress data: %w", err)
	}
	return nil
}

func (s *MemStorage) Get(key string) (val any, ok bool) {
	val, ok = s.data.Values[key]
	return
}
func (s *MemStorage) GetAll() map[string]any {
	return s.data.Values
}
func (s *MemStorage) Set(key string, value any) {
	s.data.Values[key] = value
}
