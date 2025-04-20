// Package gohelpers contains helpers for common used goroutine patterns
// like starting goroutine with "done" channel
package gohelpers

import (
	"context"
	"time"
)

// StartTickerProcess starts goroutine that runs 'f' with interval
// goroutine interrupts when receive anything from 'doneCh' channel, or it's closed
func StartTickerProcess(doneCh <-chan struct{}, f func(context.Context) error, interval time.Duration) chan error {
	ticker := time.NewTicker(interval)
	return StartProcess[time.Time](
		doneCh,
		func(ctx context.Context, _ time.Time) error { return f(ctx) },
		func() { ticker.Stop() },
		ticker.C,
	)
}

// StartProcess starts goroutine that runs 'f' with argument received from 'in' channel
// goroutine interrupts when receive anything from 'doneCh' channel, or it's closed
func StartProcess[T any](
	doneCh <-chan struct{},
	f func(ctx context.Context, arg T) error,
	afterStop func(),
	in <-chan T,
) chan error {
	errCh := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer close(errCh)
		defer afterStop()

		for {
			select {
			case v, ok := <-in:
				if !ok {
					return
				}
				err := f(ctx, v)
				if err != nil {
					errCh <- err
				}
			case <-doneCh:
				cancel()
				return
			}
		}
	}()

	return errCh
}
