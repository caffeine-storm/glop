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

type ramByteBank struct{
	data map[string][]byte
}

func MakeRamByteBank() *ramByteBank {
	return &ramByteBank{
		data: map[string][]byte{},
	}
}

func (bank *ramByteBank) Read(key string) ([]byte, bool, error) {
	if ret, ok := bank.data[key]; ok {
		return ret, true, nil
	}

	return nil, false, nil
}

func (bank *ramByteBank) Write(key string, data []byte) error {
	bank.data[key] = data

	return nil
}
