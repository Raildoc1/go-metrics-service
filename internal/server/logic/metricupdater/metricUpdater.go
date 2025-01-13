package metricupdater

import (
	"context"
	"errors"
	"fmt"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data"
)

var (
	ErrNonExistentType = errors.New("non-existent type")
	ErrWrongValueType  = errors.New("wrong value type")
)

type TransactionsHandler interface {
	BeginTransaction() (data.TransactionID, error)
	CommitTransaction(data.TransactionID) error
	RollbackTransaction(transactionID data.TransactionID) error
}

type GaugeSetter interface {
	Set(ctx context.Context, key string, value float64, transactionID data.TransactionID) error
}

type CounterChanger interface {
	Change(ctx context.Context, key string, delta int64, transactionID data.TransactionID) error
}

type MetricUpdater struct {
	gaugeSetter        GaugeSetter
	counterChanger     CounterChanger
	transactionFactory TransactionsHandler
}

func New(
	transactionFactory TransactionsHandler,
	gaugeSetter GaugeSetter,
	counterChanger CounterChanger,
) *MetricUpdater {
	return &MetricUpdater{
		gaugeSetter:        gaugeSetter,
		counterChanger:     counterChanger,
		transactionFactory: transactionFactory,
	}
}

func (m *MetricUpdater) UpdateOne(ctx context.Context, metric protocol.Metrics) error {
	transactionID, err := m.transactionFactory.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	if err := m.updateInternal(ctx, metric, transactionID); err != nil {
		rollbackErr := m.transactionFactory.RollbackTransaction(transactionID)
		if rollbackErr != nil {
			return fmt.Errorf("failed to update metric: %w, %w", err, rollbackErr)
		}
		return fmt.Errorf("failed to update metric: %w", err)
	}
	if err := m.transactionFactory.CommitTransaction(transactionID); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (m *MetricUpdater) UpdateMany(ctx context.Context, metrics []protocol.Metrics) error {
	transactionID, err := m.transactionFactory.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	for _, metric := range metrics {
		if err := m.updateInternal(ctx, metric, transactionID); err != nil {
			rollbackErr := m.transactionFactory.RollbackTransaction(transactionID)
			if rollbackErr != nil {
				return fmt.Errorf("failed to update metric: %w, %w", err, rollbackErr)
			}
			return fmt.Errorf("failed to update metric: %w", err)
		}
	}
	if err := m.transactionFactory.CommitTransaction(transactionID); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (m *MetricUpdater) updateInternal(
	ctx context.Context,
	metric protocol.Metrics,
	transactionID data.TransactionID,
) error {
	switch metric.MType {
	case protocol.Gauge:
		if metric.Value == nil {
			return ErrWrongValueType
		}
		if err := m.gaugeSetter.Set(ctx, metric.ID, *metric.Value, transactionID); err != nil {
			return fmt.Errorf("set gauge: %w", err)
		}
	case protocol.Counter:
		if metric.Delta == nil {
			return ErrWrongValueType
		}
		if err := m.counterChanger.Change(ctx, metric.ID, *metric.Delta, transactionID); err != nil {
			return fmt.Errorf("change counter: %w", err)
		}
	default:
		return ErrNonExistentType
	}
	return nil
}
