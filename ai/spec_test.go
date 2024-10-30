package ai_test

import (
	"github.com/orfjackal/gospec/src/gospec"
	"testing"
)

func TestAiSpecs(t *testing.T) {
	r := gospec.NewRunner()
	r.AddSpec(AiSpec)
	r.AddSpec(TermSpec)
	r.AddSpec(ChunkSpec)
	gospec.MainGoTest(r, t)
}
