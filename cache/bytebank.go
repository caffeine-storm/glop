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

type FsByteBank struct{}

func (*FsByteBank) Read(filename string) ([]byte, bool, error) {
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

func (*FsByteBank) Write(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}

type RamByteBank struct{
	key string
	data []byte
}

func (bank *RamByteBank) Read(key string) ([]byte, bool, error) {
	if bank.key != key {
		return nil, false, nil
	}

	return bank.data, true, nil
}

func (bank *RamByteBank) Write(key string, data []byte) error {
	bank.key = key
	bank.data = data
	return nil
}
