package gloptest

import (
	"fmt"
	"time"
)

// Returns two error values. If the first error value is non-nil, the operation
// timedout; the first error value expresses that. If the second error value is
// non-nil, the operation itself failed; the second error value is the error
// that ocurred during the operation. If both error values are nil, the
// operation succeeded within the deadline.
func RunWithDeadline(deadline time.Duration, op func()) (error, error) {
	completed := make(chan bool)
	errchan := make(chan error)
	go func() {
		defer func() {
			// If 'op' panics, return the error value it paniced on.
			if err := recover(); err != nil {
				errchan <- err.(error)
			}
		}()
		op()
		completed <- true
	}()

	select {
	case <-completed:
		return nil, nil
	case err := <-errchan:
		return nil, err
	case <-time.After(deadline):
		return fmt.Errorf("deadline (%s) exceeded", deadline), nil
	}
}
