package aggregator

import "fmt"

// Simple struct that aggregates presses and press_amts during a frame so they
// can be viewed between KeyThink()s
type keyStats struct {
	press_count   int
	release_count int
	press_amt     float64
	press_sum     float64 // TODO(#49): this is really a 'press_integral_w.r.t_time'
	press_avg     float64
}

func (ks *keyStats) String() string {
	return fmt.Sprintf("%T{press_count: %v, release_count: %v, press_amt: %v, press_sum: %v, press_avg: %v}", ks, ks.press_count, ks.release_count, ks.press_amt, ks.press_sum, ks.press_avg)
}
