package aggregator_test

import (
	"testing"

	"github.com/runningwild/glop/gin/aggregator"
	"github.com/stretchr/testify/assert"
)

func TestDecideEventType(t *testing.T) {
	t.Run("axis aggregator", func(t *testing.T) {
		assert := assert.New(t)
		axisAgg := aggregator.AggregatorForType(aggregator.AggregatorTypeAxis)
		evt_type := axisAgg.DecideEventType(0, 0)

		assert.Equal(aggregator.Adjust, evt_type)

		t.Run("won't emit release/press when moving to position 0", func(t *testing.T) {
			agg := aggregator.AggregatorForType(aggregator.AggregatorTypeAxis)

			evt_type := agg.DecideEventType(0, 42)
			assert.Equal(aggregator.Adjust, evt_type)

			evt_type = agg.DecideEventType(42, 0)
			assert.Equal(aggregator.Adjust, evt_type)
		})
	})

	t.Run("standardAggregator", func(t *testing.T) {
		assert := assert.New(t)
		stdAgg := aggregator.AggregatorForType(aggregator.AggregatorTypeStandard)
		evt_type := stdAgg.DecideEventType(0, 0)

		assert.Equal(aggregator.NoEvent, evt_type)
	})
}
