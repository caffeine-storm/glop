package systemtest

import (
	"context"
	"fmt"
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

func traceRunningExternalCommand(xdotoolArgs []any) (bs []byte, err error) {
	ctx := context.Background()
	trace.WithRegion(ctx, "systemtest-xdotool", func() {
		stringargs := stringifyArgs(xdotoolArgs)
		trace.Logf(ctx, "systemtest-xdotool", "args: %s", stringargs)
		cmd := exec.Command("xdotool", stringargs...)
		bs, err = cmd.Output()
		if err != nil {
			err = fmt.Errorf("couldn't run %q: %w", cmd.String(), err)
		}
	})
	return
}

func xDoToolRun(xdotoolArgs ...any) {
	_, err := traceRunningExternalCommand(xdotoolArgs)
	if err != nil {
		panic(err)
	}
}

func xDoToolOutput(xdotoolArgs ...any) string {
	cmd := exec.Command("xdotool", stringifyArgs(xdotoolArgs)...)

	bs, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	return string(bs)
}
