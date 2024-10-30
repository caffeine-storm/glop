package ai_test

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestAiSpecs(t *testing.T) {
	convey.Convey("Specs for the AI package", t, func() {
		convey.Convey("XGML Load", XgmlLoadSpec)
		convey.Convey("Term", TermSpec)
		convey.Convey("Chunk", ChunkSpec)
	})
}
