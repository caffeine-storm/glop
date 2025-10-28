package rendertest

import "github.com/runningwild/glop/system"

func AdvanceTime(sys system.System, delta uint64) {
	sys.(*system.MockSystem).AdvanceTime(delta)
}
