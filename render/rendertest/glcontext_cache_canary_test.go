package rendertest_test

import "fmt"

func thisFunctionDereferencesNil() {
	var nilPointer *string = nil

	_ = len(*nilPointer)

	panic(fmt.Errorf("should not get here"))
}
