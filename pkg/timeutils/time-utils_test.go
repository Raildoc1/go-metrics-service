package timeutils

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func ExampleRetry() {
	attemptCount := 0

	err := Retry(
		context.Background(),
		[]time.Duration{
			1 * time.Second,
			2 * time.Second,
			3 * time.Second,
		},
		func(ctx context.Context) error {
			attemptCount++
			if attemptCount > 2 {
				fmt.Printf("attempt #%v successed\n", attemptCount)
				return nil
			} else {
				fmt.Printf("attempt #%v failed\n", attemptCount)
				return errors.New("test error")
			}
		},
		func(err error) (needRetry bool) {
			if err != nil {
				fmt.Printf("error occured: %s\n", err)
			}
			return err != nil
		},
	)

	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// attempt #1 failed
	// error occured: test error
	// attempt #2 failed
	// error occured: test error
	// attempt #3 successed
}

func ExampleSleepCtx() {
	syncCh := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		err := SleepCtx(ctx, 10*time.Second)
		fmt.Println(err)
		syncCh <- struct{}{}
	}(ctx)

	time.Sleep(1 * time.Second)
	cancel()

	<-syncCh

	// Output:
	// sleep canceled: context canceled
}
