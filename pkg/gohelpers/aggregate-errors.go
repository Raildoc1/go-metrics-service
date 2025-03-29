package gohelpers

import "sync"

// AggregateErrors passes errors from many error channels to one
func AggregateErrors(errChs ...chan error) chan error {
	resultCh := make(chan error)

	go func() {
		defer close(resultCh)
		wg := &sync.WaitGroup{}
		for _, errCh := range errChs {
			wg.Add(1)
			errCh := errCh
			go func() {
				defer wg.Done()
				for err := range errCh {
					resultCh <- err
				}
			}()
		}
		wg.Wait()
	}()

	return resultCh
}
