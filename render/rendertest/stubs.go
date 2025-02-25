package rendertest

import (
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/render"
)

func StubShaderBank(shaderNames ...string) *render.ShaderBank {
	ret := &render.ShaderBank{
		ShaderProgs: map[string]gl.Program{},
	}

	noShader := gl.Program(0)
	for _, name := range shaderNames {
		ret.ShaderProgs[name] = noShader
	}

	return ret
}

type stubbedQueue struct{}

var _ render.RenderQueueInterface = ((*stubbedQueue)(nil))

func (*stubbedQueue) Queue(_ render.RenderJob) {}
func (*stubbedQueue) Purge()                   {}
func (*stubbedQueue) StartProcessing()         {}
func (*stubbedQueue) IsPurging() bool {
	return false
}

func MakeStubbedRenderQueue() render.RenderQueueInterface {
	return &stubbedQueue{}
}
