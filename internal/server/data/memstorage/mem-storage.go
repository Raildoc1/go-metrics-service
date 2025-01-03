package memstorage

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"go-metrics-service/internal/common/compression"
	"go-metrics-service/internal/server/data"
	"io"

	"go.uber.org/zap"
)

type MemStorage struct {
	data   rawData
	logger *zap.Logger
}

type rawData struct {
	Counters map[string]int64
	Gauges   map[string]float64
}

func NewMemStorage(logger *zap.Logger) *MemStorage {
	return &MemStorage{
		data: rawData{
			Gauges:   make(map[string]float64),
			Counters: make(map[string]int64),
		},
		logger: logger,
	}
}

func (s *MemStorage) SetCounter(key string, value int64) error {
	s.data.Counters[key] = value
	return nil
}

func (s *MemStorage) SetGauge(key string, value float64) error {
	s.data.Gauges[key] = value
	return nil
}

func (s *MemStorage) Has(key string) (bool, error) {
	_, hasCounter := s.data.Counters[key]
	_, hasGauge := s.data.Gauges[key]
	return hasCounter || hasGauge, nil
}

func (s *MemStorage) GetCounter(key string) (int64, error) {
	if _, ok := s.data.Gauges[key]; ok {
		return 0, data.ErrWrongType
	}
	if val, ok := s.data.Counters[key]; ok {
		return val, nil
	}
	return 0, data.ErrNotFound
}

func (s *MemStorage) GetGauge(key string) (float64, error) {
	if _, ok := s.data.Counters[key]; ok {
		return 0, data.ErrWrongType
	}
	if val, ok := s.data.Gauges[key]; ok {
		return val, nil
	}
	return 0, data.ErrNotFound
}

func (s *MemStorage) GetAll() (map[string]any, error) {
	res := make(map[string]any)
	for k, v := range s.data.Counters {
		res[k] = v
	}
	for k, v := range s.data.Gauges {
		res[k] = v
	}
	return res, nil
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
			Counters: readData.Counters,
			Gauges:   readData.Gauges,
		},
		logger: logger,
	}, nil
}

func (s *MemStorage) SaveTo(writer io.Writer) error {
	err := compression.GzipCompress(
		rawData{
			Counters: s.data.Counters,
			Gauges:   s.data.Gauges,
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
