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
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("couldn't os.Create(%q): %v", filename, err)
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("coudln't write payload to file %q: %v", filename, err)
	}

	return nil
}
