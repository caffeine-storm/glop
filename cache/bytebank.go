package cache

import (
	"encoding/binary"
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
	// TODO(tmckee): DRY this out
	f, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// If the file doesn't exist, we just haven't cached at this path yet.
			return nil, false, nil
		}
		// Other errors indicate something fatal.
		return nil, false, fmt.Errorf("couldn't open file %q: %v", filename, err)
	}
	defer f.Close()

	var length int32
	err = binary.Read(f, binary.LittleEndian, &length)
	if err != nil {
		return nil, false, fmt.Errorf("couldn't read length prefix: %v", err)
	}

	buf := make([]byte, length)
	_, err = f.Read(buf)
	if err != nil {
		return nil, true, fmt.Errorf("couldn't read payload: %v", err)
	}

	return buf, true, nil
}

func (*FsByteBank) Write(filename string, data []byte) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("couldn't os.Create(%q): %v", filename, err)
	}
	defer f.Close()

	// TODO(tmckee): we don't need to write a size; we can os.Stat the file for
	// its size instead.
	err = binary.Write(f, binary.LittleEndian, int32(len(data)))
	if err != nil {
		return fmt.Errorf("coudln't write length header to file %q: %v", filename, err)
	}
	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("coudln't write payload to file %q: %v", filename, err)
	}

	return nil
}
