package cache

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type ByteBank interface {
	Read(key string) ([]byte, bool, error)
	Write(key string, data []byte) error
}

type fsByteBank struct{}

func MakeFsByteBank() *fsByteBank {
	return &fsByteBank{}
}

func (*fsByteBank) Read(filename string) ([]byte, bool, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// If the file doesn't exist, we just haven't cached at this path yet.
			return nil, false, nil
		}
		// Other errors indicate something fatal.
		return nil, false, fmt.Errorf("couldn't open file %q: %v", filename, err)
	}
	return data, true, nil
}

func (*fsByteBank) Write(filename string, data []byte) error {
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
