package strmanip

import (
	"fmt"
	"strings"
)

func Show(data []string) string {
	bldr := strings.Builder{}
	bldr.WriteString("[")
	n := len(data)

	switch n {
	case 0:
		break
	case 1:
		bldr.WriteString(fmt.Sprintf("%q", data[0]))
	default:
		for i := 0; i < n-1; i++ {
			bldr.WriteString(fmt.Sprintf("%q, ", data[i]))
		}
		bldr.WriteString(fmt.Sprintf("%q", data[n-1]))
	}

	bldr.WriteString("]")
	return bldr.String()
}
