package memstorage

import (
	"compress/gzip"
	"encoding/gob"
	"errors"
	"fmt"
	"go-metrics-service/internal/common/compression"
	"go-metrics-service/internal/server/data"
	"io"
	"sync"

	"github.com/google/uuid"

	"go.uber.org/zap"
)

type MemStorage struct {
	data             rawData
	transaction      *transaction
	logger           *zap.Logger
	transactionMutex sync.Mutex
}

type rawData struct {
	Counters map[string]int64
	Gauges   map[string]float64
}

type transaction struct {
	countersToSet map[string]int64
	gaugesToSet   map[string]float64
	id            data.TransactionID
}

func NewMemStorage(logger *zap.Logger) *MemStorage {
	return &MemStorage{
		transaction: nil,
		data: rawData{
			Gauges:   make(map[string]float64),
			Counters: make(map[string]int64),
		},
		logger: logger,
	}
}

func (s *MemStorage) BeginTransaction() (data.TransactionID, error) {
	s.transactionMutex.Lock()
	if s.transaction != nil {
		return "", errors.New("transaction is already opened")
	}
	id := uuid.New()
	transactionID := data.TransactionID(id.String())
	s.transaction = &transaction{
		id:            transactionID,
		countersToSet: make(map[string]int64),
		gaugesToSet:   make(map[string]float64),
	}
	return transactionID, nil
}

func (s *MemStorage) CommitTransaction(transactionID data.TransactionID) error {
	if s.transaction == nil {
		return data.ErrNoTransactionOpened
	}
	if s.transaction.id != transactionID {
		return data.ErrWrongTransactionID
	}
	for k, v := range s.transaction.countersToSet {
		s.data.Counters[k] = v
	}
	for k, v := range s.transaction.gaugesToSet {
		s.data.Gauges[k] = v
	}
	s.transaction = nil
	s.transactionMutex.Unlock()
	return nil
}

func (s *MemStorage) RollbackTransaction(transactionID data.TransactionID) error {
	if s.transaction == nil {
		return data.ErrNoTransactionOpened
	}
	if s.transaction.id != transactionID {
		return data.ErrWrongTransactionID
	}
	s.transaction = nil
	s.transactionMutex.Unlock()
	return nil
}

func (s *MemStorage) SetCounter(key string, value int64, transactionID data.TransactionID) error {
	if err := s.validateTransactionID(transactionID); err != nil {
		return err
	}
	if _, ok := s.data.Gauges[key]; ok {
		return data.ErrWrongType
	}
	if _, ok := s.transaction.gaugesToSet[key]; ok {
		return data.ErrWrongType
	}
	s.transaction.countersToSet[key] = value
	return nil
}

func (s *MemStorage) SetGauge(key string, value float64, transactionID data.TransactionID) error {
	if err := s.validateTransactionID(transactionID); err != nil {
		return err
	}
	if _, ok := s.data.Counters[key]; ok {
		return data.ErrWrongType
	}
	if _, ok := s.transaction.countersToSet[key]; ok {
		return data.ErrWrongType
	}
	s.transaction.gaugesToSet[key] = value
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

func (s *MemStorage) validateTransactionID(transactionID data.TransactionID) error {
	if s.transaction == nil || s.transaction.id != transactionID {
		return data.ErrWrongTransactionID
	}
	return nil
}
