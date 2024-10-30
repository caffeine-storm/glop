package sprite_test

import (
	"github.com/orfjackal/gospec/src/gospec"
	"testing"
)

func TestSpriteSpecs(t *testing.T) {
	r := gospec.NewRunner()
	r.AddSpec(LoadSpriteSpec)
	r.AddSpec(CommandNSpec)
	r.AddSpec(SyncSpec)
	r.AddSpec(ManagerSpec)
	gospec.MainGoTest(r, t)
}
