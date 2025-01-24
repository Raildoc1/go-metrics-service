package storages

import (
	"context"
)

type DummyTransactionsManager struct{}

func NewDummyTransactionsManager() *DummyTransactionsManager {
	return &DummyTransactionsManager{}
}

func (tm *DummyTransactionsManager) DoWithTransaction(
	ctx context.Context,
	f func(ctx context.Context) error,
) error {
	return f(ctx)
}
