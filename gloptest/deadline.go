package gloptest

import (
	"fmt"
	"time"
)

func RunWithDeadline(deadline time.Duration, op func()) error {
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
		return nil
	case err := <-errchan:
		return err
	case <-time.After(deadline):
		return fmt.Errorf("deadline (%s) exceeded", deadline)
	}
}
