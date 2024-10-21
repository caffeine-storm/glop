package cache_test

import (
	"errors"
	"fmt"
	"io/fs"
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
	r.AddSpec(RamByteBankSpec)
	gospec.MainGoTest(r, t)
}

var (
	someData = []byte("lol")
	someOtherData = []byte("rofl")
)

func FsByteBankSpec(c gospec.Context) {
	c.Specify("An empty FsByteBank", func() {
		bank := &cache.FsByteBank{}

		c.Specify("can be constructed", func() {
			c.Expect(bank, gospec.Not(gospec.Equals), nil)
		})

		c.Specify("returns a 'miss' when reading a bogus key", func() {
			_, ok, err := bank.Read("not-present")
			c.Expect(err, gospec.IsNil)
			c.Expect(ok, gospec.Equals, false)
		})

		c.Specify("propagates file writing failures", func() {
			doesNotExistDir := "/does/not/exist/"
			_, err := os.Stat(doesNotExistDir)
			c.Assume(errors.Is(err, fs.ErrNotExist), gospec.IsTrue)

			err = bank.Write(path.Join(doesNotExistDir, "foo"), someData)
			c.Expect(err, gospec.Not(gospec.IsNil))
		})

		c.Specify("can write to a temp file", func() {
			f, err := os.CreateTemp("", "glop-test")
			if err != nil {
				panic(fmt.Errorf("couldn't create temp file: %w", err))
			}
			tmpfile := f.Name()
			defer os.Remove(tmpfile)

			err = bank.Write(tmpfile, someData)
			if err != nil {
				panic(fmt.Errorf("couldn't write data: %w", err))
			}

			c.Specify("the data can be read back", func() {
				data, ok, err := bank.Read(tmpfile)
				c.Expect(err, gospec.IsNil)
				c.Expect(ok, gospec.IsTrue)
				c.Expect(string(data), gospec.Equals, string(someData))
			})

			c.Specify("still misses for different key", func() {
				_, ok, err := bank.Read(tmpfile + "-but-miss")
				c.Expect(err, gospec.IsNil)
				c.Expect(ok, gospec.IsFalse)
			})
		})
	})

	c.Specify("An FsByteBank with some data", func() {
		bank := &cache.FsByteBank{}
		f, err := os.CreateTemp("", "glop-test")
		if err != nil {
			panic(fmt.Errorf("couldn't create temp file: %w", err))
		}
		tmpfile := f.Name()
		defer os.Remove(tmpfile)

		err = bank.Write(tmpfile, someData)
		c.Assume(err, gospec.IsNil)

		c.Specify("uses a flat format/encoding", func() {
			fileData, err := os.ReadFile(tmpfile)
			c.Assume(err, gospec.IsNil)

			c.Expect(fileData, gospec.ContainsInOrder, someData)
		})
	})
}

func RamByteBankSpec(c gospec.Context) {
	c.Specify("An empty RamByteBank", func() {
		bank := cache.MakeRamByteBank()

		c.Specify("can be constructed", func() {
			c.Expect(bank, gospec.Not(gospec.Equals), nil)
		})

		c.Specify("returns a 'miss' when reading a bogus key", func() {
			_, ok, err := bank.Read("not-present")
			c.Expect(err, gospec.IsNil)
			c.Expect(ok, gospec.Equals, false)
		})

		someKey := "some-key"
		someOtherKey := "some-other-key"

		c.Specify("can write a payload", func() {
			err := bank.Write(someKey, someData)
			if err != nil {
				panic(fmt.Errorf("couldn't write data: %w", err))
			}

			c.Specify("can read back the payload", func() {
				readData, hit, err := bank.Read(someKey)
				c.Expect(err, gospec.Equals, nil)
				c.Expect(hit, gospec.Equals, true)
				c.Expect(readData, gospec.ContainsInOrder, someData)
			})

			c.Specify("keeps payloads separate by key", func() {
				err := bank.Write(someOtherKey, someOtherData)
				c.Expect(err, gospec.IsNil)

				readData, hit, err := bank.Read(someKey)
				c.Expect(err, gospec.IsNil)
				c.Expect(hit, gospec.IsTrue)
				c.Expect(readData, gospec.ContainsInOrder, someData)
			})
		})
	})
}
