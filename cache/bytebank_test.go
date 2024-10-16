package cache_test

import (
	"fmt"
	"os"
	"path"
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

var someData []byte = []byte("lol")

func FsByteBankSpec(c gospec.Context) {
	c.Specify("An empty FsByteBank", func() {
		bank := &cache.FsByteBank{}

		c.Specify("can be constructed", func() {
			c.Expect(bank, gospec.Not(gospec.Equals), nil)
		})

		c.Specify("returns a 'miss' when reading", func() {
			_, ok, err := bank.Read("p1", "p2")
			c.Expect(err, gospec.IsNil)
			c.Expect(ok, gospec.Equals, false)
		})

		c.Specify("propagates file writing failures", func() {
			// TODO(tmckee): find a portable way to pick an unwriteable file.  This
			// is linux only for now.
			err := bank.Write("/dev", "full", someData)
			c.Expect(err, gospec.Not(gospec.IsNil))
		})

		c.Specify("can write to a temp file", func() {
			f, err := os.CreateTemp("", "glop-test")
			if err != nil {
				panic(fmt.Errorf("couldn't create temp file: %w", err))
			}
			tmpfile := f.Name()
			defer os.Remove(tmpfile)

			tmpdir, tmpname := path.Split(tmpfile)

			err = bank.Write(tmpdir, tmpname, someData)
			if err != nil {
				panic(fmt.Errorf("couldn't write data: %w", err))
			}

			c.Specify("the data can be read back", func() {
				data, ok, err := bank.Read(tmpdir, tmpname)
				c.Expect(err, gospec.IsNil)
				c.Expect(ok, gospec.IsTrue)
				c.Expect(string(data), gospec.Equals, string(someData))
			})

			c.Specify("still misses for different key", func() {
				_, ok, err := bank.Read(tmpdir, tmpname+"-but-miss")
				c.Expect(err, gospec.IsNil)
				c.Expect(ok, gospec.IsFalse)
			})
		})
	})
}
