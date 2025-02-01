package gohelpers

import "time"

func StartTickerProcess(doneCh <-chan struct{}, f func() error, interval time.Duration) chan error {
	ticker := time.NewTicker(interval)
	return StartProcess[time.Time](
		doneCh,
		func(_ time.Time) error { return f() },
		func() { ticker.Stop() },
		ticker.C,
	)
}

func StartProcess[T any](doneCh <-chan struct{}, f func(T) error, afterStop func(), in <-chan T) chan error {
	errCh := make(chan error)

	go func() {
		defer close(errCh)
		defer afterStop()

		for {
			select {
			case v, ok := <-in:
				if !ok {
					return
				}
				err := f(v)
				if err != nil {
					errCh <- err
				}
			case <-doneCh:
				return
			}
		}
	}()

	return errCh
}
