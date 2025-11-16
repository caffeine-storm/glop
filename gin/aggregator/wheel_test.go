package aggregator_test

import (
	"fmt"
	"testing"

	"github.com/caffeine-storm/glop/gin/aggregator"
)

func TestWheelAggregator(t *testing.T) {
	t.Run("can handle scroll-start and scroll-end in single frame", func(t *testing.T) {
		aggBaseType := aggregator.AggregatorForType(aggregator.AggregatorTypeWheel)
		agg, ok := aggBaseType.(aggregator.TotalingAggregator)
		if !ok {
			panic(fmt.Errorf("a wheel aggregator should implement TotalingAggregator"))
		}

		// Start scrolling up
		agg.AggregatorSetPressAmt(1, 42, aggregator.Press)
		// Stop scrolling up
		agg.AggregatorSetPressAmt(0, 42, aggregator.Release)

		if agg.CurPressTotal() == 0 {
			t.Fatalf("a zero sum is incorrect because we're pressing it")
		}
	})

	t.Run("can scroll down", func(t *testing.T) {
		aggBaseType := aggregator.AggregatorForType(aggregator.AggregatorTypeWheel)
		agg, ok := aggBaseType.(aggregator.TotalingAggregator)
		if !ok {
			panic(fmt.Errorf("a wheel aggregator should implement TotalingAggregator"))
		}

		// Start scrolling down
		agg.AggregatorSetPressAmt(-1, 42, aggregator.Press)
		// Stop scrolling down
		agg.AggregatorSetPressAmt(0, 42, aggregator.Release)

		if agg.CurPressTotal() >= 0 {
			t.Fatalf("a non-negative total is incorrect because we're scrolling down")
		}
	})
}
