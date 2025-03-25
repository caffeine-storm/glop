package systemtest

import (
	"fmt"
	"os/exec"
)

func runXDoTool(xdotoolAnyArgs ...any) {
	xdotoolArgs := make([]string, len(xdotoolAnyArgs))
	for i, arg := range xdotoolAnyArgs {
		switch v := arg.(type) {
		case string:
			xdotoolArgs[i] = v
		case fmt.Stringer:
			xdotoolArgs[i] = v.String()
		default:
			xdotoolArgs[i] = fmt.Sprintf("%v", arg)
		}
	}

	cmd := exec.Command("xdotool", xdotoolArgs...)

	err := cmd.Run()
	if err != nil {
		panic(fmt.Errorf("could not %q: %w", cmd.String(), err))
	}
}
