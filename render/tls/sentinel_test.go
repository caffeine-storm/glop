package tls_test

import (
	"runtime"
	"testing"

	"github.com/runningwild/glop/render/tls"
)

func inFreshThread(fn func()) {
	go func() {
		runtime.LockOSThread()
		tls.ClearSentinel()
		fn()
	}()
}

func blockingInFreshThread(fn func()) {
	done := make(chan bool)
	go func() {
		defer func() {
			done <- true
		}()

		runtime.LockOSThread()
		tls.ClearSentinel()
		fn()
	}()
	<-done
}

func TestSentinel(t *testing.T) {
	t.Run("a new sentinel is not set", func(t *testing.T) {
		blockingInFreshThread(func() {
			if tls.IsSentinelSet() {
				t.Fatalf("a new sentinel must not be 'set'")
			}
		})
	})
	t.Run("can set a sentinel", func(t *testing.T) {
		blockingInFreshThread(func() {
			tls.SetSentinel()

			if !tls.IsSentinelSet() {
				t.Fatalf("a 'set' sentinel must be 'set'")
			}
		})
	})
	t.Run("can clear a sentinel", func(t *testing.T) {
		blockingInFreshThread(func() {
			tls.SetSentinel()
			tls.ClearSentinel()

			if tls.IsSentinelSet() {
				t.Fatalf("a 'cleared' sentinel must not be 'set'")
			}
		})
	})
	t.Run("a sentinel's 'setted-ness' is thread-specific", func(t *testing.T) {
		failchan := make(chan string, 24)
		stepper := make(chan bool)

		inFreshThread(func() {
			if tls.IsSentinelSet() {
				failchan <- "gr1: a new sentinel must not be 'set'"
				return
			}

			// 1: Wait for "set the sentinel" event
			<-stepper
			tls.SetSentinel()
			// Report that we're done setting
			stepper <- true

			if !tls.IsSentinelSet() {
				failchan <- "gr1: a 'set' sentinel must be 'set'"
				return
			}

			// 2: Wait for "check the sentinel again" event
			<-stepper
			if !tls.IsSentinelSet() {
				failchan <- "gr1: a 'set' sentinel must remain 'set'"
				return
			}
			// Report that we're done checking
			stepper <- true

			// 3: Wait for "check the sentinel again... again!" event
			<-stepper
			if !tls.IsSentinelSet() {
				failchan <- "gr1: a 'set' sentinel must still remain 'set'"
				return
			}
			// Report that we're done checking again
			stepper <- true

			// 4: Wait for "clear the sentinel" event
			<-stepper
			tls.ClearSentinel()
			// Report that we're done clearing
			stepper <- true

			if tls.IsSentinelSet() {
				failchan <- "gr1: a 'cleared' sentinel must not be 'set'"
				return
			}

			// 5: Wait for "check the sentinel" event
			<-stepper
			if tls.IsSentinelSet() {
				failchan <- "gr1: a 'cleared' sentinel must remain 'clear'"
				return
			}
			// Report that we're done checking
			stepper <- true

			// Report finished checks
			stepper <- true
		})

		inFreshThread(func() {
			defer close(failchan)

			// Check our thread's sentinel; it should be fresh
			if tls.IsSentinelSet() {
				failchan <- "gr2: sentinel set before interacting"
				return
			}

			// 1: Tell the other GR to set its sentinel
			stepper <- true
			<-stepper

			// Our sentinel should not have changed
			if tls.IsSentinelSet() {
				failchan <- "gr2: sentinel set by other gr!"
				return
			}

			tls.SetSentinel()

			if !tls.IsSentinelSet() {
				failchan <- "gr2: couldn't set own sentinel"
				return
			}

			// 2: Tell other GR to check that we didn't clobber _it_
			stepper <- true
			<-stepper

			// Clear our sentinel
			tls.ClearSentinel()

			if tls.IsSentinelSet() {
				failchan <- "gr2: couldn't clear"
				return
			}

			// 3: Tell other GR to check again that we didn't clobber _it_
			stepper <- true
			<-stepper

			if tls.IsSentinelSet() {
				failchan <- "gr2: cross-talk set it"
				return
			}

			// 4: Tell other GR to clear theirs
			stepper <- true
			<-stepper

			// We should still be clear
			if tls.IsSentinelSet() {
				failchan <- "gr2: cross-talk set it while requesting a clear!?"
				return
			}

			tls.SetSentinel()

			if !tls.IsSentinelSet() {
				failchan <- "gr2: couldn't set our own sentinel T_T"
				return
			}

			// 5: tell other GR to check again
			stepper <- true
			<-stepper

			if !tls.IsSentinelSet() {
				failchan <- "gr2: other gr checking set our sentinel!?"
				return
			}

			// sync with other GR
			<-stepper
		})

		for msg := range failchan {
			t.Logf("fail: %s", msg)
			t.Fail()
		}

		close(stepper)
	})
}
