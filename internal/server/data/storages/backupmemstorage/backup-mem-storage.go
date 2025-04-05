// Package backupmemstorage contains memstorage wrapper that can save its state to file and then restore it
package backupmemstorage

import (
	"errors"
	"fmt"
	"go-metrics-service/internal/server/data/storages/memstorage"
	"go-metrics-service/pkg/closehelpers"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	Backup      BackupConfig
	NeedRestore bool
}

type BackupConfig struct {
	FilePath      string
	StoreInterval time.Duration
}

type BackupMemStorage struct {
	*memstorage.MemStorage
	logger       *zap.Logger
	stopCh       chan struct{}
	syncCh       chan struct{}
	backupConfig BackupConfig
}

func New(cfg Config, logger *zap.Logger) (*BackupMemStorage, error) {
	if !cfg.NeedRestore {
		return newEmpty(cfg.Backup, logger), nil
	}
	file, err := os.Open(cfg.Backup.FilePath)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			return newEmpty(cfg.Backup, logger), nil
		default:
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
	}
	defer closehelpers.CloseWithErrorLogging(file, "file", logger)
	ms, err := memstorage.LoadFrom(file, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load from file: %w", err)
	}
	return create(cfg.Backup, ms, logger), nil
}

func (s *BackupMemStorage) Stop() {
	s.stopCh <- struct{}{}
	<-s.syncCh
}

func newEmpty(backupConfig BackupConfig, logger *zap.Logger) *BackupMemStorage {
	return create(backupConfig, memstorage.New(logger), logger)
}

func create(backupConfig BackupConfig, ms *memstorage.MemStorage, logger *zap.Logger) *BackupMemStorage {
	res := &BackupMemStorage{
		MemStorage:   ms,
		logger:       logger,
		backupConfig: backupConfig,
		stopCh:       make(chan struct{}, 1),
		syncCh:       make(chan struct{}, 1),
	}
	go res.savingProcess()
	return res
}

func (s *BackupMemStorage) savingProcess() {
	saveToFileTicker := time.NewTicker(s.backupConfig.StoreInterval)
	defer saveToFileTicker.Stop()
	for {
		select {
		case <-saveToFileTicker.C:
			s.saveToFileLogError(s.backupConfig.FilePath)
		case <-s.stopCh:
			s.saveToFileLogError(s.backupConfig.FilePath)
			s.syncCh <- struct{}{}
			return
		}
	}
}

func (s *BackupMemStorage) saveToFileLogError(filePath string) {
	err := s.saveToFile(filePath)
	if err != nil {
		s.logger.Error("failed to save to file", zap.String("filePath", filePath), zap.Error(err))
	}
}

func (s *BackupMemStorage) saveToFile(filePath string) error {
	dir := filepath.Dir(filePath)
	const dirPerm = 0o700
	err := os.MkdirAll(dir, dirPerm)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrExist):
			// Ignore.
		default:
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
			s.logger.Error("failed to close file", zap.Error(err))
		}
	}(file)
	if err := s.SaveTo(file); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	return nil
}
