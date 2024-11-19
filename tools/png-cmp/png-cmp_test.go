package main_test

import "testing"
import "github.com/runningwild/glop/tools/png-cmp"

func TestMyMax(t *testing.T) {
	if main.MyMax(0, 0, 0) != 0 {
		t.Fail()
	}

	if main.MyMax(0, 1, 2) != 2 {
		t.Fail()
	}
}
