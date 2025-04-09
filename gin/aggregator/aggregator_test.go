package aggregator_test

import (
	"testing"

	"github.com/runningwild/glop/gin/aggregator"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	time1 int64 = iota*5 + 10
	time2
	time3
	time4
)

func TestAggregators(t *testing.T) {
	Convey("aggregator api", t, func() {
		Convey("standard aggregator", func() {
			agg := aggregator.AggregatorForType(aggregator.AggregatorTypeStandard)

			agg.AggregatorSetPressAmt(3, time1, aggregator.Press)
			agg.AggregatorSetPressAmt(7, time2, aggregator.Press)
			agg.AggregatorThink(time3)

			Convey("FramePressAmt should represent the instantaneous PressAmt at the end of last frame", func() {
				So(agg.FramePressAmt(), ShouldEqual, 7)
			})

			Convey("the last frame's PressAmt is the initial PressAmt for the current frame", func() {
				So(agg.CurPressAmt(), ShouldEqual, 7)
			})

			agg.AggregatorSetPressAmt(13, time4, aggregator.Press)

			Convey("the current PressAmt is updated during AggregatorSetPressAmt", func() {
				So(agg.CurPressAmt(), ShouldEqual, 13)
			})
			Convey("the frame PressAmt is _not_ updated during AggregatorSetPressAmt", func() {
				So(agg.FramePressAmt(), ShouldEqual, 7)
			})
		})

		Convey("wheel aggregator", func() {
			agg := aggregator.AggregatorForType(aggregator.AggregatorTypeWheel)

			agg.AggregatorSetPressAmt(3, time1, aggregator.Press)
			agg.AggregatorSetPressAmt(7, time2, aggregator.Press)
			agg.AggregatorThink(time3)

			Convey("FramePressAmt should represent the instantaneous PressAmt at the end of last frame", func() {
				So(agg.FramePressAmt(), ShouldEqual, 7)
			})

			Convey("the last frame's PressAmt is the initial PressAmt for the current frame", func() {
				So(agg.CurPressAmt(), ShouldEqual, 7)
			})

			agg.AggregatorSetPressAmt(13, time4, aggregator.Press)

			Convey("the current PressAmt is updated during AggregatorSetPressAmt", func() {
				So(agg.CurPressAmt(), ShouldEqual, 13)
			})
			Convey("the frame PressAmt is _not_ updated during AggregatorSetPressAmt", func() {
				So(agg.FramePressAmt(), ShouldEqual, 7)
			})
		})

		Convey("axis aggregator", func() {
			agg := aggregator.AggregatorForType(aggregator.AggregatorTypeAxis)

			agg.AggregatorSetPressAmt(3, time1, aggregator.Press)
			agg.AggregatorSetPressAmt(7, time2, aggregator.Press)
			agg.AggregatorThink(time3)

			Convey("FramePressAmt should represent the instantaneous PressAmt at the end of last frame", func() {
				So(agg.FramePressAmt(), ShouldEqual, 7)
			})

			Convey("the last frame's PressAmt is the initial PressAmt for the current frame", func() {
				So(agg.CurPressAmt(), ShouldEqual, 7)
			})

			agg.AggregatorSetPressAmt(13, time4, aggregator.Press)

			Convey("the current PressAmt is updated during AggregatorSetPressAmt", func() {
				So(agg.CurPressAmt(), ShouldEqual, 13)
			})
			Convey("the frame PressAmt is _not_ updated during AggregatorSetPressAmt", func() {
				So(agg.FramePressAmt(), ShouldEqual, 7)
			})
		})
	})
}
