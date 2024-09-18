module github.com/caffeine-storm/glop-test

go 1.23.0

require (
	github.com/go-gl-legacy/gl v0.0.0-20150223033340-df25b1fe668d
	github.com/runningwild/glop v0.0.0-20150924024344-abed7bd11be4
)

require (
	code.google.com/p/freetype-go v0.0.0-00010101000000-000000000000 // indirect
	github.com/go-gl-legacy/glu v0.0.0-20150315173544-b54aa06bc77a // indirect
)

replace code.google.com/p/freetype-go => github.com/golang/freetype v0.0.0-20120725121025-28cc5fbc5d0b

replace github.com/runningwild/glop => ../../

replace github.com/go-gl-legacy/gl => github.com/caffeine-storm/gl v0.0.0-20240901153421-ffd1b6683995

replace github.com/go-gl-legacy/glu => github.com/caffeine-storm/glu v0.0.0-20240904141638-031792da4ab6
