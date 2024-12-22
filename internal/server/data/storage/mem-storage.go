package storage

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"

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

func (m *MemStorage) SaveToFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			m.logger.Error("failed to close file", zap.Error(err))
		}
	}(file)

	gzipWriter, err := gzip.NewWriterLevel(file, gzip.BestCompression)
	defer func(gzipWriter *gzip.Writer) {
		err := gzipWriter.Close()
		if err != nil {
			m.logger.Error("failed to close gzip writer", zap.Error(err))
		}
	}(gzipWriter)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}

	err = gob.NewEncoder(gzipWriter).Encode(serializableData{Data: m.data})
	if err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}
	return nil
}

func (m *MemStorage) LoadFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			m.logger.Error("failed to close file", zap.Error(err))
		}
	}(file)

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func(gzipReader *gzip.Reader) {
		err := gzipReader.Close()
		if err != nil {
			m.logger.Error("failed to close gzip reader", zap.Error(err))
		}
	}(gzipReader)

	var readData serializableData
	gobDecoder := gob.NewDecoder(gzipReader)
	err = gobDecoder.Decode(&readData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	m.data = readData.Data
	return nil
}
