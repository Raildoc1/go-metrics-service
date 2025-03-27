package gohelpers

import "sync"

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
