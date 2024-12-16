package storage

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"go-metrics-service/internal/server/data"
	"os"
)

type MemStorage struct {
	data   map[string]any
	logger data.Logger
}

type serializableData struct {
	Data map[string]any `json:"data"`
}

func NewMemStorage(logger data.Logger) *MemStorage {
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
	serializedData, err := json.Marshal(serializableData{Data: m.data})
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			m.logger.Errorln(err)
		}
	}(file)

	gzipWriter, err := gzip.NewWriterLevel(file, gzip.BestCompression)
	defer func(gzipWriter *gzip.Writer) {
		err := gzipWriter.Close()
		if err != nil {
			m.logger.Errorln(err)
		}
	}(gzipWriter)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}

	_, err = gzipWriter.Write(serializedData)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
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
			m.logger.Errorln(err)
		}
	}(file)

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func(gzipReader *gzip.Reader) {
		err := gzipReader.Close()
		if err != nil {
			m.logger.Errorln(err)
		}
	}(gzipReader)

	var readData serializableData
	jsonDecoder := json.NewDecoder(gzipReader)
	err = jsonDecoder.Decode(&readData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	m.data = readData.Data
	return nil
}
