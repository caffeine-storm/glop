package aggregator

// Simple struct that aggregates presses and press_amts during a frame so they
// can be viewed between KeyThink()s
type keyStats struct {
	press_count   int
	release_count int
	press_amt     float64
	press_sum     float64 // TODO(#49): this is really a 'press_integral_w.r.t_time'
	press_avg     float64
}
