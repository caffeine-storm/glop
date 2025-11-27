package systemtest

import "testing"

func TestXdoTool(t *testing.T) {
	result := xDoToolOutput("getmouselocation")
	t.Log("initial pos", result)

	if len(result) == 0 {
		t.Fatalf("no output from xdotool :(")
	}
}
