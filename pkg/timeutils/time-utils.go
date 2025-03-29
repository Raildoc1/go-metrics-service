// Package timeutils contains time helpers
package timeutils

import (
	"context"
	"fmt"
	"time"
)

// Retry executes function until one of the following statements are true:
//  1. function returned no error
//  2. onFailed return false
//  3. retries max count reached
//
// retries max count equals to attemptsDelays length
//
// can be interrupted with context
func Retry(
	ctx context.Context,
	attemptDelays []time.Duration,
	function func(context.Context) error,
	onFailed func(error) (needRetry bool),
) error {
	var err error
	for _, delay := range attemptDelays {
		if ctx.Err() != nil {
			return fmt.Errorf("retry canceled: %w", ctx.Err())
		}
		err = function(ctx)
		if err == nil {
			return nil
		}
		if !onFailed(err) {
			return err
		}
		err := SleepCtx(ctx, delay)
		if err != nil {
			return err
		}
	}
	return err
}

// SleepCtx performs sleep that can be interrupted with context
func SleepCtx(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("sleep canceled: %w", ctx.Err())
	case <-time.After(d):
		return nil
	}
}
