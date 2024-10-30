package gin_test

import (
	"github.com/orfjackal/gospec/src/gospec"
	"testing"
)

func TestGinSpecs(t *testing.T) {
	r := gospec.NewRunner()
	r.AddSpec(NaturalKeySpec)
	r.AddSpec(DerivedKeySpec)
	r.AddSpec(DeviceSpec)
	r.AddSpec(DeviceFamilySpec)
	r.AddSpec(NestedDerivedKeySpec)
	r.AddSpec(EventSpec)
	r.AddSpec(AxisSpec)
	r.AddSpec(EventListenerSpec)
	gospec.MainGoTest(r, t)
}
