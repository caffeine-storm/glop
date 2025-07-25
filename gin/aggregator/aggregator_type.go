package aggregator

import "fmt"

type AggregatorType int

const (
	AggregatorTypeStandard AggregatorType = iota
	AggregatorTypeAxis
	AggregatorTypeWheel
)

func AggregatorForType(tp AggregatorType) Aggregator {
	switch tp {
	case AggregatorTypeStandard:
		return &standardAggregator{}
	case AggregatorTypeAxis:
		return &axisAggregator{
			// 'Press' and 'Release' don't make sense for these types of keys so
			// set the original PressAmount to non-zero
			baseAggregator: baseAggregator{
				this: keyStats{
					press_amt: -1,
				},
			},
		}
	case AggregatorTypeWheel:
		return &wheelAggregator{}
	}

	panic(fmt.Errorf("unknown aggregatorType: %d", tp))
}

func (at AggregatorType) MustValidate() {
	switch at {
	case AggregatorTypeStandard:
	case AggregatorTypeAxis:
	case AggregatorTypeWheel:
	default:
		panic(fmt.Errorf("invalid aggregatorType: %d", at))
	}
}
