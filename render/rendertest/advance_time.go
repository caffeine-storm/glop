package rendertest

import "github.com/caffeine-storm/glop/system"

func AdvanceTimeMillis(sys system.System, delta uint64) {
	sys.(*system.MockSystem).AdvanceTimeMillis(delta)
}
