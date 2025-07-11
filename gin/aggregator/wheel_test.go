package aggregator_test

import (
	"fmt"
	"testing"

	"github.com/runningwild/glop/gin/aggregator"
)

func TestWheelAggregator(t *testing.T) {
	t.Run("can handle scroll-start and scroll-end in single frame", func(t *testing.T) {
		agg := aggregator.AggregatorForType(aggregator.AggregatorTypeWheel)
		agg.AggregatorSetPressAmt(1, 42, aggregator.Press)
		agg.AggregatorSetPressAmt(0, 42, aggregator.Release)
		doSynthEvent, synthAmount := agg.AggregatorThink(42)

		if doSynthEvent {
			panic(fmt.Errorf("the aggregator should not have thought to generate a synthetic event"))
		}
		if synthAmount != 0 {
			panic(fmt.Errorf("there should be no synthetic event press amount"))
		}

		if agg.CurPressSum() == 0 {
			t.Fatalf("a zero sum is incorrect because we pressed it")
		}
	})
}
