package storage

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"go-metrics-service/internal/common/compression"
	"io"
	"os"
	"path/filepath"

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

func (s *MemStorage) Set(key string, value any) {
	s.data[key] = value
}

func (s *MemStorage) Has(key string) bool {
	_, ok := s.data[key]
	return ok
}

func (s *MemStorage) Get(key string) (any, bool) {
	v, ok := s.data[key]
	return v, ok
}

func (s *MemStorage) GetAll() map[string]any {
	return s.data
}

func (s *MemStorage) SaveTo(writer io.Writer) error {
	err := compression.GzipCompress(
		serializableData{Data: s.data},
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

func SaveMemStorageToFile(memStorage *MemStorage, filePath string, logger *zap.Logger) error {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		const dirPerm = 0o700
		err = os.MkdirAll(dir, dirPerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Error("failed to close file", zap.Error(err))
		}
	}(file)
	if err := memStorage.SaveTo(file); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	return nil
}
