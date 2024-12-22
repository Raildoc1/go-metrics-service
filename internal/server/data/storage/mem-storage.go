package storage

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"go-metrics-service/internal/common/compression"
	"io"

	"go.uber.org/zap"
)

type MemStorage struct {
	data   map[string]any
	logger *zap.Logger
}

type serializableData struct {
	Data map[string]any
}

func NewMemStorage(logger *zap.Logger) *MemStorage {
	return &MemStorage{
		data:   make(map[string]any),
		logger: logger,
	}
}

func LoadFrom(reader io.Reader, logger *zap.Logger) (*MemStorage, error) {
	var readData serializableData
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
		data:   readData.Data,
		logger: logger,
	}, nil
}

func (m *MemStorage) Set(key string, value any) {
	m.data[key] = value
}

func (m *MemStorage) Has(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (m *MemStorage) Get(key string) (any, bool) {
	v, ok := m.data[key]
	return v, ok
}

func (m *MemStorage) GetAll() map[string]any {
	return m.data
}

func (m *MemStorage) SaveTo(writer io.Writer) error {
	err := compression.GzipCompress(
		serializableData{Data: m.data},
		func(writer io.Writer) compression.Encoder {
			return gob.NewEncoder(writer)
		},
		writer,
		gzip.BestCompression,
		m.logger,
	)
	if err != nil {
		return fmt.Errorf("failed not compress data: %w", err)
	}
	return nil
}
