package cache_test

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"

	// TODO(tmckee): use 'convey' instead:
	// https://github.com/smartystreets/goconvey
	"github.com/orfjackal/gospec/src/gospec"
	"github.com/runningwild/glop/cache"
)

func withScratchDir(op func(string)) {
	tmpdir, err := os.MkdirTemp("", "glop-test")
	if err != nil {
		panic(fmt.Errorf("couldn't MkdirTemp: %w", err))
	}
	defer os.RemoveAll(tmpdir)

	op(tmpdir)
}

func findFilesInDir(path string) map[string]bool {
	result := make(map[string]bool)

	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			result[path] = true
		}
		return err
	})

	if err != nil {
		panic(fmt.Errorf("filepath.Walk returned an error: %w", err))
	}

	return result
}

func setDifference(included, excluded map[string]bool) map[string]bool {
	result := make(map[string]bool)

	for keyIncluded := range included {
		// If the exlcusion set does not contain the key, it's in the difference.
		if !excluded[keyIncluded] {
			result[keyIncluded] = true
		}
	}

	return result
}

func TestCacheSpecs(t *testing.T) {
	withScratchDir(func(tmpdir string) {
		r := gospec.NewRunner()
		r.AddSpec(FsByteBankSpec)
		r.AddSpec(RamByteBankSpec)
		r.AddNamedSpec("FsByteBank is a ByteBank", ImplementsByteBank(cache.MakeFsByteBank(tmpdir)))
		r.AddNamedSpec("ramByteBank is a ByteBank", ImplementsByteBank(cache.MakeRamByteBank()))
		gospec.MainGoTest(r, t)
	})
}

var (
	someData      = []byte("lol")
	someOtherData = []byte("rofl")
)

func FsByteBankSpec(c gospec.Context) {
	withScratchDir(func(tmpdir string) {
		c.Specify("An empty FsByteBank", func() {
			bank := cache.MakeFsByteBank(tmpdir)

			c.Specify("propagates file writing failures", func() {
				doesNotExistDir := "/does/not/exist/"
				_, err := os.Stat(doesNotExistDir)
				c.Assume(errors.Is(err, fs.ErrNotExist), gospec.IsTrue)

				err = bank.Write(path.Join(doesNotExistDir, "foo"), someData)
				c.Expect(err, gospec.Not(gospec.IsNil))
			})

			c.Specify("can write to a temp file", func() {
				key := "fs-byte-bank"
				f, err := os.CreateTemp(tmpdir, key)
				if err != nil {
					panic(fmt.Errorf("couldn't create temp file: %w", err))
				}
				tmpfile := f.Name()
				defer os.Remove(tmpfile)

				err = bank.Write(key, someData)
				if err != nil {
					panic(fmt.Errorf("couldn't write data: %w", err))
				}

				c.Specify("the data can be read back", func() {
					data, ok, err := bank.Read(key)
					c.Expect(err, gospec.IsNil)
					c.Expect(ok, gospec.IsTrue)
					c.Expect(string(data), gospec.Equals, string(someData))
				})

				c.Specify("still misses for different key", func() {
					_, ok, err := bank.Read(key + "-but-miss")
					c.Expect(err, gospec.IsNil)
					c.Expect(ok, gospec.IsFalse)
				})
			})

			c.Specify("relative paths refer to the bound directory", func() {
				filesInTempDirBefore := findFilesInDir(tmpdir)

				err := bank.Write("someKey", someData)
				c.Assume(err, gospec.IsNil)

				filesInTempDirAfter := findFilesInDir(tmpdir)

				delta := setDifference(filesInTempDirAfter, filesInTempDirBefore)

				c.Expect(len(delta) != 0, gospec.IsTrue)
			})
		})

		c.Specify("An FsByteBank with some data", func() {
			bank := cache.MakeFsByteBank(tmpdir)
			keyBase := "fs-byte-bank"
			f, err := os.CreateTemp(tmpdir, keyBase)
			if err != nil {
				panic(fmt.Errorf("couldn't create temp file: %w", err))
			}
			tmpfile := f.Name()
			defer os.Remove(tmpfile)

			// CreateTemp helpfully appends some digits to 'keyBase'; we can only
			// know what the file is actually called after it's created.
			_, key := path.Split(f.Name())

			err = bank.Write(key, someData)
			c.Assume(err, gospec.IsNil)

			c.Specify("uses a flat format/encoding", func() {
				fileData, err := os.ReadFile(tmpfile)
				c.Assume(err, gospec.IsNil)

				c.Expect(fileData, gospec.ContainsInOrder, someData)
			})
		})

		c.Specify("FsByteBank construction needs a directory", func() {
			c.Specify("a temp directory is okay", func() {
				bank := cache.MakeFsByteBank(tmpdir)
				c.Expect(bank, gospec.Not(gospec.IsNil))
			})
			c.Specify("a missing directory is not okay", func() {
				bank := cache.MakeFsByteBank("/does/not/exist")
				c.Expect(bank, gospec.IsNil)
			})
			c.Specify("a regular file can't be used as a directory", func() {
				regfile, err := os.Create(path.Join(tmpdir, "foo.txt"))
				c.Assume(err, gospec.IsNil)

				bank := cache.MakeFsByteBank(regfile.Name())
				c.Expect(bank, gospec.IsNil)
			})
		})
	})
}

func RamByteBankSpec(c gospec.Context) {
	c.Specify("An empty RamByteBank", func() {
		bank := cache.MakeRamByteBank()

		c.Specify("can be constructed", func() {
			c.Expect(bank, gospec.Not(gospec.Equals), nil)
		})
	})
}

func ImplementsByteBank(bb cache.ByteBank) func(gospec.Context) {
	return func(c gospec.Context) {
		c.Specify("An empty ByteBank", func() {
			c.Specify("can be constructed", func() {
				c.Expect(bb, gospec.Not(gospec.Equals), nil)
			})

			c.Specify("returns a 'miss' when reading a bogus key", func() {
				_, ok, err := bb.Read("not-present")
				c.Expect(err, gospec.IsNil)
				c.Expect(ok, gospec.Equals, false)
			})

			someKey := "some-key"
			someOtherKey := "some-other-key"

			c.Specify("can write a payload", func() {
				err := bb.Write(someKey, someData)
				if err != nil {
					panic(fmt.Errorf("couldn't write data: %w", err))
				}

				c.Specify("can read back the payload", func() {
					readData, hit, err := bb.Read(someKey)
					c.Expect(err, gospec.Equals, nil)
					c.Expect(hit, gospec.Equals, true)
					c.Expect(readData, gospec.ContainsInOrder, someData)
				})

				c.Specify("keeps payloads separate by key", func() {
					err := bb.Write(someOtherKey, someOtherData)
					c.Expect(err, gospec.IsNil)

					readData, hit, err := bb.Read(someKey)
					c.Expect(err, gospec.IsNil)
					c.Expect(hit, gospec.IsTrue)
					c.Expect(readData, gospec.ContainsInOrder, someData)
				})
			})
		})
	}
}
