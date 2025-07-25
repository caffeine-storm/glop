package aggregator

import "fmt"

type baseAggregator struct {
	this, prev keyStats
}

func (a *baseAggregator) FramePressCount() int {
	return a.prev.press_count
}

func (a *baseAggregator) FrameReleaseCount() int {
	return a.prev.release_count
}

func (a *baseAggregator) FramePressAmt() float64 {
	return a.prev.press_amt
}

func (a *baseAggregator) FramePressSum() float64 {
	return a.prev.press_sum
}

func (a *baseAggregator) FramePressAvg() float64 {
	return a.prev.press_avg
}

func (a *baseAggregator) FramePressTotal() float64 {
	panic(fmt.Errorf("should not happen"))
}

func (a *baseAggregator) CurPressCount() int {
	return a.this.press_count
}

func (a *baseAggregator) CurReleaseCount() int {
	return a.this.release_count
}

func (a *baseAggregator) CurPressAmt() float64 {
	return a.this.press_amt
}

func (a *baseAggregator) CurPressSum() float64 {
	return a.this.press_sum
}

func (a *baseAggregator) updateCounts(event_type EventType) {
	switch event_type {
	case Press:
		a.this.press_count++
	case Release:
		a.this.release_count++
	}
}

func (a *baseAggregator) SendAllNonZero() bool {
	return false
}

// The standardAggregator's sum is an integral of the press_amt over time
type standardAggregator struct {
	baseAggregator
	last_press int64
	last_think int64
}

func (sa *standardAggregator) IsDown() bool {
	return sa.this.press_amt != 0
}

func (sa *standardAggregator) AggregatorSetPressAmt(amt float64, ms int64, event_type EventType) {
	sa.this.press_sum += sa.this.press_amt * float64(ms-sa.last_press)
	sa.this.press_amt = amt
	sa.last_press = ms
	sa.updateCounts(event_type)
}

func (sa *standardAggregator) AggregatorThink(ms int64) (bool, float64) {
	sa.this.press_sum += sa.this.press_amt * float64(ms-sa.last_press)
	if ms != sa.last_think {
		sa.this.press_avg = sa.this.press_sum / float64(ms-sa.last_think)
	} else {
		sa.this.press_avg = 0
	}
	sa.prev = sa.this
	sa.this = keyStats{
		press_amt: sa.prev.press_amt,
	}
	sa.last_press = ms
	sa.last_think = ms
	return false, 0
}

// The axisAggregator's sum is the sum of all press amounts specified by
// SetPressAmt(). FramePressAvg() returns the same value as FramePressSum().
type axisAggregator struct {
	baseAggregator
	is_down bool
}

func (*axisAggregator) SendAllNonZero() bool {
	return true
}

func (aa *axisAggregator) IsDown() bool {
	return aa.is_down
}

func (aa *axisAggregator) AggregatorSetPressAmt(amt float64, ms int64, event_type EventType) {
	aa.this.press_sum += amt
	aa.this.press_amt = amt
	if amt != 0 {
		aa.is_down = true
	}
	aa.updateCounts(event_type)
}

func (aa *axisAggregator) AggregatorThink(ms int64) (bool, float64) {
	was_down := aa.prev.press_amt != 0
	aa.prev = aa.this
	aa.this = keyStats{}
	aa.prev.press_avg = aa.prev.press_sum
	if aa.prev.press_amt == 0 {
		aa.is_down = false
		if was_down {
			return true, 0
		}
	}
	return false, 0
}

// A wheelAggregator is just like a standardAggregator except:
// - It sends Adjust events for *all* non-zero press amounts
// - If a frame goes by without it receiving any input it creates a Release //
// event
// - It implements TotalingAggregator so we can expose the raw sum instead of
// the integral that Aggregator.FramePressSum() returns
type wheelAggregator struct {
	standardAggregator
	this_total, cur_total float64
}

func (wa *wheelAggregator) SendAllNonZero() bool {
	return true
}

func (wa *wheelAggregator) AggregatorSetPressAmt(amt float64, ms int64, event_type EventType) {
	wa.standardAggregator.AggregatorSetPressAmt(amt, ms, event_type)
	wa.cur_total += amt
}

func (wa *wheelAggregator) FramePressTotal() float64 {
	return wa.this_total
}

func (wa *wheelAggregator) AggregatorThink(ms int64) (bool, float64) {
	if b, _ := wa.standardAggregator.AggregatorThink(ms); b {
		panic("standardAggregator should not generate an event on AggregatorThink()")
	}

	wa.this_total = wa.cur_total
	wa.cur_total = 0

	// Note: 'CurPressAmt' here should be read as "press amount as-of end of last
	// frame" because we called standardAggregator.AggregatorThink above.
	if wa.CurPressAmt() != 0 {
		return true, 0
	}
	return false, 0
}
