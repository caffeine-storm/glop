module github.com/runningwild/glop

go 1.14

replace code.google.com/p/freetype-go => github.com/golang/freetype v0.0.0-20120725121025-28cc5fbc5d0b

replace github.com/go-gl-legacy/gl => github.com/caffeine-storm/gl v0.0.0-20240901153421-ffd1b6683995

replace github.com/go-gl-legacy/glu => github.com/caffeine-storm/glu v0.0.0-20240904141638-031792da4ab6

require (
	code.google.com/p/freetype-go v0.0.0-00010101000000-000000000000
	github.com/go-gl-legacy/gl v0.0.0-20150223033340-df25b1fe668d
	github.com/go-gl-legacy/glu v0.0.0-20150315173544-b54aa06bc77a
	github.com/orfjackal/gospec v0.0.0-20140731185859-a21081619255
	github.com/orfjackal/nanospec.go v0.0.0-20120727230329-de4694c1d701 // indirect
	github.com/runningwild/polish v0.0.0-20120524023733-9d9a1194cd81
	github.com/runningwild/yedparse v0.0.0-20120306014153-f7df1db2f9d9
	github.com/smartystreets/goconvey v1.8.1
	github.com/stretchr/testify v1.9.0
)
