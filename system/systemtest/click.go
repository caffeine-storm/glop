package systemtest

import (
	"fmt"
	"os/exec"
)

func RunXDoTool(xdotoolArgs ...string) {
	cmd := exec.Command("xdotool", xdotoolArgs...)

	err := cmd.Run()
	if err != nil {
		panic(fmt.Errorf("could not %q: %w", cmd.String(), err))
	}
}
