package aggregator_test

import (
	"testing"

	"github.com/runningwild/glop/gin/aggregator"
	"github.com/stretchr/testify/assert"
)

func TestDecideEventType(t *testing.T) {
	t.Run("axis aggregator", func(t *testing.T) {
		assert := assert.New(t)
		evt_type := aggregator.DecideEventType(0, 0, aggregator.AggregatorForType(aggregator.AggregatorTypeAxis))

		assert.Equal(aggregator.Adjust, evt_type)

		t.Run("won't emit release/press when moving to position 0", func(t *testing.T) {
			agg := aggregator.AggregatorForType(aggregator.AggregatorTypeAxis)

			evt_type := aggregator.DecideEventType(0, 42, agg)
			assert.Equal(aggregator.Adjust, evt_type)

			evt_type = aggregator.DecideEventType(42, 0, agg)
			assert.Equal(aggregator.Adjust, evt_type)
		})
	})

	t.Run("standardAggregator", func(t *testing.T) {
		assert := assert.New(t)
		evt_type := aggregator.DecideEventType(0, 0, aggregator.AggregatorForType(aggregator.AggregatorTypeStandard))

		assert.Equal(aggregator.NoEvent, evt_type)
	})
}
