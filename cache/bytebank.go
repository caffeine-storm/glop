package cache

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// TODO(tmckee): don't use two strings for the key!
type ByteBank interface {
	Read(p1, p2 string) ([]byte, bool, error)
	Write(p1, p2 string, data []byte) error
}

type FsByteBank struct{}

func (*FsByteBank) Read(p1, p2 string) ([]byte, bool, error) {
	// TODO(tmckee): DRY this out
	filename := filepath.Join(p1, p2)
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

func (*FsByteBank) Write(p1, p2 string, data []byte) error {
	filename := filepath.Join(p1, p2)

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("couldn't os.Create(%q): %v", filename, err)
	}
	defer f.Close()

	binary.Write(f, binary.LittleEndian, int32(len(data)))
	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("coudln't write to file %q: %v", filename, err)
	}

	return nil
}
