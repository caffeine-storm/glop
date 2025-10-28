package rendertest

import "github.com/runningwild/glop/system"

func AdvanceTimeMillis(sys system.System, delta uint64) {
	sys.(*system.MockSystem).AdvanceTimeMillis(delta)
}
