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
	c.Specify("An empty FsByteBank", func() {
		bank := &cache.FsByteBank{}

		c.Specify("can be constructed", func() {
			c.Expect(bank, gospec.Not(gospec.Equals), nil)
		})

		c.Specify("returns a 'miss' when reading", func() {
			data, ok, err := bank.Read("p1", "p2")
			c.Expect(err, gospec.IsNil)
			c.Expect(data == nil, gospec.IsTrue)
			c.Expect(ok, gospec.Equals, false)
		})
	})
}
