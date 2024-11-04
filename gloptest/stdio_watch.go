package gloptest

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
)

// Runs the given operation and returns a slice of strings that the operation
// wrote to log.Default().*, slog.Default().*, stdout and stderr combined.
func CollectOutput(operation func()) []string {
	read, write, err := os.Pipe()
	if err != nil {
		panic(fmt.Errorf("couldn't os.Pipe: %w", err))
	}

	go func() {
		stdlogger := log.Default()
		oldLogOut := stdlogger.Writer()
		stdlogger.SetOutput(write)
		defer stdlogger.SetOutput(oldLogOut)

		stdSlogger := slog.Default()
		pipeSlogger := slog.New(slog.NewTextHandler(write, nil))
		slog.SetDefault(pipeSlogger)
		defer slog.SetDefault(stdSlogger)

		oldStdout := os.Stdout
		os.Stdout = write
		defer func() { os.Stdout = oldStdout }()

		oldStderr := os.Stderr
		os.Stderr = write
		defer func() { os.Stderr = oldStderr }()

		// Prefer to defer closing the write end of the pipe. If operation panics,
		// the pipe still needs to be closed or else the reading goroutine would
		// block forever.
		defer write.Close()

		operation()
	}()

	byteList, err := io.ReadAll(read)
	if err != nil {
		panic(fmt.Errorf("couldn't io.ReadAll on the read end of the pipe: %w", err))
	}

	if len(byteList) == 0 {
		return []string{}
	}

	return strings.Split(string(byteList), "\n")
}
