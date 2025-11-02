package render

import (
	"fmt"

	"github.com/caffeine-storm/mathgl"
)

type Showmat mathgl.Mat4

func (m Showmat) String() string {
	elements := make([]interface{}, len(m))
	for i, elem := range m {
		// Shove each value into 9 characters and always put a decimal point.
		stringified := fmt.Sprintf("%#9.8g", elem)

		// Unfortunately, golang will add an _extra_ 4 characters if it needs to
		// tack on an exponent (e.g. e+09). Let's drop the 4 least-significant
		// digits as needed to work around it.
		if len(stringified) > 9 {
			stringified = stringified[0:5] + stringified[len(stringified)-4:]
		}
		elements[i] = stringified
	}

	lineFormat := "%s, %s, %s, %s"
	blockFormat := fmt.Sprintf("%s\n%s\n%s\n%s", lineFormat, lineFormat, lineFormat, lineFormat)

	return fmt.Sprintf(
		blockFormat,
		elements...,
	)
}
