package strmanip

import (
	"fmt"
	"strings"
)

func Show[T any](data []T) string {
	if len(data) == 0 {
		return "[]"
	}
	bldr := strings.Builder{}
	bldr.WriteString("[")
	n := len(data)

	showDatum := func(a any) string {
		switch v := a.(type) {
		case string:
			return fmt.Sprintf(`%q`, v)
		default:
			return fmt.Sprintf(`%q`, fmt.Sprintf(`%v`, v))
		}
	}

	bldr.WriteString(showDatum(data[0]))
	for i := 1; i < n; i++ {
		bldr.WriteString(", ")
		bldr.WriteString(showDatum(data[i]))
	}

	bldr.WriteString("]")
	return bldr.String()
}
