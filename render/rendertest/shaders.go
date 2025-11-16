package rendertest

import (
	"fmt"

	"github.com/caffeine-storm/glop/render"
)

const shaderProgName = "debugshaders"

func WithShaderProgs(shaders *render.ShaderBank, vertShader string, fragShader string, fn func()) {
	if !shaders.HasShader(shaderProgName) {
		err := shaders.RegisterShader(shaderProgName, vertShader, fragShader)
		if err != nil {
			panic(fmt.Errorf("couldn't register debug shaders: %w", err))
		}
	}

	err := shaders.EnableShader(shaderProgName)
	if err != nil {
		panic(fmt.Errorf("couldn't enable debug shaders: %w", err))
	}

	defer func() {
		shaders.EnableShader("")
	}()
	fn()
}
