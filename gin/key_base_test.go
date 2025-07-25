package gin_test

import (
	"fmt"
	"testing"

	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/gin/aggregator"
	"github.com/stretchr/testify/assert"
)

// Stub implementation of aggregator.SubAggregator
type stubAggregator struct{}

func (*stubAggregator) IsDown() bool             { return false }
func (*stubAggregator) FramePressCount() int     { return 0 }
func (*stubAggregator) FrameReleaseCount() int   { return 0 }
func (*stubAggregator) FramePressAmt() float64   { return 0 }
func (*stubAggregator) FramePressSum() float64   { return 0 }
func (*stubAggregator) FramePressAvg() float64   { return 0 }
func (*stubAggregator) FramePressTotal() float64 { return 0 }
func (*stubAggregator) CurPressCount() int       { return 0 }
func (*stubAggregator) CurReleaseCount() int     { return 0 }
func (*stubAggregator) CurPressAmt() float64     { return 0 }
func (*stubAggregator) CurPressSum() float64     { return 0 }

// Stub implementation of gin.Key
type stubKey struct {
	stubAggregator
}

func (sk *stubKey) String() string {
	return fmt.Sprintf("stubKey %p", sk)
}

func (sk *stubKey) Name() string {
	return sk.String()
}

func (*stubKey) Id() gin.KeyId {
	return gin.KeyId{}
}

func (*stubKey) KeySetPressAmt(amt float64, ms int64, cause gin.Event) gin.Event {
	return gin.Event{}
}

func (*stubKey) KeyThink(ms int64) (bool, float64) {
	return false, 0
}

func TestKeyInterface(t *testing.T) {
	t.Run("Key doesn't include Cursor", func(t *testing.T) {
		var _ gin.Key = (*stubKey)(nil)
	})
}

func TestDecideEventType(t *testing.T) {
	t.Run("axis aggregator", func(t *testing.T) {
		assert := assert.New(t)

		evt_type := gin.DecideEventType(0, 0, aggregator.AggregatorForType(aggregator.AggregatorTypeAxis))

		assert.Equal(aggregator.Adjust, evt_type)
	})

	t.Run("standardAggregator", func(t *testing.T) {
		assert := assert.New(t)
		// TODO: move DecideEventType to the aggregator package.
		evt_type := gin.DecideEventType(0, 0, aggregator.AggregatorForType(aggregator.AggregatorTypeStandard))

		assert.Equal(aggregator.NoEvent, evt_type)
	})
}
