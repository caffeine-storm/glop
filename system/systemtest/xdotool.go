package systemtest

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime/trace"
)

func stringifyArgs(args []any) []string {
	ret := make([]string, len(args))
	for i, arg := range args {
		switch v := arg.(type) {
		case string:
			ret[i] = v
		case fmt.Stringer:
			ret[i] = v.String()
		default:
			ret[i] = fmt.Sprintf("%v", arg)
		}
	}
	return ret
}

func getEnv(key string) string {
	ret, found := os.LookupEnv(key)
	if !found {
		panic(fmt.Errorf("couldn't find %q environment variable", key))
	}
	return ret
}

func traceRunningExternalCommand(cmd string, stringargs []string) (bs []byte, err error) {
	ctx := context.Background()
	trace.WithRegion(ctx, "systemtest-xdotool", func() {
		trace.Logf(ctx, "systemtest-xdotool", "cmd: %q, args: %s", cmd, stringargs)
		cmd := exec.Command(cmd, stringargs...)
		bs, err = cmd.Output()
		if err != nil {
			err = fmt.Errorf("couldn't run %q: %w", cmd.String(), err)
		}
	})
	return
}

func xDoToolRun(xdotoolArgs ...any) {
	_ = xDoToolOutput(xdotoolArgs...)
}

func xDoToolOutput(xdotoolArgs ...any) string {
	display := getEnv("DISPLAY")
	auth := getEnv("XAUTHORITY")

	podmanArgs := []string{
		"podman", "run",

		// Cleanup the container after use
		"--rm",

		// Use the same network namespace as the host so that we can connect to the X
		// server's unix-domain socket. Note that we don't have to bind a filesystem
		// path because we can rely on the 'abstract' socket namespace on Linux.
		"--net=host",

		// Bind-mount the Xauthority file and pass in $XAUTHORITY environment
		// variable so that in-container commands can authenticate to the X server.
		"--volume", fmt.Sprintf("%s:/.Xauthority", auth),
		"--env", "XAUTHORITY=/.Xauthority",

		// Forward $DISPLAY to the container.
		"--env", fmt.Sprintf("DISPLAY=%s", display),

		// Use xdotool version 4.
		"caffeinestorm/xdotool-4:20251204.1",

		// The container has an entrypoint of /bin/bash ... let's run xdotool
		// directly.
		"xdotool",
	}

	// Pass along arguments intended for xdotool.
	podmanArgs = append(podmanArgs, stringifyArgs(xdotoolArgs)...)

	bs, err := traceRunningExternalCommand(podmanArgs[0], podmanArgs[1:])
	if err != nil {
		panic(err)
	}

	return string(bs)
}
