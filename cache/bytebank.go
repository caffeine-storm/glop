package cache

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
)

type ByteBank interface {
	Read(key string) ([]byte, bool, error)
	Write(key string, data []byte) error
}

type fsByteBank struct {
	root string
}

func isDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func MakeFsByteBank(path string) *fsByteBank {
	if !isDir(path) {
		slog.Error("need path to existing directory", "path", path)
		return nil
	}

	return &fsByteBank{
		root: path,
	}
}

func (bank *fsByteBank) Read(key string) ([]byte, bool, error) {
	filename := path.Join(bank.root, key)
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// If the file doesn't exist, we just haven't cached at this path yet.
			return nil, false, nil
		}
		// Other errors indicate something fatal.
		return nil, false, fmt.Errorf("couldn't open file %q: %w", filename, err)
	}
	return data, true, nil
}

func (bank *fsByteBank) Write(key string, data []byte) error {
	filename := path.Join(bank.root, key)
	return os.WriteFile(filename, data, 0644)
}

type ramByteBank map[string][]byte

func MakeRamByteBank() ramByteBank {
	return ramByteBank{}
}

func (bank ramByteBank) Read(key string) ([]byte, bool, error) {
	if ret, ok := bank[key]; ok {
		return ret, true, nil
	}

	return nil, false, nil
}

func (bank ramByteBank) Write(key string, data []byte) error {
	bank[key] = data

	return nil
}
