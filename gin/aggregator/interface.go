package aggregator

// A SubAggregator is the intersection (blech!) between the Key interface and
// the aggregator interface. It's really two APIs in one; the "Frame*"
// functions for consumption external to the gin package and the "Cur*"
// functions for querying running totals during frame event processing.
type SubAggregator interface {
	IsDown() bool
	FramePressCount() int
	FrameReleaseCount() int
	FramePressAmt() float64

	// TODO(#49): FramePressSum is really FramePressIntegral
	FramePressSum() float64
	FramePressAvg() float64
	FramePressTotal() float64
	CurPressCount() int
	CurReleaseCount() int
	CurPressAmt() float64
	CurPressSum() float64
}

type Aggregator interface {
	SubAggregator
	AggregatorThink(ms int64) (bool, float64)
	AggregatorSetPressAmt(amt float64, ms int64, event_type EventType)

	// A very select set of keys should always send events when their press amt
	// is non-zero. These are typically not your ordinary keys, mouse wheels,
	// mouse pointers, etc...
	SendAllNonZero() bool
}

type TotalingAggregator interface {
	Aggregator

	// TODO(#49): Maybe this should be 'sum' once we don't use 'sum' to mean
	// 'integral'?
	FramePressTotal() float64
}
