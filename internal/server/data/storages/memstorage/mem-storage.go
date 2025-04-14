// Package memstorage contains RAM storage implementation
package memstorage

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"go-metrics-service/pkg/compression"
	"io"
	"sync"

	"go.uber.org/zap"
)

type MemStorage struct {
	data   rawData
	mux    *sync.Mutex
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
		mux:    &sync.Mutex{},
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
		mux:    &sync.Mutex{},
		logger: logger,
	}, nil
}

func (s *MemStorage) SaveTo(writer io.Writer) error {
	s.mux.Lock()
	defer s.mux.Unlock()
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
	s.mux.Lock()
	defer s.mux.Unlock()
	val, ok = s.data.Values[key]
	return
}
func (s *MemStorage) GetAll() map[string]any {
	s.mux.Lock()
	defer s.mux.Unlock()
	dataCopy := make(map[string]any)
	for k, v := range s.data.Values {
		dataCopy[k] = v
	}
	return dataCopy
}
func (s *MemStorage) Set(key string, value any) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.data.Values[key] = value
}
