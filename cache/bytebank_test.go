package cache_test

import (
	"testing"

	// TODO(tmckee): use 'convey' instead:
	// https://github.com/smartystreets/goconvey
	"github.com/orfjackal/gospec/src/gospec"
	"github.com/runningwild/glop/cache"
)

func TestCacheSpecs(t *testing.T) {
	r := gospec.NewRunner()
	r.AddSpec(FsByteBankSpec)
	gospec.MainGoTest(r, t)
}

func FsByteBankSpec(c gospec.Context) {
	c.Specify("An FsByteBank can be constructed", func() {
		bank := &cache.FsByteBank{}
		c.Expect(bank, gospec.Not(gospec.Equals), nil)
	})
}
