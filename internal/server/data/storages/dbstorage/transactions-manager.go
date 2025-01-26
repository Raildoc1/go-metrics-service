package dbstorage

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type TransactionsManager struct {
	storage *DBStorage
	logger  *zap.Logger
}

func NewTransactionsManager(storage *DBStorage, logger *zap.Logger) *TransactionsManager {
	return &TransactionsManager{
		storage: storage,
		logger:  logger,
	}
}

func (tm *TransactionsManager) DoWithTransaction(
	ctx context.Context,
	f func(ctx context.Context) error,
) error {
	ctxWithTransaction, tx, err := tm.storage.WithTransaction(ctx)
	if err != nil {
		return err
	}
	err = f(ctxWithTransaction)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			tm.logger.Error("failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}
	return nil
}
