package systemtest

import (
	"fmt"
	"os/exec"
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

func xDoToolRun(xdotoolArgs ...any) {
	cmd := exec.Command("xdotool", stringifyArgs(xdotoolArgs)...)

	err := cmd.Run()
	if err != nil {
		panic(fmt.Errorf("could not %q: %w", cmd.String(), err))
	}
}

func xDoToolOutput(xdotoolArgs ...any) string {
	cmd := exec.Command("xdotool", stringifyArgs(xdotoolArgs)...)

	bs, err := cmd.Output()
	if err != nil {
		panic(fmt.Errorf("could not %q: %w", cmd.String(), err))
	}

	return string(bs)
}
