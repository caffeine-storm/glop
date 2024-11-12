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
