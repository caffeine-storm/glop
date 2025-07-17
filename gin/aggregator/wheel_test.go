package aggregator_test

import (
	"fmt"
	"testing"

	"github.com/runningwild/glop/gin/aggregator"
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

		doSynthEvent, synthAmount := agg.AggregatorThink(42)

		if doSynthEvent {
			panic(fmt.Errorf("the aggregator should not have thought to generate a synthetic event"))
		}
		if synthAmount != 0 {
			panic(fmt.Errorf("there should be no synthetic event press amount"))
		}

		if agg.FramePressTotal() == 0 {
			t.Fatalf("a zero sum is incorrect because we pressed it")
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

		doSynthEvent, synthAmount := agg.AggregatorThink(42)

		if doSynthEvent {
			panic(fmt.Errorf("the aggregator should not have thought to generate a synthetic event"))
		}
		if synthAmount != 0 {
			panic(fmt.Errorf("there should be no synthetic event press amount"))
		}

		if agg.FramePressTotal() >= 0 {
			t.Fatalf("a non-negative total is incorrect because we scrolled down")
		}
	})
}
