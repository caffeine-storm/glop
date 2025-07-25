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
	})

	t.Run("standardAggregator", func(t *testing.T) {
		assert := assert.New(t)
		evt_type := aggregator.DecideEventType(0, 0, aggregator.AggregatorForType(aggregator.AggregatorTypeStandard))

		assert.Equal(aggregator.NoEvent, evt_type)
	})
}
