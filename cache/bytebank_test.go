package cache_test

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/runningwild/glop/cache"
	. "github.com/smartystreets/goconvey/convey"
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
		Convey("FsByteBankSpec", t, FsByteBankSpec)
		Convey("RamByteBankSpec", t, RamByteBankSpec)

		Convey("FsByteBank is a ByteBank", t, ImplementsByteBank(cache.MakeFsByteBank(tmpdir)))
		Convey("RamByteBank is a ByteBank", t, ImplementsByteBank(cache.MakeRamByteBank()))
		Convey("LockingByteBank is a ByteBank", t, ImplementsByteBank(cache.MakeLockingByteBank(cache.MakeRamByteBank())))
	})
}

var (
	someData      = []byte("lol")
	someOtherData = []byte("rofl")
)

func FsByteBankSpec() {
	withScratchDir(func(tmpdir string) {
		Convey("An empty FsByteBank", func() {
			bank := cache.MakeFsByteBank(tmpdir)

			Convey("propagates file writing failures", func() {
				doesNotExistDir := "/does/not/exist/"
				_, err := os.Stat(doesNotExistDir)
				So(errors.Is(err, fs.ErrNotExist), ShouldEqual, true)

				err = bank.Write(path.Join(doesNotExistDir, "foo"), someData)
				So(err, ShouldNotBeNil)
			})

			Convey("can write to a temp file", func() {
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

				Convey("the data can be read back", func() {
					data, ok, err := bank.Read(key)
					So(err, ShouldBeNil)
					So(ok, ShouldBeTrue)
					So(string(data), ShouldEqual, string(someData))
				})

				Convey("still misses for different key", func() {
					_, ok, err := bank.Read(key + "-but-miss")
					So(err, ShouldBeNil)
					So(ok, ShouldBeFalse)
				})
			})

			Convey("relative paths refer to the bound directory", func() {
				filesInTempDirBefore := findFilesInDir(tmpdir)

				err := bank.Write("someKey", someData)
				So(err, ShouldBeNil)

				filesInTempDirAfter := findFilesInDir(tmpdir)

				delta := setDifference(filesInTempDirAfter, filesInTempDirBefore)

				So(len(delta), ShouldNotEqual, 0)
			})
		})

		Convey("An FsByteBank with some data", func() {
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
			So(err, ShouldBeNil)

			Convey("uses a flat format/encoding", func() {
				fileData, err := os.ReadFile(tmpfile)
				So(err, ShouldBeNil)

				So(fileData, ShouldResemble, someData)
			})
		})

		Convey("FsByteBank construction needs a directory", func() {
			Convey("a temp directory is okay", func() {
				bank := cache.MakeFsByteBank(tmpdir)
				So(bank, ShouldNotBeNil)
			})
			Convey("a missing directory is not okay", func() {
				bank := cache.MakeFsByteBank("/does/not/exist")
				So(bank, ShouldBeNil)
			})
			Convey("a regular file can't be used as a directory", func() {
				regfile, err := os.Create(path.Join(tmpdir, "foo.txt"))
				So(err, ShouldBeNil)

				bank := cache.MakeFsByteBank(regfile.Name())
				So(bank, ShouldBeNil)
			})
		})
	})
}

func RamByteBankSpec() {
	Convey("An empty RamByteBank", func() {
		bank := cache.MakeRamByteBank()

		Convey("can be constructed", func() {
			So(bank, ShouldNotBeNil)
		})
	})
}

func ImplementsByteBank(bb cache.ByteBank) func() {
	return func() {
		Convey("An empty ByteBank", func() {
			Convey("can be constructed", func() {
				So(bb, ShouldNotBeNil)
			})

			Convey("returns a 'miss' when reading a bogus key", func() {
				_, ok, err := bb.Read("not-present")
				So(err, ShouldBeNil)
				So(ok, ShouldBeFalse)
			})

			someKey := "some-key"
			someOtherKey := "some-other-key"

			Convey("can write a payload", func() {
				err := bb.Write(someKey, someData)
				if err != nil {
					panic(fmt.Errorf("couldn't write data: %w", err))
				}

				Convey("can read back the payload", func() {
					readData, hit, err := bb.Read(someKey)
					So(err, ShouldBeNil)
					So(hit, ShouldBeTrue)
					So(readData, ShouldResemble, someData)
				})

				Convey("keeps payloads separate by key", func() {
					err := bb.Write(someOtherKey, someOtherData)
					So(err, ShouldBeNil)

					readData, hit, err := bb.Read(someKey)
					So(err, ShouldBeNil)
					So(hit, ShouldBeTrue)
					So(readData, ShouldResemble, someData)
				})
			})
		})
	}
}
