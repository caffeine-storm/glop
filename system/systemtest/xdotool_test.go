package systemtest

import (
	"strconv"
	"strings"
	"testing"

	"github.com/caffeine-storm/glop/gui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXdoToolVersion(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	// There was a bug (https://github.com/jordansissel/xdotool/issues/463) in
	// xdotool that we need to avoid. We check here to make sure xdotool is at
	// version 4 or greater.
	result := xDoToolOutput("--version") // Should output something like `xdotool version 3.20160805.1`
	parts := strings.Split(result, " ")
	result = parts[len(parts)-1]
	parts = strings.Split(result, ".")
	require.Len(parts, 3, "couldn't parse result (%q) as 'major.minor.patch'", result)

	asInt, err := strconv.Atoi(parts[0])
	require.NoError(err, "couldn't parse 'major' version component, %q, as an int: %s", parts[0], err)

	assert.GreaterOrEqual(asInt, 4)
}

func BenchmarkMouseMoves(b *testing.B) {
	// Before the change to run version 4+ through podman:
	// BenchmarkMouseMoves-12    10    14,754,686,470 ns/op
	// After
	// BenchmarkMouseMoves-12    10       704,084,872 ns/op
	pt := gui.PointAt(42, 13)

	WithTestWindowDriver(64, 64, func(driver Driver) {
		for b.Loop() {
			driver.MoveMouse(pt.X, pt.Y)
			driver.ProcessFrame()
		}
	})
}
